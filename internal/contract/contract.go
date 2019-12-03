package contract

import (
	"encoding/json"
	"time"
)

const (
	ElasticNewsIndexType = "news"
)

type (
	NewsOnElastic struct {
		ID      int64     `json:"id,omitempty"`
		Created time.Time `json:"created,omitempty"`
	}

	NewsData struct {
		ID      int64     `json:"id,omitempty"`
		Author  string    `json:"author,omitempty"`
		Body    string    `json:"body,omitempty"`
		Created time.Time `json:"created,omitempty"`
	}

	HTTPResponse struct {
		Error   bool        `json:"error"`
		Message string      `json:"message"`
		Meta    interface{} `json:"meta"`
	}
)

func (n NewsData) Marshal() ([]byte, error) {
	return json.Marshal(n)
}

func (n *NewsData) UnMarshal(incoming []byte) error {
	return json.Unmarshal(incoming, &n)
}

func (n NewsData) ToElasticData() NewsOnElastic {
	return NewsOnElastic{
		ID:      n.ID,
		Created: n.Created,
	}
}
