package connection

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type Connection struct {
	ConnectionId string
	UserId       string
	Created      time.Time
}

func New(connectionId string, userId string) Connection {
	return Connection{
		ConnectionId: connectionId,
		UserId:       userId,
	}
}

func NewWithConnectionId(connectionId string) Connection {
	return Connection{
		ConnectionId: connectionId,
	}
}

func NewWithUserId(userId string) Connection {
	return Connection{
		UserId: userId,
	}
}

// Create add supplied Connection struct to the given DynamoDB table
func (connection Connection) Create(dynamoDbSvc *dynamodb.DynamoDB, table string) error {
	// update time
	connection.Created = time.Now()

	// marshal
	av, err := dynamodbattribute.MarshalMap(connection)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(table),
	}

	_, err = dynamoDbSvc.PutItem(input)
	return err
}

// Delete deletes connection by the provided ConnectioId
func (connection Connection) Delete(dynamoDbSvc *dynamodb.DynamoDB, table string) error {
	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"ConnectionId": {
				S: aws.String(connection.ConnectionId),
			},
		},
		TableName: aws.String(table),
	}

	_, err := dynamoDbSvc.DeleteItem(input)
	return err
}

// GetByUserId return all connections opened by the given user
func (connection Connection) GetByUserId(dynamoDbSvc *dynamodb.DynamoDB, table string, index string) ([]Connection, error) {
	input := &dynamodb.QueryInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":v1": {
				S: aws.String(connection.UserId),
			},
		},
		KeyConditionExpression: aws.String("UserId = :v1"),
		TableName:              aws.String(table),
		IndexName:              aws.String(index),
	}

	// run the query
	res, err := dynamoDbSvc.Query(input)
	if err != nil {
		return nil, err
	}

	// prepare an emty slice
	connections := []Connection{}

	// go through items
	for _, item := range res.Items {
		c := Connection{}
		err = dynamodbattribute.UnmarshalMap(item, &c)
		if err != nil {
			return nil, err
		}

		connections = append(connections, c)
	}

	return connections, nil
}
