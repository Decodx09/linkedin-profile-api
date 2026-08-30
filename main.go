package main

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/Decodx09/linkedin-profile-api/internal/config"
	"github.com/Decodx09/linkedin-profile-api/internal/linkedin"
	"github.com/Decodx09/linkedin-profile-api/internal/models"
)

const version = "1.0.0"

type server struct {
	cfg *config.Config

	clientOnce sync.Once
	client     *linkedin.Client
	clientErr  error

	cacheMu sync.Mutex
	cache   map[string]cacheEntry
}

type cacheEntry struct {
	at   time.Time
	resp models.ProfileResponse
}

func main() {
	cfg := config.Load()
	s := &server{cfg: cfg, cache: make(map[string]cacheEntry)}

	app := fiber.New(fiber.Config{
		AppName:      "LinkedIn Profile API",
		ErrorHandler: errorHandler,
	})
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New())

	app.Get("/", s.root)
	app.Get("/health", s.health)

	api := app.Group("/api", s.apiKeyGuard)
	api.Get("/profile", s.getProfile)
	api.Post("/profile", s.postProfile)

	addr := ":" + cfg.Port
	log.Printf("LinkedIn Profile API v%s listening on %s (auth configured: %v)",
		version, addr, cfg.HasCookieAuth() || cfg.HasCredentialAuth())
	if err := app.Listen(addr); err != nil {
		log.Fatal(err)
	}
}

func (s *server) getClient() (*linkedin.Client, error) {
	s.clientOnce.Do(func() {
		s.client, s.clientErr = linkedin.New(linkedin.Option{
			LiAt:       s.cfg.LiAt,
			JSessionID: s.cfg.JSessionID,
			Email:      s.cfg.Email,
			Password:   s.cfg.Password,
		})
	})
	return s.client, s.clientErr
}

func (s *server) apiKeyGuard(c *fiber.Ctx) error {
	if s.cfg.APIKey == "" {
		return c.Next()
	}
	supplied := c.Get("X-API-Key")
	if supplied == "" {
		supplied = c.Query("api_key")
	}
	if supplied != s.cfg.APIKey {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid or missing API key")
	}
	return c.Next()
}

func (s *server) root(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"service": "LinkedIn Profile API",
		"version": version,
		"endpoints": fiber.Map{
			"profile": "GET /api/profile?url=<linkedin_profile_url>",
			"health":  "/health",
		},
		"auth_configured": s.cfg.HasCookieAuth() || s.cfg.HasCredentialAuth(),
	})
}

func (s *server) health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *server) getProfile(c *fiber.Ctx) error {
	url := c.Query("url")
	if url == "" {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "missing 'url' query parameter")
	}
	return s.fetchAndRespond(c, url)
}

func (s *server) postProfile(c *fiber.Ctx) error {
	var body struct {
		URL string `json:"url"`
	}
	if err := c.BodyParser(&body); err != nil || strings.TrimSpace(body.URL) == "" {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "body must include a 'url' field")
	}
	return s.fetchAndRespond(c, body.URL)
}

func (s *server) fetchAndRespond(c *fiber.Ctx, url string) error {
	publicID, err := linkedin.ExtractPublicID(url)
	if err != nil {
		return mapError(err)
	}

	if cached, ok := s.cacheGet(publicID); ok {
		return c.JSON(cached)
	}

	client, err := s.getClient()
	if err != nil {
		return mapError(err)
	}

	profileView, err := client.GetProfileView(publicID)
	if err != nil {
		return mapError(err)
	}
	contact := client.GetContactInfo(publicID)
	skills := client.GetSkills(publicID)

	profile := linkedin.ParseProfile(publicID, profileView, contact, skills)
	resp := models.ProfileResponse{
		OK:        true,
		Source:    "linkedin-voyager",
		Cached:    false,
		FetchedAt: time.Now().UTC().Format(time.RFC3339),
		Profile:   profile,
	}
	s.cacheSet(publicID, resp)
	return c.JSON(resp)
}

func (s *server) cacheGet(key string) (models.ProfileResponse, bool) {
	if s.cfg.CacheTTLSeconds <= 0 {
		return models.ProfileResponse{}, false
	}
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	entry, ok := s.cache[key]
	if !ok {
		return models.ProfileResponse{}, false
	}
	if time.Since(entry.at) > time.Duration(s.cfg.CacheTTLSeconds)*time.Second {
		delete(s.cache, key)
		return models.ProfileResponse{}, false
	}
	resp := entry.resp
	resp.Cached = true
	return resp, true
}

func (s *server) cacheSet(key string, resp models.ProfileResponse) {
	if s.cfg.CacheTTLSeconds <= 0 {
		return
	}
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	s.cache[key] = cacheEntry{at: time.Now(), resp: resp}
}

func mapError(err error) error {
	if le, ok := err.(*linkedin.Error); ok {
		switch le.Kind {
		case linkedin.ErrBadInput:
			return fiber.NewError(fiber.StatusUnprocessableEntity, le.Message)
		case linkedin.ErrAuth:
			return fiber.NewError(fiber.StatusBadGateway, "LinkedIn auth error: "+le.Message)
		case linkedin.ErrNotFound:
			return fiber.NewError(fiber.StatusNotFound, le.Message)
		case linkedin.ErrRateLimited:
			return fiber.NewError(fiber.StatusTooManyRequests, le.Message)
		default:
			return fiber.NewError(fiber.StatusBadGateway, "LinkedIn error: "+le.Message)
		}
	}
	return fiber.NewError(fiber.StatusInternalServerError, err.Error())
}

func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if fe, ok := err.(*fiber.Error); ok {
		code = fe.Code
	}
	return c.Status(code).JSON(fiber.Map{
		"ok":     false,
		"error":  statusName(code),
		"detail": err.Error(),
	})
}

func statusName(code int) string {
	switch code {
	case fiber.StatusUnauthorized:
		return "unauthorized"
	case fiber.StatusNotFound:
		return "not_found"
	case fiber.StatusUnprocessableEntity:
		return "invalid_input"
	case fiber.StatusTooManyRequests:
		return "rate_limited"
	case fiber.StatusBadGateway:
		return "upstream_error"
	default:
		return "internal_error"
	}
}
