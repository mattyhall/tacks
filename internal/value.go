package internal

import (
	"strconv"
	"time"

	"github.com/couchbase/gocb/v2"
)

type Stretch struct {
	ID          string            `json:"id"`
	Description string            `json:"description"`
	Start       time.Time         `json:"start"`
	End         *time.Time        `json:"end"`
	Tags        []string          `json:"tags"`
	Attributes  map[string]string `json:"attrs"`
}

func GetID(collection *gocb.Collection) (string, error) {
	id, err := collection.Binary().Increment("next-id", &gocb.IncrementOptions{Delta: 1})
	if err != nil {
		return "", err
	}

	return strconv.FormatUint(id.Content(), 10), nil
}
