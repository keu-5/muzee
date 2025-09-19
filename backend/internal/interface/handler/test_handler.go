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

// SwaggerTest
//
//  @Summary        Test endpoint for Swagger auto-generation
//  @Description    Simple test endpoint to verify GitHub Action workflow
//  @Tags           swagger-test
//  @Produce        json
//  @Success        200 {object}    map[string]interface{}
//  @Router         /swagger-test [get]
func (h *TestHandler) SwaggerTest(c *fiber.Ctx) error {
    return c.JSON(fiber.Map{
        "message":   "Swagger auto-generation workflow test",
        "status":    "success",
        "timestamp": "2025-09-19",
    })
}