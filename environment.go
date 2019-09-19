package distconf

import (
	"context"
	"os"
)

type Environment struct {
}

var _ Reader = &CommandLine{}

func (p *Environment) Read(_ context.Context, key string) ([]byte, error) {
	val := os.Getenv(key)
	if val == "" {
		return nil, nil
	}
	return []byte(val), nil
}
