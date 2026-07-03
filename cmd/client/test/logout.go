package main

import (
	"context"
	"fmt"

	"github.com/ayush00git/goth/grpc/goth"
)

func TestLogout(c goth.GothServiceClient) {
	resp, err := c.Logout(context.Background(), &goth.LogoutRequest{
		RefreshToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiY2E1MTg1ZDAtMWU0Ni00ZDdhLWFjMGItMmE4M2IyNTU5NGRlIiwiZW1haWwiOiJ0ZXN0ZXI2NDEyQGdtYWlsLmNvbSIsImRldmljZV9maW5nZXJwcmludCI6ImdvLWNsaWVudCIsInN1YiI6ImNhNTE4NWQwLTFlNDYtNGQ3YS1hYzBiLTJhODNiMjU1OTRkZSIsImV4cCI6MTc4MjkzMTU0MywiaWF0IjoxNzgyODQ1MTQzLCJqdGkiOiJmYzA5M2ZlYS03ZTZiLTRhMjItOGFhOC1lNTQ0MjA4ZmRlZjkifQ.Si88-ID05aUdY5vWSe_w1ZAul4SrZ1swUH_beZb0IR8",
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("logout success!\n")
	fmt.Printf("Success %t\n", resp.Success)
}
