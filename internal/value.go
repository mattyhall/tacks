package internal

import (
	"time"
)

type Stretch struct {
	ID          string            `json:"id"`
	Description string            `json:"description"`
	Start       time.Time         `json:"start"`
	End         *time.Time        `json:"end"`
	Tags        []string          `json:"tags"`
	Attributes  map[string]string `json:"attrs"`
}
