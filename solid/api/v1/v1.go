package v1

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/zeddo123/mlsolid/solid/controllers"
)

const (
	defaultPageLimit = 50
	maxPageLimit     = 200
)

func ctxController(ctx *fiber.Ctx) *controllers.Controller {
	if ctrl := ctx.Locals("ctrl"); ctrl != nil {
		if c, ok := ctrl.(*controllers.Controller); ok {
			return c
		}
	}

	panic("no controller found in ctx")
}

// parsePagination parses the cursor and limit query params shared by paginated list endpoints.
// cursor defaults to 0 (start) and limit defaults to defaultPageLimit, capped at maxPageLimit.
func parsePagination(c *fiber.Ctx) (cursor uint64, limit int64, err error) {
	cursor, err = strconv.ParseUint(c.Query("cursor", "0"), 10, 64)
	if err != nil {
		return 0, 0, errors.New("cursor query param is malformed")
	}

	limit = int64(c.QueryInt("limit", defaultPageLimit))
	if limit <= 0 {
		return 0, 0, errors.New("limit query param must be positive")
	}

	if limit > maxPageLimit {
		limit = maxPageLimit
	}

	return cursor, limit, nil
}

// BuildRoutes builds v1 endpoint routes.
func BuildRoutes(f *fiber.App, middlewares ...fiber.Handler) error {
	v1 := f.Group("/v1", middlewares...)

	v1.Get("/exps", experiments)
	v1.Get("/exp/:id", experiment)

	v1.Get("/exp/:id/metrics", metrics)
	v1.Get("/exp/:id/metric/:mid", metric)

	v1.Get("/exp/:id/artifacts", artifacts)
	v1.Get("/artifact/:rid/:aid", artifact)

	v1.Get("/registries", registries)
	v1.Get("/registry/:id", registry)
	v1.Post("/registry", createRegistry)

	v1.Get("/benchmarks", benchmarks)
	v1.Get("/benchmark/:id", benchmark)
	v1.Post("/benchmark", createBenchmark)
	v1.Put("/benchmark/:id/toggle", toggleBenchmark)
	v1.Patch("/benchmark/:id", updateBenchmark)
	v1.Delete("/benchmark/:id", deleteBenchmark)
	v1.Get("/benchmark/:id/runs", benchmarkRuns)
	v1.Get("/benchmark/:id/best", benchmarkBest)

	v1.Get("/keys", keys)
	v1.Post("/key", key)

	return nil
}
