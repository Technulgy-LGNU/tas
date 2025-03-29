package mail

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net/smtp"
	"tas/src/config"
)

// SendEmail tem is for template
func SendEmail(email string, subject string, content string, tem string, cfg *config.CFG) error {
	var (
		auth        = smtp.PlainAuth("", cfg.Email.SenderEmail, cfg.Email.SenderEmailPassword, cfg.Email.SenderEmail)
		host        = fmt.Sprintf("%s:%d", cfg.Email.Host, 465)
		mimeHeaders = "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
		body        bytes.Buffer
		mails       = []string{email}

		t   *template.Template
		err error
	)
	switch tem {
	case "form":
		break
	case "system":
		break
	case "resetPassword":
		t, err = template.ParseFiles("templates/resetPassword.html")
		if err != nil {
			return errors.New(fmt.Sprintf("Error parsing temaplate file: %v\n", err))
		}
		err := t.Execute(&body, struct {
			Content string
		}{
			Content: content,
		})
		if err != nil {
			return errors.New(fmt.Sprintf("Error executing template: %v\n", err))
		}
		subject = "TAS: Password reset"
		body.Write([]byte(fmt.Sprintf("Subject: %s \n%s\n\n", subject, mimeHeaders)))
		break
	case "accountCreated":
		break
	}

	err = smtp.SendMail(host, auth, cfg.Email.SenderEmail, mails, body.Bytes())

	return nil
}
