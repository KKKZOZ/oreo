package config

import "github.com/google/uuid"

type UUIDGenerator struct {
}

func NewUUIDGenerator() *UUIDGenerator {
	return &UUIDGenerator{}
}

func (u *UUIDGenerator) GenerateId() string {
	return uuid.NewString()
}
