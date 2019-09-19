package distconf

import (
	"context"
	"expvar"
	"math"
	"runtime"
	"sync"
	"time"
)

// Hooks are optional callbacks that let you get information about the internal workings and errors of distconf.
// All functions of Hooks should be thread safe.
type Hooks struct {
	// OnError is called whenever there is an error doing something internally to distconf, but the error itself
	// cannot be directly returned to the caller.
	// distconfKey is the key that caused the error.
	OnError func(msg string, distconfKey string, err error)
}

func (h Hooks) onError(msg string, distconfKey string, err error) {
	if h.OnError != nil {
		h.OnError(msg, distconfKey, err)
	}
}

// Distconf gets configuration data from the first backing that has it.  It is a race condition to modify Hooks
// or Readers after you've started using Distconf.  All public functions of Distconf are thread safe.
type Distconf struct {
	// Hooks are optional callbacks that let you get information about the internal workings and errors of distconf.
	Hooks Hooks
	// Readers are an ordered list of backends for DistConf to get configuration information from.  The list
	// order is important as information from a first backend will be returned before the later ones.
	Readers []Reader
	// How long to timeout out of band refresh calls triggered by Watch() callbacks.  Defaults to 1 second.
	RefreshTimeout         time.Duration
	varsMutex              sync.Mutex
	infoMutex              sync.RWMutex
	registeredWatchesMutex sync.Mutex
	registeredVars         map[string]*registeredVariableTracker
	distInfos              map[string]distInfo
	registeredWatches      map[string][]Watcher
	callerFunc             func(int) (uintptr, string, int, bool)
}

type registeredVariableTracker struct {
	distvar        configVariable
	hasInitialized sync.Once
}

type configVariable interface {
	Update(newValue []byte) error
	// Get but on an interface return.  Oh how I miss you templates.
	GenericGet() interface{}
	GenericGetDefault() interface{}
	Type() distType
}

type distType int

const (
	// StrType is type Str
	strType distType = iota
	// BoolType is type Bool
	boolType
	// FloatType is type Float
	floatType
	// DurationType is type Duration
	durationType
	// IntType is type Int
	intType
)

// distInfo is useful to unmarshal/marshal the Info expvar
type distInfo struct {
	File         string      `json:"file"`
	Line         int         `json:"line"`
	DefaultValue interface{} `json:"default_value"`
	DistType     distType    `json:"dist_type"`
}

func (c *Distconf) grabInfo(key string) {
	if c.callerFunc == nil {
		c.callerFunc = runtime.Caller
	}
	_, file, line, ok := c.callerFunc(2)
	if !ok {
		c.Hooks.onError("unable to find call for distconf", key, nil)
	}
	info := distInfo{
		File: file,
		Line: line,
	}
	c.infoMutex.Lock()
	defer c.infoMutex.Unlock()
	if c.distInfos == nil {
		c.distInfos = make(map[string]distInfo)
	}
	c.distInfos[key] = info
}

// Var returns an expvar variable that shows all the current configuration variables and their
// current value
func (c *Distconf) Var() expvar.Var {
	return expvar.Func(func() interface{} {
		c.varsMutex.Lock()
		defer c.varsMutex.Unlock()

		m := make(map[string]interface{}, len(c.registeredVars))
		for name, v := range c.registeredVars {
			m[name] = v.distvar.GenericGet()
		}
		return m
	})
}

// Info returns an expvar variable that shows the information for all configuration variables.
// Information consist of file, line, default value and type of variable.
func (c *Distconf) Info() expvar.Var {
	return expvar.Func(func() interface{} {
		c.infoMutex.RLock()
		defer c.infoMutex.RUnlock()

		m := make(map[string]distInfo, len(c.distInfos))
		for k, i := range c.distInfos {
			v, ok := c.registeredVars[k]
			if ok {
				v := distInfo{
					File:         i.File,
					Line:         i.Line,
					DefaultValue: v.distvar.GenericGetDefault(),
					DistType:     v.distvar.Type(),
				}
				m[k] = v
			}
		}
		return m
	})
}

