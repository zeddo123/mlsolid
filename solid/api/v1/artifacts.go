package v1

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/zeddo123/mlsolid/solid/types"
)

func artifacts(ctx *fiber.Ctx) error {
	ctrl := ctxController(ctx)
	expID := ctx.Params("id")

	runs, err := ctrl.ExpRuns(ctx.Context(), expID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if len(runs) == 0 {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "could not find runs linked to experiment",
		})
	}

	artifacts, err := ctrl.Artifacts(ctx.Context(), runs)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"details": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"artifacts": artifacts,
		"details":   "pulled all artifacts",
	})
}

func artifact(ctx *fiber.Ctx) error {
	ctrl := ctxController(ctx)
	runID := ctx.Params("id")
	artifactID := ctx.Params("aid")

	artifact, body, err := ctrl.Artifact(ctx.Context(), runID, artifactID)
	if errors.Is(err, types.ErrNotFound) {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "artifact not found",
		})
	} else if errors.Is(err, types.ErrInternal) || err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "could not retrieve artifact",
		})
	}

	defer body.Close()

	ctx.Attachment(artifact.Name)

	return ctx.SendStream(body)
}
