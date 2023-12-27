package serializer

import (
	"bytes"
	"encoding/gob"
)

type GobSerializer struct {
}

func NewGobSerializer() *GobSerializer {
	return &GobSerializer{}
}

func (s *GobSerializer) Serialize(data any) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(data)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (s *GobSerializer) Deserialize(bs []byte, tar any) error {
	buffer := bytes.NewBuffer(bs)
	decoder := gob.NewDecoder(buffer)
	return decoder.Decode(tar)
}
