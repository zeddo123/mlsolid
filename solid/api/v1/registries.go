package v1

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/zeddo123/mlsolid/solid/types"
)

func registries(ctx *fiber.Ctx) error {
	ctrl := ctxController(ctx)

	registries, err := ctrl.ModelRegistriesID(ctx.Context())
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err,
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"registries": registries,
		"details":    "registries retrieved successfully",
	})
}

func registry(ctx *fiber.Ctx) error {
	ctrl := ctxController(ctx)
	id := ctx.Params("id")

	reg, err := ctrl.ModelRegistry(ctx.Context(), id)
	if errors.Is(err, types.ErrNotFound) {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err,
		})
	} else if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err,
		})
	}

	infos := make(map[int]entryInfo, len(reg.Models))

	for v, entry := range reg.Models {
		infos[v+1] = entryInfo{
			CreatedAt: entry.Timestamp,
			Tags:      entry.Tags,
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
