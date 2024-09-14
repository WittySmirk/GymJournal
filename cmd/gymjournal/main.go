package main

import (
	"fmt"
	"github.com/WittySmirk/gymjournal/internal/auth"
	"github.com/WittySmirk/gymjournal/internal/server"
	"github.com/joho/godotenv"
	"net/http"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("failed to load env")
		return
	}

	server.CreateDb()
	auth.CreateAuth()
	mux := server.CreateRoutes()

	http.ListenAndServe(":8080", mux)
}