// Int object that can be referenced to get integer values from a backing config.
func (c *Distconf) Int(ctx context.Context, key string, defaultVal int64) *Int {
	c.grabInfo(key)
	s := &intConf{
		defaultVal: defaultVal,
		Int: Int{
			currentVal: defaultVal,
		},
	}
	// Note: in race conditions 's' may not be the thing actually returned
	ret, okCast := c.createOrGet(ctx, key, s).(*intConf)
	if !okCast {
		c.Hooks.onError("Registering key with multiple types!  FIX ME!!!!", key, nil)
		return nil
	}
	return &ret.Int
}

// Float object that can be referenced to get float values from a backing config
func (c *Distconf) Float(ctx context.Context, key string, defaultVal float64) *Float {
	c.grabInfo(key)
	s := &floatConf{
		defaultVal: defaultVal,
		Float: Float{
			currentVal: math.Float64bits(defaultVal),
		},
	}
	// Note: in race conditions 's' may not be the thing actually returned
	ret, okCast := c.createOrGet(ctx, key, s).(*floatConf)
	if !okCast {
		c.Hooks.onError("Registering key with multiple types!  FIX ME!!!!", key, nil)
		return nil
	}
	return &ret.Float
}

// Str object that can be referenced to get string values from a backing config
func (c *Distconf) Str(ctx context.Context, key string, defaultVal string) *Str {
	c.grabInfo(key)
	s := &strConf{
		defaultVal: defaultVal,
	}
	s.currentVal.Store(defaultVal)
	// Note: in race conditions 's' may not be the thing actually returned
	ret, okCast := c.createOrGet(ctx, key, s).(*strConf)
	if !okCast {
		c.Hooks.onError("Registering key with multiple types!  FIX ME!!!!", key, nil)
		return nil
	}
	return &ret.Str
}

// Bool object that can be referenced to get boolean values from a backing config
func (c *Distconf) Bool(ctx context.Context, key string, defaultVal bool) *Bool {
	c.grabInfo(key)
	var defautlAsInt int32
	if defaultVal {
		defautlAsInt = 1
	} else {
		defautlAsInt = 0
	}

	s := &boolConf{
		defaultVal: defautlAsInt,
		Bool: Bool{
			currentVal: defautlAsInt,
		},
	}
	// Note: in race conditions 's' may not be the thing actually returned
	ret, okCast := c.createOrGet(ctx, key, s).(*boolConf)
	if !okCast {
		c.Hooks.onError("Registering key with multiple types!  FIX ME!!!!", key, nil)
		return nil
	}
	return &ret.Bool
}

// Duration returns a duration object that calls ParseDuration() on the given key
func (c *Distconf) Duration(ctx context.Context, key string, defaultVal time.Duration) *Duration {
	c.grabInfo(key)
	s := &durationConf{
		defaultVal: defaultVal,
		Duration: Duration{
			currentVal: defaultVal.Nanoseconds(),
		},
		hooks:       c.Hooks,
		originalKey: key,
	}
	// Note: in race conditions 's' may not be the thing actually returned
	ret, okCast := c.createOrGet(ctx, key, s).(*durationConf)
	if !okCast {
		c.Hooks.onError("Registering key with multiple types!  FIX ME!!!!", key, nil)
		return nil
	}
	return &ret.Duration
}

// Shutdown this config framework's Readers.  Config variable results are undefined after this call.
// Returns the error of the first reader to return an error.  While Watch itself doesn't take a context, shutdown
// will stop calling future watches if context ends.
func (c *Distconf) Shutdown(ctx context.Context) error {
	c.varsMutex.Lock()
	c.registeredWatchesMutex.Lock()
	defer c.varsMutex.Unlock()
	defer c.registeredWatchesMutex.Unlock()
	var ret error
	for key, watches := range c.registeredWatches {
		for _, watch := range watches {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			if err := watch.Watch(ctx, key, nil); err != nil {
				c.Hooks.onError("error unregistering watch", key, err)
				ret = err
			}
		}
	}
	return ret
}

func (c *Distconf) watchCallback(key string) func() {
	return func() {
		timeout := c.RefreshTimeout
		if timeout == 0 {
			timeout = time.Second
		}
		refreshCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		c.Refresh(refreshCtx, key)
	}
}

