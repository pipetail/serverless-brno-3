package apigw

import (
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

// Response is a typedef for the response type provided by the SDK.
type Response = events.APIGatewayProxyResponse

// InternalServerErrorResponse returns an Amazon API Gateway Proxy Response configured with the correct HTTP status
// code.
func InternalServerErrorResponse() Response {
	return Response{StatusCode: http.StatusInternalServerError}
}

// BadRequestResponse returns an Amazon API Gateway Proxy Response configured with the correct HTTP status code.
func BadRequestResponse() Response {
	return Response{StatusCode: http.StatusBadRequest}
}

// OkResponse returns an Amazon API Gateway Proxy Response configured with the correct HTTP status code.
func OkResponse() Response {
	return Response{StatusCode: http.StatusOK}
}
