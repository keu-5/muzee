package infrastructure

import (
	"fmt"

	"github.com/keu-5/muzee/backend/config"
	"github.com/resend/resend-go/v2"
)

type EmailClient struct {
	client     *resend.Client
	logger     *Logger
	isDev      bool
	fromEmail  string
}

func NewEmailClient(cfg *config.Config, logger *Logger) *EmailClient {
	isDev := cfg.GOEnv == "development"

	var client *resend.Client
	if cfg.ResendAPIKey != "" {
		client = resend.NewClient(cfg.ResendAPIKey)
	}

	return &EmailClient{
		client:    client,
		logger:    logger,
		isDev:     isDev,
		fromEmail: cfg.ResendEmailDomain,
	}
}

func (e *EmailClient) Send(to, subject, html string) error {
	if e.isDev || e.client == nil {
		e.logger.Infow("=== メール送信 (開発モード) ===",
			"to", to,
			"subject", subject,
			"html", html,
		)
		return nil
	}

	params := &resend.SendEmailRequest{
		From:    e.fromEmail,
		To:      []string{to},
		Subject: subject,
		Html:    html,
	}

	sent, err := e.client.Emails.Send(params)
	if err != nil {
		e.logger.Errorw("Failed to send email",
			"to", to,
			"error", err,
		)
		return fmt.Errorf("failed to send email: %w", err)
	}

	e.logger.Infow("Email sent",
		"to", to,
		"email_id", sent.Id,
	)
	return nil
}