func (c *Distconf) registerWatches(ctx context.Context, key string, watches []Watcher) {
	c.registeredWatchesMutex.Lock()
	if c.registeredWatches == nil {
		c.registeredWatches = make(map[string][]Watcher)
	}
	c.registeredWatches[key] = append(c.registeredWatches[key], watches...)
	// Unlock early so we don't get in deadlock if backing.Watch() somehow executes code that gets back here
	c.registeredWatchesMutex.Unlock()
	for _, backing := range watches {
		err := backing.Watch(ctx, key, c.watchCallback(key))
		if err != nil {
			c.Hooks.onError("Unable to watch for config var", key, err)
		}
	}
}

func (c *Distconf) refresh(ctx context.Context, key string, configVar configVariable) {
	var dynamicReadersOnPath []Watcher
	defer func() {
		c.registerWatches(ctx, key, dynamicReadersOnPath)
	}()
	for _, backing := range c.Readers {
		if asW, ok := backing.(Watcher); ok {
			dynamicReadersOnPath = append(dynamicReadersOnPath, asW)
		}

		v, e := backing.Read(ctx, key)
		if e != nil {
			c.Hooks.onError("Unable to read from backing", key, e)
			continue
		}
		if v != nil {
			e = configVar.Update(v)
			if e != nil {
				c.Hooks.onError("Invalid config bytes", key, e)
			}
			return
		}
	}

	// None of the readers have this value.  Update it to nil (default).
	e := configVar.Update(nil)
	if e != nil {
		c.Hooks.onError("Unable to set bytes to nil/clear", key, e)
	}
}

func (c *Distconf) createOrGet(ctx context.Context, key string, defaultVar configVariable) configVariable {
	c.varsMutex.Lock()
	rv, exists := c.registeredVars[key]
	if !exists {
		rv = &registeredVariableTracker{
			distvar: defaultVar,
		}
		if c.registeredVars == nil {
			c.registeredVars = make(map[string]*registeredVariableTracker)
		}
		c.registeredVars[key] = rv
	}
	c.varsMutex.Unlock()

	rv.hasInitialized.Do(func() {
		c.refresh(ctx, key, rv.distvar)
	})
	return rv.distvar
}

// Refresh a single key from readers.  This will force a blocking Read from the Readers, in order, until one of them
// has the key.  It will then update the distconf value for that key and trigger any update callbacks.  You do not
// generally need to call this.  If your backends implement Watcher, they will trigger this for you.
func (c *Distconf) Refresh(ctx context.Context, key string) {
	c.varsMutex.Lock()
	m, exists := c.registeredVars[key]
	c.varsMutex.Unlock()
	if !exists {
		c.Hooks.onError("Backing callback on variable that doesn't exist", key, nil)
		return
	}
	c.refresh(ctx, key, m.distvar)
}

// Reader can get a []byte value for a config key
type Reader interface {
	// Read should lookup a key inside the configuration source.  This function should
	// be thread safe, but is allowed to be slow or block.  That block will only happen
	// on application startup.  An error will skip this source and fall back to another
	// source in the chain.  If the value is not inside this reader, return nil, nil.
	Read(ctx context.Context, key string) ([]byte, error)
}

// Shutdownable is an optional interface of Reader that allows it to be gracefully shutdown.
type Shutdownable interface {
	// Shutdown should signal to a reader it is no longer needed by Distconf. It should expect
	// to no longer require to return more recent values to distconf.
	Shutdown(ctx context.Context) error
}

// A Watcher config can change what it thinks a value is over time.
type Watcher interface {
	// Watch a key for a change in value.  When the value for that key changes,
	// execute 'callback'.  It is ok to execute callback more times than needed.
	// Each call to callback will probably trigger future calls to Get().
	// Watch may be called multiple times for a single key.  Only the latest callback needs to be executed.
	// It is possible callback may itself call watch.  Be careful with locking.
	// If callback is nil, then we are trying to remove a previously registered callback.
	Watch(ctx context.Context, key string, callback func()) error
}
