package config

import (
	"log"

	tb "github.com/tigerbeetle/tigerbeetle-go"
	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

func GetClient() tb.Client {
	clusterID := types.ToUint128(0)
	client, err := tb.NewClient(clusterID, []string{"3000"})
	if err != nil {
		log.Fatalf("❌ Failed to connect to TigerBeetle: %v", err)
	}
	return client
}
