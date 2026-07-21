package v1

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/zeddo123/mlsolid/solid/types"
)

func registries(ctx *fiber.Ctx) error {
	ctrl := ctxController(ctx)

	cursor, limit, err := parsePagination(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	registries, next, err := ctrl.ModelRegistriesIDPage(ctx.Context(), cursor, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(RegistriesResponse{
		Registries: registries,
		Cursor:     strconv.FormatUint(next, 10),
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
		Details:                 "retrieved model registry successfully",
		Name:                    reg.Name,
		LastVer:                 int64(reg.LatestVersion()),
		Tags:                    reg.Tags,
		CreatedAt:               reg.Timestamp,
		EntriesInfo:             infos,
		BenchmarkImage:          reg.BenchmarkImage,
		BenchmarkGpuPassthrough: reg.BenchmarkGpuPassthrough,
	}

	return ctx.Status(fiber.StatusOK).JSON(out)
}

func createRegistry(ctx *fiber.Ctx) error {
	ctrl := ctxController(ctx)

	var payload CreateRegistryRequest

	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	err := ctrl.CreateModelRegistry(ctx.Context(), payload.Name, types.RegistryBenchmarkOps{
		BenchmarkImage:          payload.BenchmarkImage,
		BenchmarkGpuPassthrough: payload.BenchmarkGpuPassthrough,
	})
	if errors.Is(err, types.ErrBadRequest) {
		return ctx.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: err.Error(),
		})
	} else if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	return ctx.Status(fiber.StatusCreated).JSON(CreateRegistryResponse{
		Details: "model registry created successfully",
		ID:      payload.Name,
	})
}
