package v1

import "github.com/gofiber/fiber/v2"

func keys(c *fiber.Ctx) error {
	ctrl := ctxController(c)

	labels, err := ctrl.Redis.KeyLabels(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: "could not find all api keys",
		})
	}

	return c.Status(fiber.StatusOK).JSON(KeyLabelsResponse{
		Details: "retrieved all api key labels successfully",
		Labels:  labels,
	})
}

func key(c *fiber.Ctx) error {
	ctrl := ctxController(c)

	label := c.Query("label")
	if label == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "query param `label` not set",
		})
	}

	key, err := ctrl.GenerateKey(c.Context(), label)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: "could not generate new key",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(KeyResponse{
		Details: "api key generated successfully",
		Key:     key,
	})
}
