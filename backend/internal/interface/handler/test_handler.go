package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/keu-5/muzee/backend/internal/usecase"
)

type TestHandler struct {
	uc usecase.TestUsecase
}

func NewTestHandler(uc usecase.TestUsecase) *TestHandler {
	return &TestHandler{uc: uc}
}

func (h *TestHandler) Create(c *fiber.Ctx) error {
	test, err := h.uc.CreateTest(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(test)
}

func (h *TestHandler) GetAll(c *fiber.Ctx) error {
	tests, err := h.uc.GetAllTests(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(tests)
}
