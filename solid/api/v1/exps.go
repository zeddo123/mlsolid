package v1

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zeddo123/mlsolid/solid/types"
)

func experiments(ctx *fiber.Ctx) error {
	ctrl := ctxController(ctx)

	exps, err := ctrl.Exps(ctx.Context())
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"details": "could not find exps",
		})
	}

	overview := make(map[string]any, len(exps))

	for _, exp := range exps {
		runs, err := ctrl.ExpRuns(ctx.Context(), exp)
		if err != nil {
			continue
		}

		overview[exp] = fiber.Map{
			"runs":       runs,
			"runs_count": len(runs),
		}
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"exps":    overview,
		"details": "successfully retrieved exps",
	})
}

func experiment(ctx *fiber.Ctx) error {
	ctrl := ctxController(ctx)

	expID := ctx.Params("id")

	runs, err := ctrl.ExpRuns(ctx.Context(), expID)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"details": err.Error(),
		})
	}

	rs, err := ctrl.Runs(ctx.Context(), runs)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"details": err.Error(),
		})
	}

	runsInfo := make([]runInfo, len(rs))

	for i, r := range rs {
		if r != nil {
			runsInfo[i] = runInfo{
				RunID:     r.Name,
				CreatedAt: r.Timestamp,
				Color:     r.Color,
			}
		}
	}

	out := Experiment{
		Details: "successfully retrieved experiment",
		Runs:    runsInfo,
		Metrics: types.UniqueMetrics(rs),
	}

	return ctx.Status(fiber.StatusOK).JSON(out)
}
