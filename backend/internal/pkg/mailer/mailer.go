package mailer

import (
	"context"
	"fmt"
	"log/slog"
	"net/smtp"
	"net/url"
	"strings"
)

// Config configures the SMTP mailer. When SMTPHost is empty the mailer is
// disabled and, instead of sending, logs the message at Info level (intended
// for local development so flows work without an SMTP server).
type Config struct {
	SMTPHost    string
	SMTPPort    int
	Username    string
	Password    string
	From        string
	FromName    string
	FrontendURL string
	Log         *slog.Logger
}

type Mailer struct {
	host        string
	port        int
	user        string
	pass        string
	from        string
	frontendURL string
	log         *slog.Logger
}

func New(cfg Config) *Mailer {
	from := cfg.From
	if from == "" {
		from = "no-reply@schemahub.dev"
	}
	if cfg.FromName != "" {
		from = fmt.Sprintf("%s <%s>", cfg.FromName, from)
	}
	if cfg.SMTPPort == 0 {
		cfg.SMTPPort = 587
	}
	return &Mailer{
		host:        cfg.SMTPHost,
		port:        cfg.SMTPPort,
		user:        cfg.Username,
		pass:        cfg.Password,
		from:        from,
		frontendURL: strings.TrimSuffix(cfg.FrontendURL, "/"),
		log:         cfg.Log,
	}
}

// Enabled reports whether SMTP is configured for real delivery.
func (m *Mailer) Enabled() bool { return m.host != "" }

func (m *Mailer) SendVerificationEmail(ctx context.Context, to, token string) error {
	link := fmt.Sprintf("%s/verify-email?token=%s", m.frontendURL, url.QueryEscape(token))
	subject := "Verify your email"
	text := fmt.Sprintf("Welcome to SchemaHub! Verify your email to activate your account.\n\n%s", link)
	html := fmt.Sprintf(
		`<p>Welcome to SchemaHub! Verify your email to activate your account.</p><p><a href="%s">Verify my email</a></p><p>Or paste this link: %s</p>`,
		link, link,
	)
	return m.send(ctx, to, subject, text, html)
}

func (m *Mailer) SendPasswordResetEmail(ctx context.Context, to, token string) error {
	link := fmt.Sprintf("%s/forgot-password/reset?token=%s", m.frontendURL, url.QueryEscape(token))
	subject := "Reset your password"
	text := fmt.Sprintf("We received a request to reset your SchemaHub password.\n\nOpen this link to choose a new one:\n%s\n\nIf you didn't request this, you can ignore this email.", link)
	html := fmt.Sprintf(
		`<p>We received a request to reset your SchemaHub password.</p><p><a href="%s">Choose a new password</a></p><p>If you didn't request this, you can ignore this email.</p>`,
		link,
	)
	return m.send(ctx, to, subject, text, html)
}

func (m *Mailer) SendInvitationEmail(ctx context.Context, to, projectName, token string) error {
	link := fmt.Sprintf("%s/invite/accept?token=%s", m.frontendURL, url.QueryEscape(token))
	subject := fmt.Sprintf("You've been invited to join %s on SchemaHub", projectName)
	text := fmt.Sprintf(
		"Someone invited you to join the project \"%s\" on SchemaHub.\n\nAccept the invitation here:\n%s\n\nThis link expires in 7 days. If you don't have an account yet, you'll be able to create one in one click.",
		projectName, link,
	)
	html := fmt.Sprintf(
		`<p>Someone invited you to join the project <strong>%s</strong> on SchemaHub.</p><p><a href="%s">Accept invitation</a></p><p>Or paste this link: %s</p><p>This link expires in 7 days.</p>`,
		projectName, link, link,
	)
	return m.send(ctx, to, subject, text, html)
}

func (m *Mailer) send(ctx context.Context, to, subject, text, html string) error {
	if !m.Enabled() {
		if m.log != nil {
			m.log.Info("email not sent: SMTP not configured (dev mode)",
				"to", to, "subject", subject)
			m.log.Info("dev email content", "to", to, "subject", subject, "text", text)
		}
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	msg := strings.Join([]string{
		"From: " + m.from,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
		"",
		html,
	}, "\r\n")

	addr := fmt.Sprintf("%s:%d", m.host, m.port)
	var auth smtp.Auth
	if m.user != "" {
		auth = smtp.PlainAuth("", m.user, m.pass, m.host)
	}
	if err := smtp.SendMail(addr, auth, m.from, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("sending email to %s: %w", to, err)
	}
	return nil
}
