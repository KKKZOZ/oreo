package serializer

type Serializer interface {
	Serialize(data any) ([]byte, error)
	Deserialize(bs []byte, tar any) error
}
