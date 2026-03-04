package api //nolint: revive

import (
	"github.com/gofiber/fiber/v2"

	v1 "github.com/zeddo123/mlsolid/solid/api/v1"
	"github.com/zeddo123/mlsolid/solid/controllers"
)

func NewAPI(ctrl *controllers.Controller) *fiber.App {
	app := fiber.New(fiber.Config{})

	// Inject controller into fiber's context
	app.Use(func(ctx *fiber.Ctx) error {
		ctx.Locals("ctrl", ctrl)

		return ctx.Next()
	})

	err := v1.BuildRoutes(app)
	if err != nil {
		panic(err)
	}

	return app
}
