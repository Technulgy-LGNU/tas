package contact

import (
	"crypto/tls"
	"errors"
	"fmt"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Host     string `toml:"smtp_host"`
	Port     int    `toml:"smtp_port"`
	Username string `toml:"smtp_username"`
	Password string `toml:"smtp_password"`
	TLSMode  string `toml:"tls_mode"`
	From     string `toml:"from"`
	To       string `toml:"to"`
}

func (c Config) Ready() bool {
	return c.Host != "" && c.Port > 0 && c.Port <= 65535 && ValidAddress(c.From) && ValidAddress(c.To) && (c.Username == "" || c.Password != "") && (c.TLSMode == "starttls" || c.TLSMode == "tls")
}
func ValidAddress(value string) bool {
	if strings.ContainsAny(value, "\r\n") {
		return false
	}
	parsed, err := mail.ParseAddress(value)
	return err == nil && parsed.Address == value && len(value) <= 254
}

type Message struct{ Name, Email, Subject, Body, Language string }

func Encode(c Config, m Message) []byte {
	subject := mime.QEncoding.Encode("utf-8", "Website contact: "+m.Subject)
	body := fmt.Sprintf("Name: %s\nEmail: %s\nLanguage: %s\n\n%s", m.Name, m.Email, m.Language, m.Body)
	body = strings.ReplaceAll(strings.ReplaceAll(body, "\r\n", "\n"), "\n", "\r\n")
	return []byte("From: " + c.From + "\r\nTo: " + c.To + "\r\nReply-To: " + m.Email + "\r\nSubject: " + subject + "\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\nContent-Transfer-Encoding: 8bit\r\n\r\n" + body + "\r\n")
}
func Send(c Config, m Message) error {
	if !c.Ready() {
		return errors.New("SMTP is not configured")
	}
	address := net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	tlsConfig := &tls.Config{ServerName: c.Host, MinVersion: tls.VersionTLS12}
	var conn net.Conn
	var err error
	if c.TLSMode == "tls" {
		conn, err = tls.DialWithDialer(dialer, "tcp", address, tlsConfig)
	} else {
		conn, err = dialer.Dial("tcp", address)
	}
	if err != nil {
		return errors.New("SMTP connection failed")
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(20 * time.Second))
	client, err := smtp.NewClient(conn, c.Host)
	if err != nil {
		return errors.New("SMTP handshake failed")
	}
	defer client.Close()
	if c.TLSMode == "starttls" {
		if err := client.StartTLS(tlsConfig); err != nil {
			return errors.New("SMTP STARTTLS failed")
		}
	}
	if c.Username != "" {
		if err := client.Auth(smtp.PlainAuth("", c.Username, c.Password, c.Host)); err != nil {
			return errors.New("SMTP authentication failed")
		}
	}
	if err := client.Mail(c.From); err != nil {
		return errors.New("SMTP sender rejected")
	}
	if err := client.Rcpt(c.To); err != nil {
		return errors.New("SMTP recipient rejected")
	}
	writer, err := client.Data()
	if err != nil {
		return errors.New("SMTP DATA failed")
	}
	if _, err := writer.Write(Encode(c, m)); err != nil {
		return errors.New("SMTP message write failed")
	}
	if err := writer.Close(); err != nil {
		return errors.New("SMTP message delivery failed")
	}
	// DATA acceptance is the delivery handoff; a subsequent QUIT failure must not prompt duplicate sends.
	_ = client.Quit()
	return nil
}
