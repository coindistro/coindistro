package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"time"

	"go.uber.org/zap"
)

// Message represents an email message.
type Message struct {
	From        string            `json:"from"`
	To          []string          `json:"to"`
	CC          []string          `json:"cc,omitempty"`
	BCC         []string          `json:"bcc,omitempty"`
	Subject     string            `json:"subject"`
	PlainText   string            `json:"plain_text,omitempty"`
	HTML        string            `json:"html,omitempty"`
	Attachments []Attachment      `json:"attachments,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
}

// Attachment represents an email attachment.
type Attachment struct {
	Filename string `json:"filename"`
	Content  []byte `json:"content"`
	MIMEType string `json:"mime_type"`
}

// Sender defines the interface for sending emails.
type Sender interface {
	// Send sends an email message.
	Send(ctx context.Context, msg Message) error
	// SendBatch sends multiple email messages.
	SendBatch(ctx context.Context, messages []Message) error
	// Close closes the sender connection.
	Close() error
}

// SMTPConfig holds SMTP server configuration.
type SMTPConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
	FromName string `mapstructure:"from_name"`
	UseTLS   bool   `mapstructure:"use_tls"`
}

// SMTPSender sends emails via SMTP.
type SMTPSender struct {
	config SMTPConfig
	logger *zap.Logger
}

// NewSMTPSender creates a new SMTP email sender.
func NewSMTPSender(config SMTPConfig, logger *zap.Logger) *SMTPSender {
	return &SMTPSender{
		config: config,
		logger: logger,
	}
}

// Send sends an email via SMTP.
func (s *SMTPSender) Send(ctx context.Context, msg Message) error {
	if msg.From == "" {
		msg.From = s.config.From
	}

	start := time.Now()
	err := s.sendSMTP(msg)
	duration := time.Since(start)

	if err != nil {
		s.logger.Error("email send failed",
			zap.Strings("to", msg.To),
			zap.String("subject", msg.Subject),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return fmt.Errorf("failed to send email: %w", err)
	}

	s.logger.Info("email sent",
		zap.Strings("to", msg.To),
		zap.String("subject", msg.Subject),
		zap.Duration("duration", duration),
	)
	return nil
}

// SendBatch sends multiple emails via SMTP.
func (s *SMTPSender) SendBatch(ctx context.Context, messages []Message) error {
	for _, msg := range messages {
		if err := s.Send(ctx, msg); err != nil {
			return err
		}
	}
	return nil
}

// Close implements the Sender interface.
func (s *SMTPSender) Close() error {
	return nil
}

func (s *SMTPSender) sendSMTP(msg Message) error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	// Build auth
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)

	// Build email body
	body := s.buildBody(msg)

	// Connect and send
	if s.config.UseTLS {
		return s.sendTLS(addr, auth, msg.From, msg.To, body)
	}
	return smtp.SendMail(addr, auth, msg.From, msg.To, body)
}

func (s *SMTPSender) sendTLS(addr string, auth smtp.Auth, from string, to []string, body []byte) error {
	tlsConfig := &tls.Config{
		ServerName: s.config.Host,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return err
	}

	client, err := smtp.NewClient(conn, s.config.Host)
	if err != nil {
		return err
	}
	defer client.Close()

	if err = client.Auth(auth); err != nil {
		return err
	}

	if err = client.Mail(from); err != nil {
		return err
	}

	for _, recipient := range to {
		if err = client.Rcpt(recipient); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	_, err = w.Write(body)
	if err != nil {
		return err
	}

	return w.Close()
}

func (s *SMTPSender) buildBody(msg Message) []byte {
	body := fmt.Sprintf("From: %s\r\n", msg.From)
	body += fmt.Sprintf("To: %s\r\n", joinAddresses(msg.To))
	if len(msg.CC) > 0 {
		body += fmt.Sprintf("Cc: %s\r\n", joinAddresses(msg.CC))
	}
	body += fmt.Sprintf("Subject: %s\r\n", msg.Subject)
	body += "MIME-Version: 1.0\r\n"
	body += "Content-Type: text/html; charset=\"UTF-8\"\r\n"
	body += fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z))

	// Custom headers
	for k, v := range msg.Headers {
		body += fmt.Sprintf("%s: %s\r\n", k, v)
	}

	body += "\r\n"
	if msg.HTML != "" {
		body += msg.HTML
	} else if msg.PlainText != "" {
		body += msg.PlainText
	}

	return []byte(body)
}

func joinAddresses(addrs []string) string {
	result := ""
	for i, addr := range addrs {
		if i > 0 {
			result += ", "
		}
		result += addr
	}
	return result
}

// NoopSender is a no-op email sender for development/testing.
type NoopSender struct {
	logger *zap.Logger
}

// NewNoopSender creates a new no-op email sender.
func NewNoopSender(logger *zap.Logger) *NoopSender {
	return &NoopSender{logger: logger}
}

// Send logs the email without actually sending it.
func (s *NoopSender) Send(ctx context.Context, msg Message) error {
	s.logger.Info("email (noop)",
		zap.Strings("to", msg.To),
		zap.String("subject", msg.Subject),
	)
	return nil
}

// SendBatch logs multiple emails without sending them.
func (s *NoopSender) SendBatch(ctx context.Context, messages []Message) error {
	for _, msg := range messages {
		_ = s.Send(ctx, msg)
	}
	return nil
}

// Close implements the Sender interface.
func (s *NoopSender) Close() error {
	return nil
}

// Predefined email template types for the Coindistro platform.
const (
	EmailTypeVerification     = "verification"
	EmailTypePasswordReset    = "password_reset"
	EmailTypeWelcome          = "welcome"
	EmailTypeKYCApproved      = "kyc_approved"
	EmailTypeKYCRejected      = "kyc_rejected"
	EmailTypeDepositConfirmed = "deposit_confirmed"
	EmailTypeWithdrawal       = "withdrawal"
	EmailTypeSignalAlert      = "signal_alert"
	EmailTypeMerchantApproved = "merchant_approved"
	EmailTypeSecurityAlert    = "security_alert"
	EmailTypeNewsletter       = "newsletter"
)
