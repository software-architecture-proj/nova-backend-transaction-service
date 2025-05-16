package main

import (
	"log"
	"nova-backend-transaction-service/config" // adjust if using module path
)

func main() {
	client := config.GetClient()
	// Optional: do something with client to verify it works
	log.Println("✅ TigerBeetle client initialized.", client)
}
