package distconf

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvConf(t *testing.T) {
	e := &Environment{}
	b, err := e.Read("not_in_env_i_hope_SDFSDFSDFSDFSDF")
	assert.NoError(t, err)
	assert.Nil(t, b)
	assert.NoError(t, os.Setenv("test_TestEnvConf", "abc"))
	b, err = e.Read("test_TestEnvConf")
	assert.NoError(t, err)
	assert.Equal(t, []byte("abc"), b)
}
