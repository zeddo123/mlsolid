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
		return ctx.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(RegistriesResponse{
		Registries: registries,
		Details:    "registries retrieved successfully",
	})
}

func registry(ctx *fiber.Ctx) error {
	ctrl := ctxController(ctx)
	id := ctx.Params("id")

	reg, err := ctrl.ModelRegistry(ctx.Context(), id)
	if errors.Is(err, types.ErrNotFound) {
		return ctx.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Error: err.Error(),
		})
	} else if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	infos := make(map[int]entryInfo, len(reg.Models))

	for v, entry := range reg.Models {
		infos[v+1] = entryInfo{
			CreatedAt: entry.Timestamp,
			Tags:      entry.Tags,
			Name:      entry.Name,
			Run:       entry.Run,
		}
	}

	out := RegistryResponse{
		Details:     "retrieved model registry successfully",
		Name:        reg.Name,
		LastVer:     int64(reg.LatestVersion()),
		Tags:        reg.Tags,
		CreatedAt:   reg.Timestamp,
		EntriesInfo: infos,
	}

	return ctx.Status(fiber.StatusOK).JSON(out)
}
