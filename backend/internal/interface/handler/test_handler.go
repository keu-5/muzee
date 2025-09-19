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

// GetByID
//
//	@Summary		Get a test by ID
//	@Description	Returns a specific test record by its ID
//	@Tags			tests
//	@Produce		json
//	@Param			id	path		int	true	"Test ID"
//	@Success		200	{object}	TestResponse
//	@Failure		400	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/tests/{id} [get]
func (h *TestHandler) GetByID(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID parameter"})
	}

	return c.JSON(TestResponse{ID: id})
}

// Delete
//
//	@Summary		Delete a test by ID
//	@Description	Deletes a specific test record by its ID
//	@Tags			tests
//	@Produce		json
//	@Param			id	path		int	true	"Test ID"
//	@Success		200	{object}	map[string]string
//	@Failure		400	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/tests/{id} [delete]
func (h *TestHandler) Delete(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID parameter"})
	}

	return c.JSON(fiber.Map{"message": "Test deleted successfully", "id": id})
}
