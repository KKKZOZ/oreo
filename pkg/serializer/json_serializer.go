package serializer

import "encoding/json"

type JSONSerializer struct {
}

func NewJSONSerializer() *JSONSerializer {
	return &JSONSerializer{}
}

func (s *JSONSerializer) Serialize(data any) ([]byte, error) {
	return json.Marshal(data)
}

func (s *JSONSerializer) Deserialize(bs []byte, tar any) error {
	return json.Unmarshal(bs, tar)
}
