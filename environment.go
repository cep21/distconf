package distconf

import "os"

type Environment struct {
	osGetenv func(key string) string
}

var _ Reader = &CommandLine{}

func (p *Environment) Read(key string) ([]byte, error) {
	val := os.Getenv(key)
	if val == "" {
		return nil, nil
	}
	return []byte(val), nil
}
