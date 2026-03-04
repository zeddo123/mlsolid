package v1

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zeddo123/mlsolid/solid/controllers"
)

func ctxController(ctx *fiber.Ctx) *controllers.Controller {
	if ctrl := ctx.Locals("ctrl"); ctrl != nil {
		if c, ok := ctrl.(*controllers.Controller); ok {
			return c
		}
	}

	panic("no controller found in ctx")
}

func BuildRoutes(f *fiber.App) error {
	v1 := f.Group("/v1")

	v1.Get("/exps", experiments)
	v1.Get("/exp/:id", experiment)

	v1.Get("/exp/:id/metrics", metrics)
	v1.Get("/exp/:id/metric/:mid", metric)

	v1.Get("/exp/:id/artifacts", artifacts)
	v1.Get("artifact/:rid/:aid", artifact)

	return nil
}
