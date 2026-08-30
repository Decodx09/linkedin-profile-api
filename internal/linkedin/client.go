package linkedin

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	voyagerBase = "https://www.linkedin.com/voyager/api"
	authBase    = "https://www.linkedin.com"
	userAgent   = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 " +
		"(KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"
)

type ErrKind int

const (
	ErrUpstream ErrKind = iota
	ErrAuth
	ErrNotFound
	ErrRateLimited
	ErrBadInput
)

type Error struct {
	Kind    ErrKind
	Message string
}

func (e *Error) Error() string { return e.Message }

func newErr(kind ErrKind, format string, a ...any) *Error {
	return &Error{Kind: kind, Message: fmt.Sprintf(format, a...)}
}

var publicIDRe = regexp.MustCompile(`^[\w\-.%]+$`)
var inPathRe = regexp.MustCompile(`/in/([^/?#]+)`)

func ExtractPublicID(input string) (string, error) {
	value := strings.TrimSpace(input)
	if value == "" {
		return "", newErr(ErrBadInput, "empty LinkedIn profile URL / id")
	}

	var publicID string
	if strings.Contains(value, "linkedin.com") || strings.HasPrefix(value, "http") {
		raw := value
		if !strings.Contains(raw, "://") {
			raw = "https://" + raw
		}
		u, err := url.Parse(raw)
		if err != nil {
			return "", newErr(ErrBadInput, "could not parse URL: %v", err)
		}
		m := inPathRe.FindStringSubmatch(u.Path)
		if len(m) < 2 {
			return "", newErr(ErrBadInput, "could not find '/in/<id>' in URL: %s", input)
		}
		publicID = m[1]
	} else {
		publicID = value
	}

	publicID = strings.Trim(publicID, "/")
	if !publicIDRe.MatchString(publicID) {
		return "", newErr(ErrBadInput, "invalid public identifier parsed: %q", publicID)
	}
	return publicID, nil
}

type Client struct {
	http      *http.Client
	csrfToken string
}

type Option struct {
	LiAt       string
	JSessionID string
	Email      string
	Password   string
}

func New(opt Option) (*Client, error) {
	jar, _ := cookiejar.New(nil)
	c := &Client{
		http: &http.Client{
			Jar:     jar,
			Timeout: 25 * time.Second,
		},
	}

	switch {
	case opt.LiAt != "":
		if err := c.authWithCookie(opt.LiAt, opt.JSessionID); err != nil {
			return nil, err
		}
	case opt.Email != "" && opt.Password != "":
		if err := c.authWithCredentials(opt.Email, opt.Password); err != nil {
			return nil, err
		}
	default:
		return nil, newErr(ErrAuth,
			"no credentials supplied: provide li_at (recommended) or email+password")
	}
	return c, nil
}

func linkedinURL() *url.URL {
	u, _ := url.Parse("https://www.linkedin.com")
	return u
}

func (c *Client) setCookie(name, value string) {
	c.http.Jar.SetCookies(linkedinURL(), []*http.Cookie{{Name: name, Value: value}})
}

func (c *Client) getCookie(name string) string {
	for _, ck := range c.http.Jar.Cookies(linkedinURL()) {
		if ck.Name == name {
			return ck.Value
		}
	}
	return ""
}

func (c *Client) authWithCookie(liAt, jsessionid string) error {
	c.setCookie("li_at", liAt)

	if jsessionid == "" {
		req, _ := http.NewRequest(http.MethodGet, authBase+"/feed/", nil)
		req.Header.Set("User-Agent", userAgent)
		if resp, err := c.http.Do(req); err == nil {
			_ = resp.Body.Close()
		}
		jsessionid = c.getCookie("JSESSIONID")
	}
	if jsessionid == "" {
		return newErr(ErrAuth,
			"could not obtain a JSESSIONID/CSRF token; the li_at cookie may be expired")
	}

	token := strings.Trim(jsessionid, `"`)
	c.setCookie("JSESSIONID", `"`+token+`"`)
	c.csrfToken = token
	return nil
}

