package distconf

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type allErrorBacking struct {
}

var errNope = errors.New("nope")

var _ ReaderWriterWatcher = &allErrorBacking{}

func (m *allErrorBacking) Read(ctx context.Context, key string) ([]byte, error) {
	return nil, errNope
}

func (m *allErrorBacking) Write(ctx context.Context, key string, value []byte) error {
	return errNope
}

func (m *allErrorBacking) Watch(ctx context.Context, key string, callback func()) error {
	return errNope
}

type allErrorconfigVariable struct {
}

func (a *allErrorconfigVariable) Update(newValue []byte) error {
	return errNope
}
func (a *allErrorconfigVariable) GenericGet() interface{} {
	return errNope
}
func (a *allErrorconfigVariable) GenericGetDefault() interface{} {
	return errNope
}
func (a *allErrorconfigVariable) Type() distType {
	return intType
}

type ReaderWriterWatcher interface {
	Reader
	Watcher
	Write(ctx context.Context, key string, value []byte) error
}

func makeConf() (ReaderWriterWatcher, *Distconf) {
	memConf := &Mem{}
	conf := &Distconf{
		Readers: []Reader{memConf},
	}
	return memConf, conf
}

func mustShutdown(t *testing.T, s Shutdownable) {
	require.NoError(t, s.Shutdown(context.Background()))
}

func TestDistconf_Int(t *testing.T) {
	memConf, conf := makeConf()
	defer mustShutdown(t, conf)

	// default
	val := conf.Int(context.Background(), "testval", 1)
	assert.Equal(t, int64(1), val.Get())
	totalWatches := 0
	val.Watch(IntWatch(func(str *Int, oldValue int64) {
		totalWatches++
	}))

	// update valid
	require.NoError(t, memConf.Write(context.Background(), "testval", []byte("2")))
	assert.Equal(t, int64(2), val.Get())

	// check already registered
	conf.Str(context.Background(), "testval_other", "moo")
	var nilInt *Int
	assert.Equal(t, nilInt, conf.Int(context.Background(), "testval_other", 0))

	// update to invalid
	require.NoError(t, memConf.Write(context.Background(), "testval", []byte("invalidint")))
	assert.Equal(t, int64(2), val.Get())

	// update to nil
	require.NoError(t, memConf.Write(context.Background(), "testval", nil))
	assert.Equal(t, int64(1), val.Get())

	// check callback
	assert.Equal(t, 2, totalWatches)
	assert.Contains(t, conf.Var().String(), "testval")
}

func TestDistconf_Float(t *testing.T) {
	memConf, conf := makeConf()
	defer mustShutdown(t, conf)

	// default
	val := conf.Float(context.Background(), "testval", 3.14)
	assert.Equal(t, float64(3.14), val.Get())
	totalWatches := 0
	val.Watch(FloatWatch(func(float *Float, oldValue float64) {
		totalWatches++
	}))

	// update to valid
	require.NoError(t, memConf.Write(context.Background(), "testval", []byte("4.771")))
	assert.Equal(t, float64(4.771), val.Get())

	// check already registered
	conf.Str(context.Background(), "testval_other", "moo")
	var nilFloat *Float
	assert.Equal(t, nilFloat, conf.Float(context.Background(), "testval_other", 0.0))

	// update to invalid
	require.NoError(t, memConf.Write(context.Background(), "testval", []byte("invalidfloat")))
	assert.Equal(t, float64(4.771), val.Get())

	// update to nil
	require.NoError(t, memConf.Write(context.Background(), "testval", nil))
	assert.Equal(t, float64(3.14), val.Get())

	// check callback
	assert.Equal(t, 2, totalWatches)
	assert.Contains(t, conf.Var().String(), "testval")
}

func TestDistconf_Str(t *testing.T) {
	memConf, conf := makeConf()
	defer mustShutdown(t, conf)

	// default
	val := conf.Str(context.Background(), "testval", "default")
	assert.Equal(t, "default", val.Get())
	totalWatches := 0
	val.Watch(StrWatch(func(str *Str, oldValue string) {
		totalWatches++
	}))

	// update to valid
	require.NoError(t, memConf.Write(context.Background(), "testval", []byte("newval")))
	assert.Equal(t, "newval", val.Get())

	// check already registered
	conf.Int(context.Background(), "testval_other", 0)
	var nilStr *Str
	assert.Equal(t, nilStr, conf.Str(context.Background(), "testval_other", ""))

	// update to nil
	require.NoError(t, memConf.Write(context.Background(), "testval", nil))
	assert.Equal(t, "default", val.Get())

	// check callback
	assert.Equal(t, 2, totalWatches)
	assert.Contains(t, conf.Var().String(), "testval_other")

}

