package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	v1 "github.com/zeddo123/mlsolid/solid/api/v1"
	"github.com/zeddo123/mlsolid/solid/controllers"
)

func NewAPI(ctrl *controllers.Controller) *fiber.App {
	app := fiber.New(fiber.Config{})

	app.Use(logger.New())
	app.Use(cors.New())

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

func StartServer(port string, ctrl *controllers.Controller) {
	app := NewAPI(ctrl)

	if err := app.Listen(":" + port); err != nil {
		panic(err)
	}
}
