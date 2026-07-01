package main

import (
	"context"
	"fmt"

	"github.com/ayush00git/goth/grpc/goth"
)

func TestRefreshToken(c goth.GothServiceClient) {
	resp, err := c.RefreshToken(context.Background(), &goth.RefreshTokenRequest{
		RefreshToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiY2E1MTg1ZDAtMWU0Ni00ZDdhLWFjMGItMmE4M2IyNTU5NGRlIiwiZW1haWwiOiJ0ZXN0ZXI2NDEyQGdtYWlsLmNvbSIsImRldmljZV9maW5nZXJwcmludCI6ImdvLWNsaWVudCIsInN1YiI6ImNhNTE4NWQwLTFlNDYtNGQ3YS1hYzBiLTJhODNiMjU1OTRkZSIsImV4cCI6MTc4Mjk5NTAxOSwiaWF0IjoxNzgyOTA4NjE5LCJqdGkiOiJlODgzYTEyMC1jYjViLTQyZDItODE5ZC0zNTM2NzY1NTNkNTMifQ.o0ffs4bkOrE_Tfa9Y9Er96Vimt1FJNlx1IO7RzsSm9w",
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("refresh token revoked!\n")
	fmt.Printf("refresh token: %s\n", resp.RefreshToken)
	fmt.Printf("access token: %s\n", resp.AccessToken)
}
