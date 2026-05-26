// Command server runs the HTTP server for local development.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/hickmann/flow-service/internal/api"
	"github.com/hickmann/flow-service/internal/config"
	"github.com/hickmann/flow-service/internal/dynamodb"
)

func main() {
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("config: %v", err)
	}
	ctx := context.Background()

	dynamo, err := dynamodb.NewClient(ctx, cfg.AWSRegion, cfg.DynamoDBEndpoint)
	if err != nil {
		log.Fatalf("dynamodb client: %v", err)
	}

	handler := api.NewRouter(api.Deps{Cfg: cfg, Dynamo: dynamo})

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("flow-service listening on %s (table=%s, endpoint=%q)", addr, cfg.TableName, cfg.DynamoDBEndpoint)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("server: %v", err)
	}
}
