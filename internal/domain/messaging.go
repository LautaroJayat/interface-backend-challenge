package domain

import (
	"fmt"
	"time"
)

const (
	// Topic prefixes for message broadcasting
	MessageTopicPrefix = "messages"
	StatusTopicPrefix  = "status"
)

type MessageType string

const (
	MessageTypeNewMessage   MessageType = "new_message"
	MessageTypeStatusUpdate MessageType = "status_update"
)

type StatusType string

const (
	StatusOnline  StatusType = "online"
	StatusOffline StatusType = "offline"
	StatusTyping  StatusType = "typing"
)


type MessageEnvelope struct {
	Type      MessageType `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      Message     `json:"data"`
}

type StatusUpdateEnvelope struct {
	Type      MessageType `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

func GetMessageTopic(receiverID string) string {
	return fmt.Sprintf("%s.%s", MessageTopicPrefix, receiverID)
}

func GetStatusTopic(userID string) string {
	return fmt.Sprintf("%s.%s", StatusTopicPrefix, userID)
}