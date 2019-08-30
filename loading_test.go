package distconf

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFromLoaders(t *testing.T) {
	c := FromLoaders([]BackingLoader{MemLoader()}, nil)
	assert.Equal(t, 1, len(c.readers))
}

func TestFromLoadersWithErrors(t *testing.T) {
	counter := 0
	c := FromLoaders([]BackingLoader{BackingLoaderFunc(func() (Reader, error) {
		return nil, errors.New("nope")
	})}, func(err error, loader BackingLoader) {
		counter++
	})
	assert.Equal(t, 0, len(c.readers))
	assert.Equal(t, 1, counter)
}
