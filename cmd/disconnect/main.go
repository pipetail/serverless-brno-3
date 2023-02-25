package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	apigw "github.com/pipetail/sst-websocket/pkg/apigateway"
	connection "github.com/pipetail/sst-websocket/pkg/connection"
	"go.uber.org/zap"
)

type handlerDependencies struct {
	DynamoDB  *dynamodb.DynamoDB
	Logger    *zap.Logger
	TableName string
}

func main() {
	// create AWS session
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// create AWS dynamodb client
	dynamoDbSvc := dynamodb.New(sess)

	// get dynamodb table name
	table := os.Getenv("CONFIG_CONNECTIONS_TABLE_ID")

	// create a logger
	logger, _ := zap.NewProduction()

	// start the main handler
	lambda.Start(
		handler(
			handlerDependencies{
				DynamoDB:  dynamoDbSvc,
				Logger:    logger,
				TableName: table,
			},
		),
	)
}

func handler(d handlerDependencies) func(_ context.Context, req *events.APIGatewayWebsocketProxyRequest) (apigw.Response, error) {
	return func(_ context.Context, req *events.APIGatewayWebsocketProxyRequest) (apigw.Response, error) {

		// get and validate the parameters
		connectionId := req.RequestContext.ConnectionID

		// log some information
		d.Logger.Info("removing connection",
			zap.String("connectionId", connectionId),
		)

		// delete the connection
		err := connection.NewWithConnectionId(connectionId).Delete(d.DynamoDB, d.TableName)
		if err != nil {
			d.Logger.Error("could not delete dynamodb record",
				zap.Error(err),
			)
			return apigw.InternalServerErrorResponse(), fmt.Errorf("could not delete DynamoDB record: %s", err)
		}

		// all good
		return apigw.OkResponse(), nil
	}
}
