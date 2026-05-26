// Command lambda runs the same chi router as an API Gateway HTTP API Lambda.
package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	chiadapter "github.com/awslabs/aws-lambda-go-api-proxy/chi"

	"github.com/hickmann/flow-service/internal/api"
	"github.com/hickmann/flow-service/internal/config"
	"github.com/hickmann/flow-service/internal/dynamodb"
)

var chiLambda *chiadapter.ChiLambdaV2

func init() {
	cfg := config.Load()
	ctx := context.Background()

	dynamo, err := dynamodb.NewClient(ctx, cfg.AWSRegion, cfg.DynamoDBEndpoint)
	if err != nil {
		log.Fatalf("dynamodb client: %v", err)
	}

	router := api.NewRouter(api.Deps{Cfg: cfg, Dynamo: dynamo})
	chiLambda = chiadapter.NewV2(router)
}

func main() {
	lambda.Start(chiLambda.ProxyWithContextV2)
}
