package notification

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type UserNotification struct {
	UserId string `json:"userId"`
	Data   string `json:"data"`
}

type ConnectionNotification struct {
	ConnectionId string `json:"connectionId"`
	Data         string `json:"data"`
}

// UserFromString decodes json to UserNotification
func UserFromString(notification string) (UserNotification, error) {
	u := UserNotification{}
	err := json.Unmarshal([]byte(notification), &u)
	return u, err
}

// ConnectionFromString decodes json to ConnectionNotification
func ConnectionFromString(notification string) (ConnectionNotification, error) {
	u := ConnectionNotification{}
	err := json.Unmarshal([]byte(notification), &u)
	return u, err
}

func (n ConnectionNotification) NotifySQS(sqsSvc *sqs.SQS, url string) error {
	// serialize ConnectionNotification
	data, err := json.Marshal(n)
	if err != nil {
		return fmt.Errorf("could not encode message body: %s", err)
	}

	// send message to SQS
	_, err = sqsSvc.SendMessage(&sqs.SendMessageInput{
		QueueUrl:    aws.String(url),
		MessageBody: aws.String(string(data)),
	})

	return err
}
