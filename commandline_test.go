package distconf

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandLine_Read_prefix(t *testing.T) {
	l := CommandLine{
		Prefix: "pre",
		Source: []string{"wha", "prebob=3"},
	}
	b, err := l.Read(context.Background(), "bob")
	assert.NoError(t, err)
	assert.Equal(t, []byte("3"), b)

	b, err = l.Read(context.Background(), "nothere")
	assert.NoError(t, err)
	assert.Nil(t, b)
}

func TestCommandLine_Read_defaults(t *testing.T) {
	l := CommandLine{
		Prefix: "pre",
	}
	b, err := l.Read(context.Background(), strings.Repeat("s", 1024))
	assert.NoError(t, err)
	var expected []byte
	assert.Equal(t, expected, b)
}
