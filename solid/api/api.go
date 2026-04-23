package api

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/shareed2k/goth_fiber"
	v1 "github.com/zeddo123/mlsolid/solid/api/v1"
	"github.com/zeddo123/mlsolid/solid/controllers"
	"github.com/zeddo123/mlsolid/solid/oauth"
)

// NewAPI inits a new api router.
func NewAPI(ctrl *controllers.Controller, cfg oauth.Config) *fiber.App {
	app := fiber.New(fiber.Config{}) //nolint: exhaustruct

	// Create session store
	config := session.Config{ //nolint: exhaustruct
		KeyLookup:      "cookie:mlsolid",
		Expiration:     24 * time.Hour, //nolint: mnd
		CookieSecure:   cfg.Prod,
		CookieHTTPOnly: true,
	}

	session := session.New(config)

	oauth.NewAuth(cfg)

	app.Use(logger.New())
	app.Use(cors.New())
	app.Use(csrf.New())

	// Inject controller into fiber's context
	app.Use(func(ctx *fiber.Ctx) error {
		ctx.Locals("ctrl", ctrl)

		return ctx.Next()
	})

	err := v1.BuildRoutes(app, authedUser(session))
	if err != nil {
		panic(err)
	}

	app.Get("/login/:provider", goth_fiber.BeginAuthHandler)

	app.Get("/authorize", authedUser(session), func(c *fiber.Ctx) error {
		key, err := ctrl.GenerateKey(c.Context())
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "could not generate api key",
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"api-key": key,
		})
	})

	app.Get("/auth/callback/:provider", func(c *fiber.Ctx) error {
		user, err := goth_fiber.CompleteUserAuth(c)
		if err != nil {
			log.Println(err)

			return c.Status(fiber.StatusUnauthorized).SendString("could not authenticate user")
		}

		if !oauth.IsAllowed(user, cfg) {
			return c.Status(fiber.StatusUnauthorized).SendString("unauthorized to access resource")
		}

		ses, err := session.Get(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("session error")
		}

		ses.Set("authenticated", true)
		ses.Set("user_id", user.UserID)
		ses.Set("email", user.Email)
		ses.Set("access_token", user.AccessTokenSecret)
		ses.Set("refresh_token", user.RefreshToken)

		if err := ses.Save(); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("failed to save user session")
		}

		log.Println(ses.Keys())

		return c.SendString(user.Email)
	})

	app.Get("/logout", func(c *fiber.Ctx) error {
		err := goth_fiber.Logout(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("could not logout user")
		}

		ses, err := session.Get(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("session error")
		}

		err = ses.Destroy()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("could not clear session")
		}

		return c.Status(fiber.StatusOK).SendString("logged out")
	})

	app.Get("/authorized", authedUser(session), func(c *fiber.Ctx) error {
		authenticated, ok := c.Locals("authenticated").(bool)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).SendString("internal error")
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"authenticated": authenticated,
		})
	})

	return app
}

func StartServer(port string, ctrl *controllers.Controller, config oauth.Config) {
	app := NewAPI(ctrl, config)

	if err := app.Listen(":" + port); err != nil {
		panic(err)
	}
}
