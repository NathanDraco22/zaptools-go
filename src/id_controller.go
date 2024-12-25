package src

import (
	"fmt"

	"github.com/google/uuid"
)


func GenerateID() string {
	prefix := "zpt"
	connId := fmt.Sprintf("%s-%s", prefix, uuid.New().String())
	return connId
}