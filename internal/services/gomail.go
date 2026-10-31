package services

import (
	"log"
	"os"
	"strconv"

	"github.com/ayush00git/goth/internal/helpers"
	"github.com/wneessen/go-mail"
)

// SendMail creates an go-mail client and sends the mail
// to the designated gRPC server's client.
func SendMail(to, subject, body string) error {
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
	message.SetBodyString(mail.TypeTextHTML, body)
	
	// SMTP server defaults to gmail. MAIL_HOST / MAIL_PORT override it, and
	// MAIL_INSECURE=true turns off TLS and auth for a local test server (CI uses mailpit).
	host := os.Getenv("MAIL_HOST")
	if host == "" {
		host = "smtp.gmail.com"
	}
	port, err := strconv.Atoi(os.Getenv("MAIL_PORT"))
	if err != nil {
		port = 587
	}

	options := []mail.Option{mail.WithPort(port)}
	if os.Getenv("MAIL_INSECURE") == "true" {
		options = append(options, mail.WithTLSPolicy(mail.NoTLS))
	} else {
		options = append(options,
			mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
			mail.WithUsername(userName),
			mail.WithPassword(password),
		)
	}

	client, err := mail.NewClient(host, options...)
	if err != nil {
		log.Printf("Failed to create the mail client: %s", err)
	}

	if err := client.DialAndSend(message); err != nil {
		log.Printf("Failed to send mail: %s", err)
	}
	log.Printf("Mail sent to %s", to)
	return nil
}
