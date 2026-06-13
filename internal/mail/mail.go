// Package mail sends transactional emails. It uses Resend when an API key is
// configured, otherwise it logs emails to the console (handy for development).
package mail

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"strings"
	"time"
)

// Mailer sends application emails.
type Mailer interface {
	SendVerification(ctx context.Context, to, link string) error
	SendPasswordReset(ctx context.Context, to, link string) error
}

// Options configures which mail transport to use.
type Options struct {
	From         string
	ResendAPIKey string
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPass     string
}

// New picks a transport: SMTP if a host is configured, else Resend if an API
// key is set, else a console mailer that logs messages (development).
func New(o Options) Mailer {
	switch {
	case o.SMTPHost != "":
		return &smtpMailer{
			addr: o.SMTPHost + ":" + o.SMTPPort,
			host: o.SMTPHost,
			user: o.SMTPUser,
			pass: o.SMTPPass,
			from: o.From,
		}
	case o.ResendAPIKey != "":
		return &resendMailer{apiKey: o.ResendAPIKey, from: o.From, client: &http.Client{Timeout: 10 * time.Second}}
	default:
		return &consoleMailer{}
	}
}

// --- Console (dev) -------------------------------------------------------

type consoleMailer struct{}

func (c *consoleMailer) SendVerification(_ context.Context, to, link string) error {
	log.Printf("[mail:dev] verify email for %s -> %s", to, link)
	return nil
}

func (c *consoleMailer) SendPasswordReset(_ context.Context, to, link string) error {
	log.Printf("[mail:dev] password reset for %s -> %s", to, link)
	return nil
}

// --- Resend --------------------------------------------------------------

type resendMailer struct {
	apiKey string
	from   string
	client *http.Client
}

func (m *resendMailer) SendVerification(ctx context.Context, to, link string) error {
	return m.send(ctx, to, "Verify your Nut Cracker email",
		htmlButton("Welcome to Nut Cracker! 🥜", "Confirm your email address to finish setting up your account.", "Verify email", link))
}

func (m *resendMailer) SendPasswordReset(ctx context.Context, to, link string) error {
	return m.send(ctx, to, "Reset your Nut Cracker password",
		htmlButton("Password reset", "Click below to choose a new password. If you didn't request this, you can ignore this email.", "Reset password", link))
}

func (m *resendMailer) send(ctx context.Context, to, subject, html string) error {
	body, _ := json.Marshal(map[string]any{
		"from":    m.from,
		"to":      []string{to},
		"subject": subject,
		"html":    html,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+m.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("sending email: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("resend returned status %d", resp.StatusCode)
	}
	return nil
}

// --- SMTP (e.g. Gmail) ---------------------------------------------------

type smtpMailer struct {
	addr string // host:port
	host string
	user string
	pass string
	from string
}

func (m *smtpMailer) SendVerification(_ context.Context, to, link string) error {
	return m.send(to, "Verify your Nut Cracker email",
		htmlButton("Welcome to Nut Cracker! 🥜", "Confirm your email address to finish setting up your account.", "Verify email", link))
}

func (m *smtpMailer) SendPasswordReset(_ context.Context, to, link string) error {
	return m.send(to, "Reset your Nut Cracker password",
		htmlButton("Password reset", "Click below to choose a new password. If you didn't request this, you can ignore this email.", "Reset password", link))
}

func (m *smtpMailer) send(to, subject, html string) error {
	headers := map[string]string{
		"From":         m.from,
		"To":           to,
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": `text/html; charset="UTF-8"`,
	}
	var msg strings.Builder
	for k, v := range headers {
		fmt.Fprintf(&msg, "%s: %s\r\n", k, v)
	}
	msg.WriteString("\r\n")
	msg.WriteString(html)

	auth := smtp.PlainAuth("", m.user, m.pass, m.host)
	if err := smtp.SendMail(m.addr, auth, m.user, []string{to}, []byte(msg.String())); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	return nil
}

func htmlButton(heading, body, cta, link string) string {
	return fmt.Sprintf(`
<div style="font-family:system-ui,sans-serif;max-width:480px;margin:0 auto;padding:24px">
  <h2 style="color:#0f172a">%s</h2>
  <p style="color:#475569;line-height:1.5">%s</p>
  <p style="margin:28px 0">
    <a href="%s" style="background:#10b981;color:#fff;text-decoration:none;padding:12px 20px;border-radius:8px;font-weight:600">%s</a>
  </p>
  <p style="color:#94a3b8;font-size:13px">Or paste this link into your browser:<br>%s</p>
</div>`, heading, body, link, cta, link)
}
