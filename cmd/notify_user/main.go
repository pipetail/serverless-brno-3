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
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/sqs"
	apigw "github.com/pipetail/sst-websocket/pkg/apigateway"
	"github.com/pipetail/sst-websocket/pkg/connection"
	"github.com/pipetail/sst-websocket/pkg/notification"
	"go.uber.org/zap"
)

type handlerDependencies struct {
	DynamoDB           *dynamodb.DynamoDB
	IndexName          string
	TableName          string
	Logger             *zap.Logger
	ApiGateway         *apigatewaymanagementapi.ApiGatewayManagementApi
	ApiGatewayEndpoint string
	SQS                *sqs.SQS
	SQSURL             string
}

func main() {
	// get dynamodb table and index name
	table := os.Getenv("CONFIG_CONNECTIONS_TABLE_ID")
	index := os.Getenv("CONFIG_USER_ID_INDEX_NAME")
	queue := os.Getenv("CONFIG_SQS_NOTIFY_CONNECTION_URL")

	// get API Gateway endpoint
	endpoint := apigw.SanitizeURL(os.Getenv("CONFIG_API_GATEWAY_ENDPOINT"))

	// create AWS session
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// create apigateway client
	apiGatewaySess, _ := session.NewSession(&aws.Config{
		Endpoint: aws.String(endpoint),
	})
	apiGatewaySvc := apigatewaymanagementapi.New(apiGatewaySess)

	// create AWS dynamodb client
	dynamoDbSvc := dynamodb.New(sess)

	// create SQS client
	sqsSvc := sqs.New(sess)

	// create a logger
	logger, _ := zap.NewProduction()

	// start the main handler
	lambda.Start(handler(
		handlerDependencies{
			DynamoDB:           dynamoDbSvc,
			IndexName:          index,
			TableName:          table,
			Logger:             logger,
			ApiGateway:         apiGatewaySvc,
			ApiGatewayEndpoint: endpoint,
			SQS:                sqsSvc,
			SQSURL:             queue,
		},
	))
}

func handler(d handlerDependencies) func(ctx context.Context, sqsEvent events.SQSEvent) error {
	return func(ctx context.Context, sqsEvent events.SQSEvent) error {
		for _, message := range sqsEvent.Records {
			// indicate start of the processing
			d.Logger.Info("handling notification",
				zap.String("payload", message.Body),
			)

			// get the notification
			n, err := notification.UserFromString(message.Body)
			if err != nil {
				d.Logger.Error("could not parse notification body",
					zap.String("payload", message.Body),
					zap.Error(err),
				)
				return fmt.Errorf("could not parse notification: %s", err)
			}

			// get all conections for the provided user
			conns, err := connection.NewWithUserId(n.UserId).GetByUserId(d.DynamoDB, d.TableName, d.IndexName)
			if err != nil {
				d.Logger.Error("could get list of connections",
					zap.Error(err),
				)
				return fmt.Errorf("could not get connections: %s", err)
			}

			// process all connections associated with the user and send message
			// to the notifyconnecion lambda function
			for _, c := range conns {
				d.Logger.Info("handling connection",
					zap.String("userId", n.UserId),
					zap.String("connectionId", c.ConnectionId),
				)

				connectionNotitication := notification.ConnectionNotification{
					ConnectionId: c.ConnectionId,
					Data:         n.Data,
				}

				err = connectionNotitication.NotifySQS(d.SQS, d.SQSURL)
				if err != nil {
					d.Logger.Error("could not notify the connection",
						zap.String("connectionId", c.ConnectionId),
						zap.Error(err),
					)
					return err
				}
			}
		}

		return nil
	}
}
