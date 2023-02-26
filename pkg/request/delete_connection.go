package request

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type DeleteConnection struct {
	ConnectionId string `json:"connectionId"`
}

// DeleteConnectionFromString decodes json to DeleteConnection
func DeleteConnectionFromString(request string) (DeleteConnection, error) {
	u := DeleteConnection{}
	err := json.Unmarshal([]byte(request), &u)
	return u, err
}

// DeleteConnectionFromId decodes creates request just from the id
func DeleteConnectionFromId(id string) DeleteConnection {
	return DeleteConnection{
		ConnectionId: id,
	}
}

func (n DeleteConnection) DeleteDelayedSQS(sqsSvc *sqs.SQS, url string) error {
	// serialize ConnectionNotification
	data, err := json.Marshal(n)
	if err != nil {
		return fmt.Errorf("could not encode message body: %s", err)
	}

	// send message to SQS
	_, err = sqsSvc.SendMessage(&sqs.SendMessageInput{
		QueueUrl:     aws.String(url),
		MessageBody:  aws.String(string(data)),
		DelaySeconds: aws.Int64(900),
	})

	return err
}
