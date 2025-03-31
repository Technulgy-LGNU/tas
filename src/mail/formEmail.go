package mail

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net/smtp"
	"tas/src/config"
)

func SendEmailForm(email string, Name string, formEmail string, content string, cfg *config.CFG) error {
	var (
		to          = []string{email}
		body        bytes.Buffer
		mimeHeaders = "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"

		err error
	)

	// Set the subject and headers
	body.Write([]byte(fmt.Sprintf("Subject: Technulgy Form \n%s\n\n", mimeHeaders)))

	// Parse the template file
	t, err := template.ParseFiles("templates/FormTemplate.html")
	if err != nil {
		return errors.New(fmt.Sprintf("Error parsing template file: %v\n", err))
	}

	err = t.Execute(&body, struct {
		Name    string
		Email   string
		Content string
	}{
		Name:    Name,
		Email:   formEmail,
		Content: content,
	})
	if err != nil {
		return errors.New(fmt.Sprintf("Error executing template: %v\n", err))
	}

	// Authentication.
	auth := smtp.PlainAuth("", cfg.Email.SenderEmail, cfg.Email.SenderEmailPassword, cfg.Email.Host)
	// Sending email.
	err = smtp.SendMail(cfg.Email.Host+":587", auth, cfg.Email.SenderEmail, to, body.Bytes())
	if err != nil {
		return errors.New(fmt.Sprintf("Error sending email: %v\n", err))
	}
	return nil
}
