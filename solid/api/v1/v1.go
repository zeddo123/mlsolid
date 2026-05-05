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

// BuildRoutes builds v1 endpoint routes.
func BuildRoutes(f *fiber.App, middlewares ...fiber.Handler) error {
	v1 := f.Group("/v1", middlewares...)

	v1.Get("/exps", experiments)
	v1.Get("/exp/:id", experiment)

	v1.Get("/exp/:id/metrics", metrics)
	v1.Get("/exp/:id/metric/:mid", metric)

	v1.Get("/exp/:id/artifacts", artifacts)
	v1.Get("artifact/:rid/:aid", artifact)

	v1.Get("/registries", registries)
	v1.Get("/registry/:id", registry)

	v1.Get("/benchmarks", benchmarks)
	v1.Get("/benchmark/:id", benchmark)
	v1.Post("/benchmark", createBenchmark)
	v1.Patch("/benchmark/:id/toggle", toogleBenchmark)
	v1.Patch("/benchmark/:id", updateBenchmark)
	v1.Delete("/benchmark/:id", deleteBenchmark)
	v1.Get("/benchmark/:id/runs", benchmarkRuns)
	v1.Get("/benchmark/:id/best", benchmarkBest)

	return nil
}
