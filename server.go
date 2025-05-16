package main

import (
	"log"
	"nova-backend-transaction-service/config" // adjust if using module path
)

func main() {
	client := config.GetClient()
	log.Println("✅ TigerBeetle client initialized.", client)

}
