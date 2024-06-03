package main

type BenchmarkConfig struct {
	Connections map[string]ConnectionConfig `yaml:"connections"`
}

type ConnectionConfig struct {
	Address  string `yaml:"address"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}
