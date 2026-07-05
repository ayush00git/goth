package services

import (
	"log"

	"github.com/ayush00git/goth/internal/helpers"
	"github.com/wneessen/go-mail"
)

// SendMail creates an go-mail client and sends the mail
// to the designated gRPC server's client.
func SendMail(to, subject string) error {
	password := helpers.GetEnvVar("MAIL_PASS")
	userName := helpers.GetEnvVar("MAIL_USERNAME")

	message := mail.NewMsg()

	if err := message.From(userName); err != nil {
		log.Printf("Failed to set From address: %s", err)
	}

	if err := message.To(to); err != nil {
		log.Printf("Failed to set To address: %s", err)
	}

	message.Subject(subject)
	
	client, err := mail.NewClient(
		"smtp.example.com",
		mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
		mail.WithUsername(userName),
		mail.WithPassword(password),
	)
	if err != nil {
		log.Printf("Failed to create the mail client: %s", err)
	}

	if err := client.DialAndSend(message); err != nil {
		log.Printf("Failed to send mail: %s", err)
	}
	log.Printf("Mail sent to %s", to)
	return nil
}
