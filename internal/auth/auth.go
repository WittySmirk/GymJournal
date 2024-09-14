package auth

import (
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
	"os"
)

func CreateAuth() {
	goth.UseProviders(google.New(os.Getenv("GOOGLE_CLIENT_ID"), os.Getenv("GOOGLE_SECRET"), "http://localhost:8080/auth/google/callback"))
}
