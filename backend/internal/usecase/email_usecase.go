package usecase

import (
	"fmt"

	"github.com/keu-5/muzee/backend/internal/infrastructure"
)

type EmailUsecase interface {
	SendVerificationCode(email, code string) error
}

type emailUsecase struct {
	emailClient *infrastructure.EmailClient
}

func NewEmailUsecase(emailClient *infrastructure.EmailClient) EmailUsecase {
	return &emailUsecase{
		emailClient: emailClient,
	}
}

func (u *emailUsecase) SendVerificationCode(email, code string) error {
	subject := "【Muzee】確認コード"
	html := fmt.Sprintf(`
		<div style="font-family: sans-serif; max-width: 600px; margin: 0 auto;">
			<h2>確認コード</h2>
			<p>以下の確認コードを入力してください：</p>
			<div style="background-color: #f5f5f5; padding: 20px; text-align: center; font-size: 32px; font-weight: bold; letter-spacing: 8px;">
				%s
			</div>
			<p style="color: #666; font-size: 14px;">
				※このコードの有効期限は15分です<br>
				※心当たりがない場合は、このメールを無視してください
			</p>
		</div>
	`, code)

	return u.emailClient.Send(email, subject, html)
}