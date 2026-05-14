package v1

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/zeddo123/mlsolid/solid/types"
)

func benchmarks(c *fiber.Ctx) error {
	ctrl := ctxController(c)

	benchmarks, err := ctrl.Benchmarks(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{ //nolint: wrapcheck
			Error: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(BenchmarksResponse{ //nolint: wrapcheck
		Benchmarks: benchmarks,
		Details:    "benchmarks retrieved successfully",
	})
}

func benchmark(c *fiber.Ctx) error {
	ctrl := ctxController(c)

	id := c.Params("id")

	bench, err := ctrl.Benchmark(c.Context(), id)
	if errors.Is(err, types.ErrNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{ //nolint: wrapcheck
			Error: err.Error(),
		})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{ //nolint: wrapcheck
			Error: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(BenchmarkResponse{ //nolint: wrapcheck
		Benchmark: *bench,
		Details:   "benchmark retrieved successfully",
	})
}

func createBenchmark(c *fiber.Ctx) error {
	ctrl := ctxController(c)

	var request CreateBenchmarkRequest

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{ //nolint: wrapcheck
			Error: err.Error(),
		})
	}

	id, created, err := ctrl.CreateBenchmark(c.Context(), request.bench())
	status := fiber.StatusCreated

	switch {
	case errors.Is(err, types.ErrBadRequest):
		status = fiber.StatusBadRequest
	case errors.Is(err, types.ErrNotFound):
		status = fiber.StatusNotFound
	case err != nil:
		status = fiber.StatusInternalServerError
	}

	if err != nil {
		return c.Status(status).JSON(ErrorResponse{ //nolint: wrapcheck
			Error: err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(CreateBenchmarkResponse{ //nolint: wrapcheck
		ID:      id,
		Created: created,
		Details: "benchmark created successfully",
	})
}

func toggleBenchmark(c *fiber.Ctx) error {
	ctrl := ctxController(c)

	id := c.Params("id")
	paused := c.Query("paused")

	toggle, err := strconv.ParseBool(paused)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{ //nolint: wrapcheck
			Error: "paused query param is malformed",
		})
	}

	err = ctrl.ToggleBenchmark(c.Context(), id, toggle)
	if errors.Is(err, types.ErrNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{ //nolint: wrapcheck
			Error: err.Error(),
		})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{ //nolint: wrapcheck
			Error: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(BenchmarkToggleResponse{ //nolint: wrapcheck
		Details: "benchmark successfully toggled",
	})
}

func updateBenchmark(c *fiber.Ctx) error {
	ctrl := ctxController(c)

	id := c.Params("id")

	var updateBenchmark types.UpdateBench

	if err := c.BodyParser(&updateBenchmark); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{ //nolint: wrapcheck
			Error: err.Error(),
		})
	}

	err := ctrl.UpdateBenchmark(c.Context(), id, updateBenchmark)
	if errors.Is(err, types.ErrNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{ //nolint: wrapcheck
			Error: err.Error(),
		})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{ //nolint: wrapcheck
			Error: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(BenchmarkUpdateResponse{ //nolint: wrapcheck
		Details: "benchmark updated successfully",
	})
}

func deleteBenchmark(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(ErrorResponse{
		Error: "delete benchmark is not implemented",
	})
}

func benchmarkRuns(c *fiber.Ctx) error {
	ctrl := ctxController(c)

	id := c.Params("id")

	runs, err := ctrl.BenchmarkRuns(c.Context(), id)
	if errors.Is(err, types.ErrNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{ //nolint: wrapcheck
			Error: err.Error(),
		})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{ //nolint: wrapcheck
			Error: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(BenchmarkRunsResponse{ //nolint: wrapcheck
		Runs:    runs,
		Details: "benchmark runs retrieved successfully",
	})
}

func benchmarkBest(c *fiber.Ctx) error {
	ctrl := ctxController(c)

	id := c.Params("id")

	var metrics []string

	if err := c.BodyParser(&metrics); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{ //nolint: wrapcheck
			Error: err.Error(),
		})
	}

	runs, err := ctrl.BestRuns(c.Context(), id, metrics...)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{ //nolint: wrapcheck
			Error: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{ //nolint: wrapcheck
		"runs":    runs,
		"details": "best models retrieved successfully",
	})
}
