package distconf_test

import (
	"expvar"
	"fmt"
	"github.com/cep21/distconf"
)

func ExampleDistconf() {
	m := distconf.Mem{}
	if err := m.Write("value", []byte("true")); err != nil {
		panic("never happens")
	}
	d := distconf.Distconf{
		Readers: []distconf.Reader{&m},
	}
	x := d.Bool("value", false)
	fmt.Println(x.Get())
	// Output: true
}

func ExampleDistconf_Bool() {
	m := distconf.Mem{}
	if err := m.Write("value", []byte("true")); err != nil {
		panic("never happens")
	}
	d := distconf.Distconf{
		Readers: []distconf.Reader{&m},
	}
	x := d.Bool("value", false)
	fmt.Println(x.Get())
	// Output: true
}

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

func ExampleDistconf_defaults() {
	d := distconf.Distconf{
		Readers: []distconf.Reader{&distconf.Mem{}},
	}
	x := d.Float("value", 1.1)
	fmt.Println(x.Get())
	// Output: 1.1
}

func ExampleDistconf_Var() {
	d := distconf.Distconf{}
	expvar.Publish("distconf", d.Var())
	// Output:
}

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
