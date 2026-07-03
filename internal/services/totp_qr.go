package services

import (
	"fmt"
	"net/url"
)

func GenerateTOTPQRURL(email, totp_secret string) string {
	return fmt.Sprintf("otpauth://totp/goth:%s?secret=%s&issuer=goth",
		url.QueryEscape(email),
		totp_secret,
	)
}
