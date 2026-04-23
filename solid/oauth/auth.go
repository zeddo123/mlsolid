// Package oauth inits oauth providers using goth.
package oauth

import (
	"errors"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
	"github.com/shareed2k/goth_fiber"
)

// Config configuration struct for oauth.
type Config struct {
	Prod                 bool
	RootURL              string
	GoogleClientID       string
	GoogleClientSecret   string
	GoogleAllowedDomains []string
}

// NewAuth inits oauth providers.
func NewAuth(config Config) {
	sesCfg := session.Config{ //nolint: exhaustruct
		KeyLookup:      "cookie:oauth-mlsolid",
		Expiration:     24 * time.Minute, //nolint: mnd
		CookieSecure:   config.Prod,
		CookieHTTPOnly: true,
		CookieSameSite: "Lax",
	}

	goth_fiber.SessionStore = session.New(sesCfg)

	googleCallback, err := url.JoinPath(config.RootURL, "/auth/callback/google")
	if err != nil {
		panic(err)
	}

	goth.UseProviders(
		google.New(config.GoogleClientID, config.GoogleClientSecret, googleCallback),
	)
}

// IsAllowed checks if authentication is allowed depending on auth providers.
func IsAllowed(user goth.User, config Config) bool {
	switch user.Provider {
	case "google":
		val, ok := user.RawData["hd"]
		if !ok {
			domain, err := getEmailDomain(user.Email)
			if err != nil {
				return false
			}

			val = domain
		}

		domain, ok := val.(string)
		if !ok {
			return false
		}

		return slices.Contains(config.GoogleAllowedDomains, domain)
	default:
		return false
	}
}

func getEmailDomain(email string) (string, error) {
	addr := strings.Split(email, "@")

	if len(addr) == 1 {
		return "", errors.New("invalid email address")
	}

	return addr[1], nil
}
