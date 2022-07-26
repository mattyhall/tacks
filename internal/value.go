package internal

import (
	"time"

	"github.com/couchbase/gocb/v2"
)

type Stretch struct {
	ID          uint64            `json:"id"`
	Description string            `json:"description"`
	Start       time.Time         `json:"start"`
	End         *time.Time        `json:"end"`
	Tags        []string          `json:"tags"`
	Attributes  map[string]string `json:"attrs"`
}

func GetID(collection *gocb.Collection) (uint64, error) {
	id, err := collection.Binary().Increment("next-id", &gocb.IncrementOptions{Delta: 1})
	if err != nil {
		return 0, err
	}

	return id.Content(), nil
}
