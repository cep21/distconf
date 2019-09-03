package distconf

import (
	"os"
	"strings"
)

type CommandLine struct {
	Prefix string
	Source []string
}

var _ Reader = &CommandLine{}

func (p *CommandLine) Read(key string) ([]byte, error) {
	if p.Source == nil {
		p.Source = os.Args
	}
	argPrefix := p.Prefix + key + "="
	for _, arg := range p.Source {
		if !strings.HasPrefix(arg, argPrefix) {
			continue
		}
		argSuffix := arg[len(argPrefix):]
		return []byte(argSuffix), nil
	}
	return nil, nil
}
