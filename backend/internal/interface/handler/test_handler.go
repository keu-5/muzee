package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/keu-5/muzee/backend/internal/usecase"
)

// TestResponse represents the response schema for Test
type TestResponse struct {
	ID int `json:"id"`
}

type TestHandler struct {
	uc usecase.TestUsecase
}

func NewTestHandler(uc usecase.TestUsecase) *TestHandler {
	return &TestHandler{uc: uc}
}

// Create
//
//	@Summary		Create a new test
//	@Description	Creates a test record and returns it
//	@Tags			tests
//	@Produce		json
//	@Success		200	{object}	TestResponse
//	@Failure		500	{object}	map[string]string
//	@Router			/tests [post]
func (h *TestHandler) Create(c *fiber.Ctx) error {
	test, err := h.uc.CreateTest(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(test)
}

// GetAll
//
//	@Summary		List all tests
//	@Description	Returns all test records
//	@Tags			tests
//	@Produce		json
//	@Success		200	{array}		TestResponse
//	@Failure		500	{object}	map[string]string
//	@Router			/tests [get]
func (h *TestHandler) GetAll(c *fiber.Ctx) error {
	tests, err := h.uc.GetAllTests(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(tests)
}

// HealthCheck
//
//	@Summary		Health check for testing Swagger generation
//	@Description	Simple health check endpoint to test auto-swagger workflow
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Router			/tests/health [get]
func (h *TestHandler) HealthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "ok",
		"message": "Swagger auto-generation test endpoint is working",
	})
}
