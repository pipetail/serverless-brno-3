package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
	apigw "github.com/pipetail/sst-websocket/pkg/apigateway"
	"github.com/pipetail/sst-websocket/pkg/request"
	"go.uber.org/zap"
)

type handlerDependencies struct {
	Logger             *zap.Logger
	ApiGateway         *apigatewaymanagementapi.ApiGatewayManagementApi
	ApiGatewayEndpoint string
}

func main() {
	// get API Gateway endpoint
	endpoint := apigw.SanitizeURL(os.Getenv("CONFIG_API_GATEWAY_ENDPOINT"))

	// create apigateway client
	apiGatewaySess, _ := session.NewSession(&aws.Config{
		Endpoint: aws.String(endpoint),
	})
	apiGatewaySvc := apigatewaymanagementapi.New(apiGatewaySess)

	// create a logger
	logger, _ := zap.NewProduction()

	// start the main handler
	lambda.Start(handler(
		handlerDependencies{
			Logger:             logger,
			ApiGateway:         apiGatewaySvc,
			ApiGatewayEndpoint: endpoint,
		},
	))
}

func handler(d handlerDependencies) func(ctx context.Context, sqsEvent events.SQSEvent) error {
	return func(ctx context.Context, sqsEvent events.SQSEvent) error {

		for _, message := range sqsEvent.Records {
			// indicate start of the processing
			d.Logger.Info("handling delete connection request",
				zap.String("payload", message.Body),
			)

			// get the delete request
			r, err := request.DeleteConnectionFromString(message.Body)
			if err != nil {
				d.Logger.Error("could not parse request body",
					zap.String("payload", message.Body),
					zap.Error(err),
				)
				return fmt.Errorf("could not parse request: %s", err)
			}

			// send notification about the deletion into the connection
			_, err = d.ApiGateway.PostToConnection(&apigatewaymanagementapi.PostToConnectionInput{
				ConnectionId: aws.String(r.ConnectionId),
				Data:         []byte("$disconnect"),
			})
			if err != nil {
				d.Logger.Error("could not send deletion notification into the connection",
					zap.String("connectionId", r.ConnectionId),
					zap.Error(err),
				)
				return err
			}

			// delete the connection
			_, err = d.ApiGateway.DeleteConnection(&apigatewaymanagementapi.DeleteConnectionInput{
				ConnectionId: aws.String(r.ConnectionId),
			})
			if err != nil {
				d.Logger.Error("could not delete the connection",
					zap.String("connectionId", r.ConnectionId),
					zap.Error(err),
				)
				return err
			}
		}

		return nil
	}
}