func TestDistconf_Duration(t *testing.T) {
	ctx := context.Background()
	memConf, conf := makeConf()
	defer mustShutdown(t, conf)

	//default

	val := conf.Duration(ctx, "testval", time.Second)
	assert.Equal(t, time.Second, val.Get())
	totalWatches := 0
	val.Watch(DurationWatch(func(*Duration, time.Duration) {
		totalWatches++
	}))

	// update valid
	require.NoError(t, memConf.Write(ctx, "testval", []byte("10ms")))
	assert.Equal(t, time.Millisecond*10, val.Get())

	// check already registered
	conf.Str(ctx, "testval_other", "moo")
	var nilDuration *Duration
	assert.Equal(t, nilDuration, conf.Duration(ctx, "testval_other", 0))

	// update to invalid
	require.NoError(t, memConf.Write(ctx, "testval", []byte("abcd")))
	assert.Equal(t, time.Second, val.Get())

	// update to nil
	require.NoError(t, memConf.Write(ctx, "testval", nil))
	assert.Equal(t, time.Second, val.Get())

	assert.Equal(t, 2, totalWatches)
	assert.Contains(t, conf.Var().String(), "testval")
}

func TestDistconf_Bool(t *testing.T) {
	ctx := context.Background()
	memConf, conf := makeConf()
	defer mustShutdown(t, conf)

	//default

	val := conf.Bool(ctx, "testval", false)
	assert.False(t, val.Get())
	totalWatches := 0
	val.Watch(BoolWatch(func(*Bool, bool) {
		totalWatches++
	}))

	// update valid
	require.NoError(t, memConf.Write(ctx, "testval", []byte("true")))
	assert.True(t, val.Get())

	// update valid
	require.NoError(t, memConf.Write(ctx, "testval", []byte("FALSE")))
	assert.False(t, val.Get())

	// check already registered
	conf.Str(ctx, "testval_other", "moo")
	var nilBool *Bool
	assert.Equal(t, nilBool, conf.Bool(ctx, "testval_other", true))

	// update to invalid
	require.NoError(t, memConf.Write(ctx, "testval", []byte("__")))
	assert.False(t, val.Get())

	// update to nil
	require.NoError(t, memConf.Write(ctx, "testval", nil))
	assert.False(t, val.Get())

	assert.Equal(t, 2, totalWatches)
	assert.Contains(t, conf.Var().String(), "testval")
}

func TestDistconf_errors(t *testing.T) {
	ctx := context.Background()
	conf := &Distconf{
		Readers: []Reader{&allErrorBacking{}},
	}

	iVal := conf.Int(ctx, "testval", 1)
	assert.Equal(t, int64(1), iVal.Get())

	assert.NotPanics(t, func() {
		conf.Refresh(ctx, "not_in_map")
	})

	assert.NotPanics(t, func() {
		conf.Refresh(ctx, "testval2")
	})
}

func testInfo(t *testing.T, dat map[string]distInfo, key string, val interface{}, dtype distType) {
	v, ok := dat[key]
	assert.True(t, ok)
	assert.NotEqual(t, v.Line, 0)
	assert.NotEqual(t, v.File, "")
	assert.Equal(t, v.DistType, dtype)
	assert.Equal(t, v.DefaultValue, val)
}

func TestHooks(t *testing.T) {
	c := 0
	h := Hooks{OnError: func(_ string, _ string, _ error) {
		c++
	}}
	h.onError("", "", nil)
	assert.Equal(t, 1, c)
}

func TestDistconf_Info(t *testing.T) {
	ctx := context.Background()
	_, conf := makeConf()
	defer mustShutdown(t, conf)

	conf.Bool(ctx, "testbool", true)
	conf.Str(ctx, "teststr", "123")
	conf.Int(ctx, "testint", int64(12))
	conf.Duration(ctx, "testdur", time.Millisecond)
	conf.Float(ctx, "testfloat", float64(1.2))

	x := conf.Info()
	assert.NotNil(t, x)
	assert.NotNil(t, x.String())
	var dat map[string]distInfo
	err := json.Unmarshal([]byte(x.String()), &dat)
	assert.NoError(t, err)
	assert.Equal(t, len(dat), 5)
	testInfo(t, dat, "testbool", float64(1), boolType)
	testInfo(t, dat, "teststr", "123", strType)
	testInfo(t, dat, "testint", float64(12), intType)
	testInfo(t, dat, "testdur", time.Millisecond.String(), durationType)
	testInfo(t, dat, "testfloat", float64(1.2), floatType)

	_, conf = makeConf()
	c := int64(0)
	conf.Hooks.OnError = func(msg string, distconfKey string, err error) {
		c++
	}
	conf.callerFunc = func(n int) (uintptr, string, int, bool) {
		return 0, "", 0, false
	}
	assert.Equal(t, c, int64(0))
	conf.Bool(ctx, "testbool", true)
	assert.Equal(t, c, int64(1))
}
