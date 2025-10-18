package helper

import (
	"fmt"

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
