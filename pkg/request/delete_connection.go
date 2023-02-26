package request

import "encoding/json"

type DeleteConnection struct {
	ConnectionId string `json:"connectionId"`
}

// UserFromString decodes json to UserNotification
func DeleteConnectionFromString(request string) (DeleteConnection, error) {
	u := DeleteConnection{}
	err := json.Unmarshal([]byte(request), &u)
	return u, err
}
