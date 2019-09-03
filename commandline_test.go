package distconf

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrefixParse(t *testing.T) {
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
