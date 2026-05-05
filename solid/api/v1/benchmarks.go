package v1

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/zeddo123/mlsolid/solid/types"
)

func benchmarks(c *fiber.Ctx) error {
	ctrl := ctxController(c)

	benchmarks, err := ctrl.Benchmarks(c.Context())
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"benchmarks": benchmarks,
		"details":    "benchmarks retrieved successfully",
	})
}

func benchmark(c *fiber.Ctx) error {
	ctrl := ctxController(c)

	id := c.Params("id")

	bench, err := ctrl.Benchmark(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"bench":   bench,
		"details": "benchmark retrieved successfully",
	})
}

func createBenchmark(c *fiber.Ctx) error {
	ctrl := ctxController(c)

	var request CreateBenchmarkRequest

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err,
		})
	}

	id, created, err := ctrl.CreateBenchmark(c.Context(), types.Bench{ //nolint: exhaustruct
		Name:           request.Name,
		EagerStart:     request.EagerStart,
		AutoTag:        request.AutoTag,
		Tag:            request.Tag,
		DecisionMetric: request.DecisionMetric,
		Registries:     request.Registries,
		Metrics:        request.Metrics,
		DatasetName:    request.DatasetName,
		DatasetURL:     request.DatasetURL,
		FromS3:         request.DatasetFromS3,
		Timestamp:      time.Now(),
	})
	if !created {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err,
		})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":      id,
		"details": "benchmark created",
	})
}

func toogleBenchmark(c *fiber.Ctx) error {
	ctrl := ctxController(c)

	id := c.Params("id")
	paused := c.Query("paused")

	toggle, err := strconv.ParseBool(paused)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "paused query param is malformed",
		})
	}

	err = ctrl.ToggleBenchmark(c.Context(), id, toggle)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"details": "benchmark successfully toggled",
	})
}

func updateBenchmark(c *fiber.Ctx) error {
	ctrl := ctxController(c)

	id := c.Params("id")

	var updateBenchmark types.UpdateBench

	if err := c.BodyParser(&updateBenchmark); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err,
		})
	}

	err := ctrl.UpdateBenchmark(c.Context(), id, updateBenchmark)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"details": "benchmark updated successfully",
	})
}

func deleteBenchmark(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "delete benchmark is not implemented",
	})
}

func benchmarkRuns(c *fiber.Ctx) error {
	ctrl := ctxController(c)

	id := c.Params("id")

	runs, err := ctrl.BenchmarkRuns(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"runs":    runs,
		"details": "benchmark runs retrieved successfully",
	})
}

func benchmarkBest(c *fiber.Ctx) error {
	ctrl := ctxController(c)

	id := c.Params("id")

	var metrics []string

	if err := c.BodyParser(&metrics); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err,
		})
	}

	runs, err := ctrl.BestRuns(c.Context(), id, metrics...)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"runs":    runs,
		"details": "best models retrieved successfully",
	})
}
