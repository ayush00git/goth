// totpcode prints the current TOTP code for a secret, so the e2e tests can
// act like an authenticator app. Usage: go run ./scripts/e2e/totpcode <secret>
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/pquerna/otp/totp"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: totpcode <secret>")
		os.Exit(2)
	}

	code, err := totp.GenerateCode(os.Args[1], time.Now())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Print(code)
}
