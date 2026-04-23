package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

func authedUser(session *session.Store) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ses, err := session.Get(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("session error")
		}

		if ses.Get("authenticated") != true {
			return c.Status(fiber.StatusUnauthorized).SendString("unauthenticated")
		}

		c.Locals("authenticated", ses.Get("authenticated"))
		c.Locals("user_id", ses.Get("user_id"))
		c.Locals("email", ses.Get("email"))
		c.Locals("access_token", ses.Get("access_token"))
		c.Locals("refresh_token", ses.Get("refresh_token"))

		return c.Next()
	}
}
