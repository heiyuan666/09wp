package service

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
)

type SMTPConfig struct {
	Host string
	Port int
	User string
	Pass string
	From string
}

func LoadSMTPConfig() (SMTPConfig, bool, error) {
	var cfg model.SystemConfig
	if err := database.DB().Order("id ASC").First(&cfg).Error; err != nil {
		return SMTPConfig{}, false, err
	}
	out := SMTPConfig{
		Host: strings.TrimSpace(cfg.SmtpHost),
		Port: cfg.SmtpPort,
		User: strings.TrimSpace(cfg.SmtpUser),
		Pass: strings.TrimSpace(cfg.SmtpPass),
		From: strings.TrimSpace(cfg.SmtpFrom),
	}
	enabled := out.Host != "" && out.Port > 0 && out.From != ""
	return out, enabled, nil
}

func SendHTMLEmail(to string, subject string, htmlBody string) error {
	smtpCfg, enabled, err := LoadSMTPConfig()
	if err != nil {
		return err
	}
	if !enabled {
		return fmt.Errorf("smtp 未配置")
	}

	to = strings.TrimSpace(to)
	if to == "" {
		return fmt.Errorf("收件人为空")
	}

	msg := buildMIMEMessage(smtpCfg.From, to, subject, htmlBody)

	addr := net.JoinHostPort(smtpCfg.Host, fmt.Sprintf("%d", smtpCfg.Port))
	auth := smtp.PlainAuth("", smtpCfg.User, smtpCfg.Pass, smtpCfg.Host)

	// 465：implicit TLS；其它端口：优先 STARTTLS（若服务器支持）
	if smtpCfg.Port == 465 {
		return sendViaImplicitTLS(addr, smtpCfg.Host, auth, smtpCfg.From, to, msg)
	}
	return sendViaStartTLS(addr, smtpCfg.Host, auth, smtpCfg.From, to, msg)
}

func buildMIMEMessage(from string, to string, subject string, htmlBody string) []byte {
	// 最小可用 MIME（UTF-8 + HTML）
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("From: %s\r\n", from))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", to))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", encodeSubject(subject)))
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	buf.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	buf.WriteString("\r\n")
	buf.WriteString(htmlBody)
	return buf.Bytes()
}

func encodeSubject(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "通知"
	}
	// 简化：直接返回，常见 SMTP 服务可接受 UTF-8；如需严格 RFC2047 可后续完善
	return s
}

func sendViaImplicitTLS(addr string, host string, auth smtp.Auth, from string, to string, msg []byte) error {
	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 8 * time.Second}, "tcp", addr, &tls.Config{
		ServerName: host,
		MinVersion: tls.VersionTLS12,
	})
	if err != nil {
		return err
	}
	defer conn.Close()

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer c.Quit()

	if auth != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err := c.Auth(auth); err != nil {
				return err
			}
		}
	}
	if err := c.Mail(from); err != nil {
		return err
	}
	if err := c.Rcpt(to); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		_ = w.Close()
		return err
	}
	return w.Close()
}

func sendViaStartTLS(addr string, host string, auth smtp.Auth, from string, to string, msg []byte) error {
	c, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer c.Quit()

	_ = c.Hello("localhost")
	if ok, _ := c.Extension("STARTTLS"); ok {
		_ = c.StartTLS(&tls.Config{ServerName: host, MinVersion: tls.VersionTLS12})
	}

	if auth != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err := c.Auth(auth); err != nil {
				return err
			}
		}
	}

	if err := c.Mail(from); err != nil {
		return err
	}
	if err := c.Rcpt(to); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		_ = w.Close()
		return err
	}
	return w.Close()
}

