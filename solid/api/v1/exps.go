package v1

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/zeddo123/mlsolid/solid/types"
)

func experiments(ctx *fiber.Ctx) error {
	ctrl := ctxController(ctx)

	cursor, limit, err := parsePagination(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	exps, next, err := ctrl.ExpsPage(ctx.Context(), cursor, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: "could not retrieve experiments",
		})
	}

	overview := make(map[string][]string, len(exps))

	for _, exp := range exps {
		runs, err := ctrl.ExpRuns(ctx.Context(), exp)
		if err != nil {
			continue
		}

		overview[exp] = runs
	}

	return ctx.Status(fiber.StatusOK).JSON(ExperimentsResponse{
		Details: "successfully retrieved exps",
		Exps:    overview,
		Cursor:  strconv.FormatUint(next, 10),
	})
}

func experiment(ctx *fiber.Ctx) error {
	ctrl := ctxController(ctx)

	expID := ctx.Params("id")

	runs, err := ctrl.ExpRuns(ctx.Context(), expID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	if len(runs) == 0 {
		return ctx.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Error: "not runs found for this experiment",
		})
	}

	rs, err := ctrl.Runs(ctx.Context(), runs)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Error: err.Error(),
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

	out := ExperimentResponse{
		Details: "successfully retrieved experiment",
		Runs:    runsInfo,
		Metrics: types.UniqueMetrics(rs),
	}

	return ctx.Status(fiber.StatusOK).JSON(out)
}