func (c *Client) authWithCredentials(email, password string) error {
	req, _ := http.NewRequest(http.MethodGet, authBase+"/uas/login", nil)
	req.Header.Set("User-Agent", userAgent)
	resp, err := c.http.Do(req)
	if err != nil {
		return newErr(ErrUpstream, "failed to load login page: %v", err)
	}
	_ = resp.Body.Close()

	jsessionid := c.getCookie("JSESSIONID")
	if jsessionid == "" {
		return newErr(ErrAuth, "login page did not set a JSESSIONID cookie")
	}
	csrfParam := strings.Trim(jsessionid, `"`)

	form := url.Values{}
	form.Set("session_key", email)
	form.Set("session_password", password)
	form.Set("loginCsrfParam", csrfParam)

	req2, _ := http.NewRequest(http.MethodPost, authBase+"/uas/login-submit",
		strings.NewReader(form.Encode()))
	req2.Header.Set("User-Agent", userAgent)
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp2, err := c.http.Do(req2)
	if err != nil {
		return newErr(ErrUpstream, "login submit failed: %v", err)
	}
	defer resp2.Body.Close()

	if c.getCookie("li_at") == "" {
		finalURL := resp2.Request.URL.String()
		if strings.Contains(finalURL, "challenge") || strings.Contains(finalURL, "checkpoint") {
			return newErr(ErrAuth,
				"LinkedIn issued a security checkpoint/2FA during password login; "+
					"use the LINKEDIN_LI_AT cookie method instead")
		}
		return newErr(ErrAuth, "login failed — no li_at cookie was issued; check credentials")
	}

	token := strings.Trim(c.getCookie("JSESSIONID"), `"`)
	c.csrfToken = token
	return nil
}

func (c *Client) get(path string, query url.Values) (map[string]any, error) {
	full := voyagerBase + path
	if len(query) > 0 {
		full += "?" + query.Encode()
	}

	req, _ := http.NewRequest(http.MethodGet, full, nil)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("x-restli-protocol-version", "2.0.0")
	req.Header.Set("x-li-lang", "en_US")
	req.Header.Set("csrf-token", c.csrfToken)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, newErr(ErrUpstream, "request failed: %v", err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusUnauthorized:
		return nil, newErr(ErrAuth, "401 Unauthorized — the LinkedIn session is invalid or expired")
	case resp.StatusCode == http.StatusForbidden:
		return nil, newErr(ErrRateLimited,
			"403 Forbidden — LinkedIn blocked the request (bot challenge or profile out of reach)")
	case resp.StatusCode == http.StatusNotFound:
		return nil, newErr(ErrNotFound, "404 — profile not found or not public")
	case resp.StatusCode == http.StatusTooManyRequests:
		return nil, newErr(ErrRateLimited, "429 Too Many Requests — you are being throttled")
	case resp.StatusCode >= 400:
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 300))
		return nil, newErr(ErrUpstream, "unexpected HTTP %d from LinkedIn: %s",
			resp.StatusCode, string(body))
	}

	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, newErr(ErrUpstream, "LinkedIn returned non-JSON (often an auth wall): %v", err)
	}
	return out, nil
}

func (c *Client) GetProfileView(publicID string) (map[string]any, error) {
	return c.get("/identity/profiles/"+publicID+"/profileView", nil)
}

func (c *Client) GetContactInfo(publicID string) map[string]any {
	out, err := c.get("/identity/profiles/"+publicID+"/profileContactInfo", nil)
	if err != nil {
		return nil
	}
	return out
}

func (c *Client) GetSkills(publicID string) map[string]any {
	q := url.Values{}
	q.Set("count", "100")
	q.Set("start", "0")
	out, err := c.get("/identity/profiles/"+publicID+"/skills", q)
	if err != nil {
		return nil
	}
	return out
}
