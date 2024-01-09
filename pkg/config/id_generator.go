package config

type IdGenerator interface {
	GenerateId() string
}
