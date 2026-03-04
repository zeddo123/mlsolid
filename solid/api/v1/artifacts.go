package v1

import "github.com/gofiber/fiber/v2"

func artifacts(ctx *fiber.Ctx) error {
	ctrl := ctxController(ctx)
	expID := ctx.Params("id")

	runs, err := ctrl.ExpRuns(ctx.Context(), expID)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"details": err.Error(),
		})
	}

	artifacts, err := ctrl.Artifacts(ctx.Context(), runs)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
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
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"details": err.Error(),
		})
	}
	defer body.Close()

	ctx.Attachment(artifact.Name)

	return ctx.SendStream(body)
}
