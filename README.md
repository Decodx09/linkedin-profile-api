# LinkedIn Profile API (Go + Fiber)

A hosted HTTPS API that accepts a **LinkedIn profile URL** and returns most of
the information on that profile page as **structured JSON** — name, headline,
location, about, experience, education, skills, certifications, languages,
profile/background images, and (when visible) contact info.

It works by reverse-engineering LinkedIn's own internal **Voyager API** — the
private JSON API the linkedin.com web app itself calls — and driving it with a
server-side authenticated session. Built in **Go** with the
[**Fiber**](https://gofiber.io) web framework.

```
GET https://<your-host>/api/profile?url=https://www.linkedin.com/in/williamhgates/
```

> 📄 **New here or want the non-technical overview?** See **[NOTES.md](NOTES.md)** for a
> plain-English summary of what this does, what currently works, what doesn't yet,
> and why.

---

## Table of contents
- [How it works (approach)](#how-it-works-approach)
- [Response schema](#response-schema)
- [Setup](#setup)
- [Getting your `li_at` cookie](#getting-your-li_at-cookie)
- [Run locally](#run-locally)
- [Deploy publicly over HTTPS](#deploy-publicly-over-https)
- [API documentation](#api-documentation)
- [Known limitations](#known-limitations)
- [Legal & ethics](#legal--ethics)
- [Project layout](#project-layout)

---

## How it works (approach)

LinkedIn's profile pages are rendered client-side. The HTML from `GET /in/<id>`
contains almost no profile data; the browser fetches it afterwards from a private
API:

```
https://www.linkedin.com/voyager/api/identity/profiles/{public_id}/profileView
```

Every authenticated Voyager request needs three things, which this project
reproduces exactly (see [`internal/linkedin/client.go`](internal/linkedin/client.go)):

| Requirement | Where it comes from |
|---|---|
| `li_at` **session cookie** | Your logged-in LinkedIn session (the real auth token) |
| `csrf-token` **header** | Must equal the value of the `JSESSIONID` cookie |
| `x-restli-protocol-version: 2.0.0` header + a browser User-Agent | Static headers the web app always sends |

**Pipeline:**

1. **Authenticate** — reuse a `li_at` cookie you provide (recommended), or perform
   the `/uas/login-submit` username+password handshake as a fallback. A Go
   `cookiejar` holds the session; the CSRF token is mirrored into the header.
2. **Normalize input** — extract the public identifier from any profile URL
   (`/in/<id>` … `?query`), or accept a bare id.
3. **Fetch** three decorations:
   - `…/profileView` → bio, positions, education, certifications, languages,
     volunteering, publications, honors, and skills.
   - `…/profileContactInfo` → emails, phone numbers, websites, Twitter.
   - `…/skills` → the fuller, paginated skills list.
4. **Transform** the deeply-nested Voyager JSON into a clean, stable schema
   (see [`internal/linkedin/parser.go`](internal/linkedin/parser.go)). All access
   is defensive, so a renamed/missing field degrades to empty rather than
   crashing. Image URLs are reconstructed from LinkedIn's `VectorImage`
   (rootUrl + largest artifact).
5. **Serve** it via Fiber over HTTPS, with optional API-key auth, a small
   in-process TTL cache, panic recovery, and structured JSON errors.

A full example payload lives in
[`examples/sample_response.json`](examples/sample_response.json).

---

## Response schema

```jsonc
{
  "ok": true,
  "source": "linkedin-voyager",
  "cached": false,
  "fetched_at": "2026-08-30T15:04:13Z",
  "profile": {
    "public_id": "williamhgates",
    "profile_url": "https://www.linkedin.com/in/williamhgates/",
    "first_name": "Bill",
    "last_name": "Gates",
    "full_name": "Bill Gates",
    "headline": "Co-chair, Bill & Melinda Gates Foundation",
    "summary": "Chair of the Gates Foundation. Founder of Breakthrough Energy...",
    "location": "Seattle, Washington, United States",
    "country": "United States",
    "industry": "Philanthropy",
    "profile_picture": "https://media.licdn.com/dms/image/.../400_400/...jpg",
    "background_image": "https://media.licdn.com/dms/image/...",
    "experience": [
      {
        "title": "Co-chair",
        "company": "Bill & Melinda Gates Foundation",
        "company_linkedin_url": "https://www.linkedin.com/company/1446919/",
        "employment_type": "",
        "location": "Seattle, WA",
        "description": "...",
        "date_range": { "start": "2000-01", "end": "" },
        "company_logo": "https://media.licdn.com/dms/image/..."
      }
    ],
    "education":      [ { "school": "...", "degree": "...", "field_of_study": "...", "date_range": {"start": "...", "end": "..."} } ],
    "skills":         [ "Leadership", "Philanthropy" ],
    "certifications": [ { "name": "...", "authority": "...", "license_number": "...", "url": "...", "date_range": {} } ],
    "languages":      [ { "name": "English", "proficiency": "NATIVE_OR_BILINGUAL" } ],
    "volunteer":      [ { "role": "...", "organization": "...", "cause": "...", "date_range": {} } ],
    "publications":   [ { "name": "...", "publisher": "...", "url": "...", "date": "2021" } ],
    "honors":         [ { "title": "...", "issuer": "...", "date": "2019" } ],
    "contact_info":   { "emails": [], "phone_numbers": [], "websites": [], "twitter": [] }
  }
}
```

`date_range.end == ""` means "present". Any field that isn't available on the
profile (or isn't visible to your account) is returned empty (`""` / `[]`) rather
than omitted, so the shape is stable. Full field definitions live in
[`internal/models/models.go`](internal/models/models.go).

---

## Setup

**Requirements:** Go 1.23+.

```bash
git clone https://github.com/Decodx09/linkedin-profile-api.git
cd linkedin-profile-api
go mod download
cp .env.example .env      # then edit .env
```

Fill in `.env`:

```ini
LINKEDIN_LI_AT=<your li_at cookie>     # recommended
# or, as a fallback:
# LINKEDIN_EMAIL=you@example.com
# LINKEDIN_PASSWORD=...
API_KEY=<optional shared secret to protect your endpoint>
CACHE_TTL_SECONDS=3600
PORT=8000
```

> **Secrets never go in the repo.** `.env`, `*.pem`, and cookie files are
> git-ignored. On a host, set these as dashboard environment variables.

### Getting your `li_at` cookie

Using the cookie (instead of email/password) avoids storing your password and is
far less likely to trip a security checkpoint.

1. Log in to LinkedIn in your browser.
2. Open **DevTools → Application → Cookies → `https://www.linkedin.com`**.
3. Copy the **Value** of the `li_at` cookie into `LINKEDIN_LI_AT`.
4. *(Optional)* copy `JSESSIONID` into `LINKEDIN_JSESSIONID` — otherwise the app
   fetches a fresh CSRF token on first use.

---

## Run locally

```bash
go run .
# or build a binary:
go build -o server . && ./server
```

Then:

```bash
# health check
curl http://localhost:8000/health

# fetch a profile (GET)
curl "http://localhost:8000/api/profile?url=https://www.linkedin.com/in/williamhgates/"

# fetch a profile (POST)
curl -X POST http://localhost:8000/api/profile \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://www.linkedin.com/in/williamhgates/"}'

# with API key protection enabled
curl -H "X-API-Key: $API_KEY" \
  "http://localhost:8000/api/profile?url=https://www.linkedin.com/in/williamhgates/"
```

Run the tests:

```bash
go test ./...
```

### With Docker

```bash
docker build -t linkedin-profile-api .
docker run --rm -p 8000:8000 --env-file .env linkedin-profile-api
```

---

## Deploy publicly over HTTPS

Any container/PaaS host works. The repo ships a multi-stage `Dockerfile`, a
`Procfile`, and a Render blueprint. All hosts below terminate TLS for you, so the
API is served over **HTTPS** automatically.

**Render (recommended, one-click):**
1. Push this repo to GitHub.
2. In Render: **New → Blueprint**, point it at the repo (it reads `render.yaml`).
3. Add the secret env vars `LINKEDIN_LI_AT` (and optionally `API_KEY`) in the
   dashboard — they are marked `sync: false` so they are never committed.
4. Deploy → you get `https://<name>.onrender.com`.

**Railway / Fly.io / Google Cloud Run / any Docker host:**
- Build the `Dockerfile`, set the same env vars, expose `$PORT`. Example (Cloud Run):
  ```bash
  gcloud run deploy linkedin-profile-api --source . \
    --set-env-vars LINKEDIN_LI_AT=xxxx,API_KEY=yyyy --allow-unauthenticated
  ```

---

## API documentation

| Method | Path | Description |
|---|---|---|
| `GET`  | `/` | Service metadata |
| `GET`  | `/health` | Liveness probe |
| `GET`  | `/api/profile?url=<url>` | Fetch a profile as JSON |
| `POST` | `/api/profile` | Same, with `{"url": "..."}` body |

**Auth (optional):** if `API_KEY` is set, send it as the `X-API-Key` header or
`?api_key=` query param. If unset, the endpoint is open.

**Input accepted for `url`:** a full profile URL
(`https://www.linkedin.com/in/<id>/`), a bare `linkedin.com/in/<id>?...`, or the
raw public id.

**Status codes:**

| Code | `error` | Meaning |
|---|---|---|
| `200` | — | Success |
| `401` | `unauthorized` | Missing/invalid `X-API-Key` |
| `404` | `not_found` | Profile not found or not visible to your account |
| `422` | `invalid_input` | Malformed URL / missing `url` |
| `429` | `rate_limited` | LinkedIn is throttling the session |
| `502` | `upstream_error` | LinkedIn auth error or upstream failure |

Errors share one envelope: `{"ok": false, "error": "<slug>", "detail": "<msg>"}`.

---

## Known limitations

- **Requires a valid LinkedIn session.** LinkedIn has no public/official profile
  API; this uses a private endpoint via your own cookie. If the cookie expires
  you must refresh `LINKEDIN_LI_AT`.
- **Visibility follows your account.** You can only see what *your* logged-in
  account can see. Out-of-network profiles may be partial; fields like exact
  connections/follower counts and contact info are often withheld and returned
  empty.
- **Rate limits / anti-bot.** LinkedIn actively throttles automation. High volume
  can trigger `429`/`403` or a security checkpoint. Use conservatively, keep the
  cache on, and add your own rate limiting for production. A residential/clean IP
  is more reliable than a datacenter IP.
- **Password login is fragile.** The email/password fallback frequently hits a
  2FA/checkpoint. The `li_at` cookie method is strongly preferred.
- **Schema drift.** Voyager is a private API; LinkedIn can rename fields at any
  time. Parsing is defensive (missing fields → empty), but a large change may
  need a parser update in [`internal/linkedin/parser.go`](internal/linkedin/parser.go).
- **Not affiliated with LinkedIn.** For authorized/personal use; see below.

---

## Legal & ethics

This project is intended for **authorized, personal, and educational use** with
**your own** LinkedIn account and data you are permitted to access. Automated
access to LinkedIn may conflict with LinkedIn's Terms of Service. You are
responsible for how you use it, for respecting rate limits and privacy, and for
complying with applicable laws (e.g. GDPR/CCPA). Do not use it to collect data at
scale or for purposes people haven't consented to.

---

## Project layout

```
.
├── main.go                        # Fiber app: routes, middleware, cache, errors
├── internal/
│   ├── config/config.go           # env-based settings
│   ├── linkedin/
│   │   ├── client.go              # Voyager API client + auth (reverse-engineered core)
│   │   ├── parser.go              # raw Voyager JSON -> clean schema
│   │   └── parser_test.go         # URL parsing + parser unit tests
│   └── models/models.go           # response structs (the schema)
├── examples/
│   └── sample_response.json
├── Dockerfile                     # multi-stage build -> tiny alpine image
├── Procfile
├── render.yaml                    # one-click Render deploy blueprint
├── go.mod / go.sum
├── .env.example
└── README.md
```
