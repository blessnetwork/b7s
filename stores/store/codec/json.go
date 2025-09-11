package codec

import (
	"encoding/json"
)

type JSON struct{}

func NewJSONCodec() JSON {
	return JSON{}
}

func (c JSON) Marshal(obj any) ([]byte, error) {
	return json.Marshal(obj)
}

func (c JSON) Unmarshal(data []byte, obj any) error {
	return json.Unmarshal(data, obj)
}
