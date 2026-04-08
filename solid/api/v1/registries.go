package v1

import (
	"github.com/gofiber/fiber/v2"
)

func registries(ctx *fiber.Ctx) error {
	_ = ctxController(ctx)

	return nil
}

func registry(ctx *fiber.Ctx) error {
	ctrl := ctxController(ctx)
	id := ctx.Params("id")

	reg, err := ctrl.ModelRegistry(ctx.Context(), id)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"details": err.Error(),
		})
	}

	infos := make(map[int]entryInfo, len(reg.Models))

	for v, entry := range reg.Models {
		infos[v+1] = entryInfo{
			CreatedAt: entry.Timestamp,
		}
	}

	out := Registry{
		Details:     "retrieved model registry successfully",
		Name:        reg.Name,
		LastVer:     int64(reg.LatestVersion()),
		Tags:        reg.Tags,
		CreatedAt:   reg.Timestamp,
		EntriesInfo: infos,
	}

	return ctx.Status(fiber.StatusOK).JSON(out)
}
