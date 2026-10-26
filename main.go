package main

import (
	"log"
	"os"

	"turkbot"
)

func envOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func main() {
	host := envOr("TB_BIND", "127.0.0.1")
	port := envOr("TB_PORT", "8000")

	if err := turkbot.Run(host, port); err != nil {
		log.Fatal(err)
	}
}
