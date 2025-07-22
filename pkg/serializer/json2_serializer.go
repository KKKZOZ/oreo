package serializer

import jsoniter "github.com/json-iterator/go"

var json2 = jsoniter.ConfigCompatibleWithStandardLibrary

type JSON2Serializer struct{}

func NewJSON2Serializer() *JSON2Serializer {
	return &JSON2Serializer{}
}

func (s *JSON2Serializer) Serialize(data any) ([]byte, error) {
	return json2.Marshal(data)
}

func (s *JSON2Serializer) Deserialize(bs []byte, tar any) error {
	return json2.Unmarshal(bs, tar)
}
