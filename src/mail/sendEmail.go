package mail

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net/smtp"
	"tas/src/config"
)

// SendEmailPWDReset SendEmail tem is for template
func SendEmailPWDReset(email string, subject string, content string, cfg *config.CFG) error {
	var (
		auth        = smtp.PlainAuth("", cfg.Email.SenderEmail, cfg.Email.SenderEmailPassword, cfg.Email.SenderEmail)
		host        = fmt.Sprintf("%s:%d", cfg.Email.Host, 465)
		mimeHeaders = "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
		body        bytes.Buffer
		mails       = []string{email}

		t   *template.Template
		err error
	)
	t, err = template.ParseFiles("templates/resetPassword.html")
	if err != nil {
		return errors.New(fmt.Sprintf("Error parsing temaplate file: %v\n", err))
	}
	err = t.Execute(&body, struct {
		Content string
	}{
		Content: content,
	})
	if err != nil {
		return errors.New(fmt.Sprintf("Error executing template: %v\n", err))
	}
	subject = "TAS: Password reset"
	body.Write([]byte(fmt.Sprintf("Subject: %s \n%s\n\n", subject, mimeHeaders)))

	err = smtp.SendMail(host, auth, cfg.Email.SenderEmail, mails, body.Bytes())

	return nil
}

/*
func SendEmailForm(email string, Name string, formEmail string, content string, cfg *config.CFG) error {
	var (
		host        = fmt.Sprintf("%s:%d", cfg.Email.Host, 465)
		auth        = smtp.PlainAuth("", cfg.Email.SenderEmail, cfg.Email.SenderEmailPassword, cfg.Email.Host)
		mimeHeaders = "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
		body        bytes.Buffer
		mails       = []string{email}

		t   *template.Template
		err error
	)

	// Parse the template
	t, err = template.ParseFiles("templates/FormTemplate.html")
	if err != nil {
		return errors.New(fmt.Sprintf("Error parsing template file: %v\n", err))
	}

	// Create the email body
	body.Write([]byte(fmt.Sprintf("Subject: T.A.S. Form\n%s\n\n", mimeHeaders)))

	// Execute the template with the data
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

	// Send the email
	err = smtp.SendMail(host, auth, cfg.Email.SenderEmail, mails, body.Bytes())
	if err != nil {
		return errors.New(fmt.Sprintf("Error sending email: %v\n", err))
	}
	log.Printf("INFO: Email sent to %s\n", email)
	return nil
}
*/
