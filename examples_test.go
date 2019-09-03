package distconf_test

import (
	"context"
	"expvar"
	"fmt"

	"github.com/cep21/distconf"
)

func ExampleDistconf() {
	ctx := context.Background()
	m := distconf.Mem{}
	if err := m.Write(ctx, "value", []byte("true")); err != nil {
		panic("never happens")
	}
	d := distconf.Distconf{
		Readers: []distconf.Reader{&m},
	}
	x := d.Bool(ctx, "value", false)
	fmt.Println(x.Get())
	// Output: true
}

func ExampleDistconf_Bool() {
	ctx := context.Background()
	m := distconf.Mem{}
	if err := m.Write(ctx, "value", []byte("true")); err != nil {
		panic("never happens")
	}
	d := distconf.Distconf{
		Readers: []distconf.Reader{&m},
	}
	x := d.Bool(ctx, "value", false)
	fmt.Println(x.Get())
	// Output: true
}

func ExampleDistconf_Float() {
	ctx := context.Background()
	m := distconf.Mem{}
	if err := m.Write(ctx, "value", []byte("3.2")); err != nil {
		panic("never happens")
	}
	d := distconf.Distconf{
		Readers: []distconf.Reader{&m},
	}
	x := d.Float(ctx, "value", 1.0)
	fmt.Println(x.Get())
	// Output: 3.2
}

func ExampleFloat_Get_inloop() {
	ctx := context.Background()
	m := distconf.Mem{}
	if err := m.Write(ctx, "value", []byte("2.0")); err != nil {
		panic("never happens")
	}
	d := distconf.Distconf{
		Readers: []distconf.Reader{&m},
	}
	x := d.Float(ctx, "value", 1.0)
	sum := 0.0
	for i := 0; i < 1000; i++ {
		sum += x.Get()
	}
	fmt.Println(sum)
	// Output: 2000
}

func ExampleDistconf_defaults() {
	ctx := context.Background()
	d := distconf.Distconf{}
	x := d.Float(ctx, "value", 1.1)
	fmt.Println(x.Get())
	// Output: 1.1
}

func ExampleDistconf_Var() {
	d := distconf.Distconf{}
	expvar.Publish("distconf", d.Var())
	// Output:
}

func ExampleFloat_Watch() {
	ctx := context.Background()
	m := distconf.Mem{}
	d := distconf.Distconf{
		Readers: []distconf.Reader{&m},
	}
	x := d.Float(ctx, "value", 1.0)
	x.Watch(func(f *distconf.Float, oldValue float64) {
		fmt.Println("Change from", oldValue, "to", f.Get())
	})
	fmt.Println("first", x.Get())
	if err := m.Write(ctx, "value", []byte("2.1")); err != nil {
		panic("never happens")
	}
	fmt.Println("second", x.Get())
	// Output: first 1
	// Change from 1 to 2.1
	// second 2.1
}
