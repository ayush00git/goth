package helpers

import (
	"fmt"

	"gopkg.in/gomail.v2"
)

func SendEmail (to, subject, body string) error {

	senderEmail := GetEnvVar("SENDER_EMAIL")
	appPass := GetEnvVar("GMAIL_APP_PASSWORD")
	
	// Set up a new message
	m := gomail.NewMessage()
	m.SetHeader("From", senderEmail)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	// set up the dialer
	d := gomail.NewDialer("smtp.gmail.com", 587, senderEmail, appPass)

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("could not send email %s", err)
	}
	return nil
}
