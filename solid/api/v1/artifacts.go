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
		return ctx.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{ //nolint: wrapcheck
			Error: err.Error(),
		})
	}

	if len(runs) == 0 {
		return ctx.Status(fiber.StatusNotFound).JSON(ErrorResponse{ //nolint: wrapcheck
			Error: "could not find runs linked to experiment",
		})
	}

	artifacts, err := ctrl.Artifacts(ctx.Context(), runs)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{ //nolint: wrapcheck
			Error: err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(ArtifactsResponse{ //nolint: wrapcheck
		Artifacts: artifacts,
		Details:   "successfully retrieved experiment artifacts!",
	})
}

func artifact(ctx *fiber.Ctx) error {
	ctrl := ctxController(ctx)
	runID := ctx.Params("rid")
	artifactID := ctx.Params("aid")

	artifact, body, err := ctrl.Artifact(ctx.Context(), runID, artifactID)
	if errors.Is(err, types.ErrNotFound) {
		return ctx.Status(fiber.StatusNotFound).JSON(ErrorResponse{ //nolint: wrapcheck
			Error: "artifact not found",
		})
	} else if errors.Is(err, types.ErrInternal) || err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{ //nolint: wrapcheck
			Error: "could not retrieve artifact",
		})
	}

	defer body.Close() //nolint: errcheck

	ctx.Attachment(artifact.Name)

	return ctx.SendStream(body) //nolint: wrapcheck
}
