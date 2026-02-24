package mail

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

// SESSender sends emails via AWS SES.
type SESSender struct {
	client    *ses.Client
	fromEmail string
}

// NewSESSender creates a real SES sender using static credentials from config.
func NewSESSender(region, fromEmail, accessKeyID, secretAccessKey string) (*SESSender, error) {
	cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("load AWS config: %w", err)
	}

	return &SESSender{
		client:    ses.NewFromConfig(cfg),
		fromEmail: fromEmail,
	}, nil
}

func (s *SESSender) SendVerificationEmail(toEmail, token, baseURL string) error {
	link := fmt.Sprintf("%s/verify?token=%s", baseURL, token)

	subject := "Verify your email"
	textBody := fmt.Sprintf("Click the link to verify your email:\n%s", link)
	htmlBody := fmt.Sprintf(
		`<p>Click the link below to verify your email address:</p>`+
			`<p><a href="%s">Verify Email</a></p>`+
			`<p>If you did not register, you can ignore this email.</p>`,
		link,
	)

	// Build a raw MIME message — uses SendRawEmail which works with
	// SES SMTP IAM users that only have ses:SendRawEmail permission.
	boundary := "----=_Boundary_pierakladnia"
	var msg strings.Builder
	fmt.Fprintf(&msg, "From: %s\r\n", s.fromEmail)
	fmt.Fprintf(&msg, "To: %s\r\n", toEmail)
	fmt.Fprintf(&msg, "Subject: %s\r\n", subject)
	fmt.Fprintf(&msg, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&msg, "Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary)
	fmt.Fprintf(&msg, "\r\n")
	// Plain text part
	fmt.Fprintf(&msg, "--%s\r\n", boundary)
	fmt.Fprintf(&msg, "Content-Type: text/plain; charset=UTF-8\r\n\r\n")
	fmt.Fprintf(&msg, "%s\r\n", textBody)
	// HTML part
	fmt.Fprintf(&msg, "--%s\r\n", boundary)
	fmt.Fprintf(&msg, "Content-Type: text/html; charset=UTF-8\r\n\r\n")
	fmt.Fprintf(&msg, "%s\r\n", htmlBody)
	// End
	fmt.Fprintf(&msg, "--%s--\r\n", boundary)

	input := &ses.SendRawEmailInput{
		Source:       aws.String(s.fromEmail),
		Destinations: []string{toEmail},
		RawMessage: &types.RawMessage{
			Data: []byte(msg.String()),
		},
	}

	_, err := s.client.SendRawEmail(context.Background(), input)
	if err != nil {
		log.Printf("SES SendRawEmail error: %v", err)
		return fmt.Errorf("send verification email: %w", err)
	}

	log.Printf("Verification email sent to %s", toEmail)
	return nil
}
