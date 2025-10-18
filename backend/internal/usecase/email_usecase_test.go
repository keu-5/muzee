package usecase

import (
	"errors"
	"strings"
	"testing"
)

// Mock EmailClient
type mockEmailClient struct {
	sendFunc func(to, subject, html string) error
}

func (m *mockEmailClient) Send(to, subject, html string) error {
	if m.sendFunc != nil {
		return m.sendFunc(to, subject, html)
	}
	return nil
}

func TestNewEmailUsecase(t *testing.T) {
	mockClient := &mockEmailClient{}
	usecase := NewEmailUsecase(mockClient)

	if usecase == nil {
		t.Fatal("Expected usecase to be non-nil")
	}
}

func TestSendVerificationCode(t *testing.T) {
	tests := []struct {
		name       string
		email      string
		code       string
		mockSend   func(to, subject, html string) error
		wantErr    bool
		validateFn func(t *testing.T, to, subject, html string)
	}{
		{
			name:  "successful send",
			email: "test@example.com",
			code:  "123456",
			mockSend: func(to, subject, html string) error {
				return nil
			},
			wantErr: false,
			validateFn: func(t *testing.T, to, subject, html string) {
				if to != "test@example.com" {
					t.Errorf("Expected email to be 'test@example.com', got '%s'", to)
				}
				if subject != "【Muzee】確認コード" {
					t.Errorf("Expected subject to be '【Muzee】確認コード', got '%s'", subject)
				}
				if !strings.Contains(html, "123456") {
					t.Error("Expected HTML to contain the verification code '123456'")
				}
				if !strings.Contains(html, "確認コード") {
					t.Error("Expected HTML to contain '確認コード'")
				}
				if !strings.Contains(html, "15分") {
					t.Error("Expected HTML to contain expiration message '15分'")
				}
			},
		},
		{
			name:  "email client error",
			email: "error@example.com",
			code:  "999999",
			mockSend: func(to, subject, html string) error {
				return errors.New("failed to send email")
			},
			wantErr: true,
		},
		{
			name:  "empty email",
			email: "",
			code:  "123456",
			mockSend: func(to, subject, html string) error {
				return nil
			},
			wantErr: false,
			validateFn: func(t *testing.T, to, subject, html string) {
				if to != "" {
					t.Errorf("Expected empty email, got '%s'", to)
				}
			},
		},
		{
			name:  "empty code",
			email: "test@example.com",
			code:  "",
			mockSend: func(to, subject, html string) error {
				return nil
			},
			wantErr: false,
			validateFn: func(t *testing.T, to, subject, html string) {
				// HTML should still be generated, even with empty code
				if !strings.Contains(html, "確認コード") {
					t.Error("Expected HTML to contain '確認コード'")
				}
			},
		},
		{
			name:  "long verification code",
			email: "test@example.com",
			code:  "ABCDEF123456",
			mockSend: func(to, subject, html string) error {
				return nil
			},
			wantErr: false,
			validateFn: func(t *testing.T, to, subject, html string) {
				if !strings.Contains(html, "ABCDEF123456") {
					t.Error("Expected HTML to contain the verification code 'ABCDEF123456'")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedTo, capturedSubject, capturedHTML string
			mockClient := &mockEmailClient{
				sendFunc: func(to, subject, html string) error {
					capturedTo = to
					capturedSubject = subject
					capturedHTML = html
					return tt.mockSend(to, subject, html)
				},
			}
			usecase := NewEmailUsecase(mockClient)

			err := usecase.SendVerificationCode(tt.email, tt.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("SendVerificationCode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.validateFn != nil {
				tt.validateFn(t, capturedTo, capturedSubject, capturedHTML)
			}
		})
	}
}

func TestSendVerificationCodeHTMLStructure(t *testing.T) {
	var capturedHTML string
	mockClient := &mockEmailClient{
		sendFunc: func(to, subject, html string) error {
			capturedHTML = html
			return nil
		},
	}
	usecase := NewEmailUsecase(mockClient)

	code := "123456"
	err := usecase.SendVerificationCode("test@example.com", code)
	if err != nil {
		t.Fatalf("SendVerificationCode() failed: %v", err)
	}

	// Verify HTML structure contains expected elements
	expectedElements := []string{
		"<div",
		"<h2>",
		"確認コード",
		code,
		"15分",
		"font-family",
		"background-color",
	}

	for _, element := range expectedElements {
		if !strings.Contains(capturedHTML, element) {
			t.Errorf("HTML does not contain expected element: %s", element)
		}
	}
}
