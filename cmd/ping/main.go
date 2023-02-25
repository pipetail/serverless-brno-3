package main

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/events"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	apigw "github.com/pipetail/sst-websocket/pkg/apigateway"
	"github.com/pipetail/sst-websocket/pkg/notification"
	"go.uber.org/zap"
)

type handlerDependencies struct {
	Logger *zap.Logger
	SQS    *sqs.SQS
	SQSURL string
}

func main() {
	queue := os.Getenv("CONFIG_SQS_NOTIFY_CONNECTION_URL")

	// create AWS session
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// create SQS client
	sqsSvc := sqs.New(sess)

	// create a logger
	logger, _ := zap.NewProduction()

	// start the main handler
	lambda.Start(
		handler(
			handlerDependencies{
				Logger: logger,
				SQS:    sqsSvc,
				SQSURL: queue,
			},
		),
	)
}

func handler(d handlerDependencies) func(_ context.Context, req *events.APIGatewayWebsocketProxyRequest) (apigw.Response, error) {
	return func(_ context.Context, req *events.APIGatewayWebsocketProxyRequest) (apigw.Response, error) {

		// get the connection id
		connectionId := req.RequestContext.ConnectionID
		payload := req.Body

		// log the attempt
		d.Logger.Info("a new message received",
			zap.String("connectionId", connectionId),
			zap.String("payload", payload),
		)

		// send pong to the connection
		connectionNotitication := notification.ConnectionNotification{
			ConnectionId: connectionId,
			Data:         "pong, motherfucker!",
		}

		err := connectionNotitication.NotifySQS(d.SQS, d.SQSURL)
		if err != nil {

			// log and raise the error, also the client will receive
			// internal server error message
			d.Logger.Error("could not notify the connection",
				zap.String("connectionId", connectionId),
				zap.Error(err),
			)
			return apigw.InternalServerErrorResponse(), err
		}

		// all good
		return apigw.OkResponse(), nil
	}
}
