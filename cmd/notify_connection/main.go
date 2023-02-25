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
	"github.com/pipetail/sst-websocket/pkg/notification"
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

		// we can handle even the bigger batches but the ideal
		// batch is 1 to send the message as soon as possible
		// especially in some smaller applications with low
		// traffic
		for _, message := range sqsEvent.Records {
			// indicate start of the processing
			d.Logger.Info("handling notification",
				zap.String("payload", message.Body),
			)

			// get the notification
			n, err := notification.ConnectionFromString(message.Body)
			if err != nil {
				d.Logger.Error("could not parse notification body",
					zap.String("payload", message.Body),
					zap.Error(err),
				)
				return fmt.Errorf("could not parse notification: %s", err)
			}

			// send message to connection
			_, err = d.ApiGateway.PostToConnection(&apigatewaymanagementapi.PostToConnectionInput{
				ConnectionId: aws.String(n.ConnectionId),
				Data:         []byte(n.Data),
			})
			if err != nil {
				// fail if connection was not notified, maybe the message
				// will be eventually delivered
				d.Logger.Error("could not notify the connection",
					zap.String("connectionId", n.ConnectionId),
					zap.Error(err),
				)
				return err
			}
		}

		return nil
	}
}
