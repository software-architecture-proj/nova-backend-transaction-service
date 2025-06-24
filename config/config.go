package config

import (
	"log"
	"net"
	"os"

	tb "github.com/tigerbeetle/tigerbeetle-go"
	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

func GetClient() tb.Client {
	clusterID := types.ToUint128(0)

	tigerbeetleHost := os.Getenv("TIGERBEETLE_HOST") // Read TigerBeetle host from environment variable. Check dock-compose.yml (the big one)
	if tigerbeetleHost == "" {
		tigerbeetleHost = "192.168.50.254:3000" // Default for MY PC, you may need to change this by using IP in go run main.go
	}
	log.Println("Connecting to TigerBeetle at ", tigerbeetleHost)
	client, err := tb.NewClient(clusterID, []string{tigerbeetleHost})
	if err != nil {
		log.Fatalf("❌ Failed to connect to TigerBeetle: %v", err)
	}
	return client
}

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}
