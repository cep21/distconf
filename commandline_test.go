package distconf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrefixParse(t *testing.T) {
	l := CommandLine{
		Prefix: "pre",
		Source: []string{"wha", "prebob=3"},
	}
	b, err := l.Read("bob")
	assert.NoError(t, err)
	assert.Equal(t, []byte("3"), b)

	b, err = l.Read("nothere")
	assert.NoError(t, err)
	assert.Nil(t, b)
}
