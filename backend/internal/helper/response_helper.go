package helper

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error   string                   `json:"error"`
	Message string                   `json:"message"`
	Details []map[string]interface{} `json:"details,omitempty"`
}

// GetValidationMessage returns a localized validation error message
func GetValidationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "必須項目です"
	case "email":
		return "有効なメールアドレスを入力してください"
	case "min":
		return fmt.Sprintf("%s文字以上で入力してください", fe.Param())
	case "max":
		return fmt.Sprintf("%s文字以内で入力してください", fe.Param())
	case "len":
		return fmt.Sprintf("%s文字で入力してください", fe.Param())
	default:
		return "入力内容が正しくありません"
	}
}

// BuildValidationErrorResponse builds a standardized validation error response
func BuildValidationErrorResponse(err error) ErrorResponse {
	details := make([]map[string]interface{}, 0)
	for _, err := range err.(validator.ValidationErrors) {
		details = append(details, map[string]interface{}{
			"field":   strings.ToLower(err.Field()),
			"message": GetValidationMessage(err),
		})
	}
	return ErrorResponse{
		Error:   "validation_error",
		Message: "入力内容に誤りがあります",
		Details: details,
	}
}
