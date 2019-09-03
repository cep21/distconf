# distconf
[![CircleCI](https://circleci.com/gh/cep21/distconf.svg)](https://circleci.com/gh/cep21/distconf)
[![GoDoc](https://godoc.org/github.com/cep21/distconf?status.svg)](https://godoc.org/github.com/cep21/distconf)
[![codecov](https://codecov.io/gh/cep21/distconf/branch/master/graph/badge.svg)](https://codecov.io/gh/cep21/distconf)

distconf is a distributed configuration framework for Go.

All applications need to load configuration somehow.  Configuration can be loaded
in many different ways

* Environment variables
* Command line parameters
* ZooKeeper or consul

How configuration is loaded should ideally be abstracted from the need for configuration.

An additional complication is that configuration can change while an application is live.  It is sometimes
useful to allow applications to update their configuration without having to restart.  Unfortunately,
systems like zookeeper can be slow so your application also needs to atomically cache configuration, while
also monitoring for changes.

Distconf does all that

* Abstract the need for configuration from the source of configuration
* Fast, atomic loading of configuration
* Monitoring of configuration updates

# Usage

## Getting a float value from distconf

```go
    func ExampleDistconf_Float() {
        m := distconf.Mem{}
        if err := m.Write("value", []byte("3.2")); err != nil {
            panic("never happens")
        }
        d := distconf.Distconf{
            Readers: []distconf.Reader{&m},
        }
        x := d.Float("value", 1.0)
        fmt.Println(x.Get())
        // Output: 3.2
    }
```

## Getting the default value from distconf

```go
    func ExampleDistconf_defaults() {
        d := distconf.Distconf{}
        x := d.Float("value", 1.1)
        fmt.Println(x.Get())
        // Output: 1.1
    }
```

## Watching for updates for values

```go
    func ExampleFloat_Watch() {
        m := distconf.Mem{}
        d := distconf.Distconf{
            Readers: []distconf.Reader{&m},
        }
        x := d.Float("value", 1.0)
        x.Watch(func(f *distconf.Float, oldValue float64) {
            fmt.Println("Change from", oldValue, "to", f.Get())
        })
        fmt.Println("first", x.Get())
        if err := m.Write("value", []byte("2.1")); err != nil {
            panic("never happens")
        }
        fmt.Println("second", x.Get())
        // Output: first 1
        // Change from 1 to 2.1
        // second 2.1
    }
```

# Design Rational

The core component of distconf is an interface with only one method.

```go
    // Reader can get a []byte value for a config key
    type Reader interface {
        // Read should lookup a key inside the configuration source.  This function should
        // be thread safe, but is allowed to be slow or block.  That block will only happen
        // on application startup.  An error will skip this source and fall back to another
        // source in the chain.
        Read(key string) ([]byte, error)
    }
```

From this simple interface, we can derive the rest of distconf.  Readers are used by Distconf to fetch configuration
information.

Some readers, such as zookeeper, allow you to watch for specific keys and get notifications live when they change.  To
support this, readers can also optionally implement the Watcher interface.

```go
    // A Watcher config can change what it thinks a value is over time.
    type Watcher interface {
        // Watch a key for a change in value.  When the value for that key changes,
        // execute 'callback'.  It is ok to execute callback more times than needed.
        // Each call to callback will probably trigger future calls to Get().
        // Watch may be called multiple times for a single key.  Only the latest callback needs to be executed.
        // It is possible callback may itself call watch.  Be careful with locking.
        // If callback is nil, then we are trying to remove a previously registered callback.
        Watch(key string, callback func()) error
    }
```

Any Reader that implements Watcher will get a Watch() function call whenever a key should be watched for changers.  That
Watcher should then forever notify Distconf whenever that key's value changes by executing callback.  When you're done
with distconf, call `Shutdown` to deregister watches.

# Contributing

Contributions welcome!  Submit a pull request on github and make sure your code passes `make lint test`.  For
large changes, I strongly recommend [creating an issue](https://github.com/cep21/distconf/issues) on GitHub first to
confirm your change will be accepted before writing a lot of code.  GitHub issues are also recommended, at your discretion,
for smaller changes or questions.

# License

This library is licensed under the Apache 2.0 License, forked from https://github.com/signalfx/golib
under the Apache 2.0 License.