package distconf

import "os"

type Environment struct {
	osGetenv func(key string) string
}

func (p *Environment) Read(key string) ([]byte, error) {
	val := os.Getenv(key)
	if val == "" {
		return nil, nil
	}
	return []byte(val), nil
}
