package v1

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zeddo123/mlsolid/solid/types"
)

func metrics(ctx *fiber.Ctx) error {
	ctrl := ctxController(ctx)
	expID := ctx.Params("id")

	rs, err := ctrl.RunsFromExp(ctx.Context(), expID)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"details": err.Error(),
		})
	}

	metrics := types.UniqueMetrics(rs)

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"metrics": metrics,
		"details": "successfully retrieved experiment",
	})
}

func metric(ctx *fiber.Ctx) error {
	ctrl := ctxController(ctx)
	expID := ctx.Params("id")
	metricID := ctx.Params("mid")

	rs, err := ctrl.RunsFromExp(ctx.Context(), expID)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"details": err.Error(),
		})
	}

	metric, kind := types.CollectMetric(rs, metricID)

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"metric": metric,
		"kind":   kind,
	})
}
