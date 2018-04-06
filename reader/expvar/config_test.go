// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvar_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/arsham/expipe/datatype"
	"github.com/arsham/expipe/reader/expvar"
	"github.com/arsham/expipe/tools"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

func TestWithLogger(t *testing.T) {
	l := (tools.FieldLogger)(nil)
	c := new(expvar.Config)
	err := expvar.WithLogger(l)(c)
	if err == nil {
		t.Error("err = (nil); want (error)")
	}
	l = tools.DiscardLogger()
	err = expvar.WithLogger(l)(c)
	if err != nil {
		t.Errorf("err = (%v); want (nil)", err)
	}
	if c.Logger() != l {
		t.Errorf("c.Logger() = (%v); want (%v)", c.Logger(), l)
	}
}

type unmarshaller interface {
	UnmarshalKey(key string, rawVal interface{}) error
	AllKeys() []string
}

func TestWithViper(t *testing.T) {
	v := viper.New()
	v.SetConfigType("yaml")
	c := new(expvar.Config)
	input := `
    recorders:
        recorder1:
            endpoint: http://127.0.0.1:9200
            type_name: %s
            map_file: noway
            timeout: 10s
            interval: 1s
    `

	in := bytes.NewBufferString(fmt.Sprintf(input, ""))
	v.ReadConfig(in)
	err := expvar.WithViper(v, "name", "recorders.recorder1")(c)
	if err == nil {
		t.Error("err = (nil); want (error): empty typeName")
	}

	in = bytes.NewBufferString(fmt.Sprintf(input, ""))
	v.ReadConfig(in)
	err = expvar.WithViper(v, "name", "")(c)
	if err == nil {
		t.Error("err = (nil); want (error): empty key")
	}

	err = expvar.WithViper(nil, "name", "recorders.recorder1")(c)
	if err == nil {
		t.Error("err = (nil); want (error): nil viper")
	}
}

func TestWithViperSuccess(t *testing.T) {
	v := viper.New()
	v.SetConfigType("yaml")

	input := bytes.NewBuffer([]byte(`
    recorders:
        recorder1:
            endpoint: http://127.0.0.1:9200
            type_name: example_type
            map_file: noway
            timeout: 10s
            interval: 1s
    `))
	v.ReadConfig(input)
	c := new(expvar.Config)
	err := expvar.WithViper(v, "recorder1", "recorders.recorder1")(c)
	if err != nil {
		t.Fatalf("err = (%v); want (nil)", err)
	}
	if c.Timeout() != 10*time.Second {
		t.Errorf("c.Timeout() = (%d); want (%d)", c.Timeout(), 10*time.Second)
	}
	if c.Endpoint() != "http://127.0.0.1:9200" {
		t.Errorf("c.Endpoint() = (%s); want (http://127.0.0.1:9200)", c.Endpoint())
	}
	if c.TypeName() != "example_type" {
		t.Errorf("c.TypeName() = (%s); want (example_type)", c.TypeName())
	}
}

type badMarshaller struct{}

func (badMarshaller) UnmarshalKey(key string, rawVal interface{}) error { return errors.New("text") }
func (badMarshaller) AllKeys() []string                                 { return []string{} }

func TestWithViperBadFile(t *testing.T) {
	v := viper.New()
	v.SetConfigType("yaml")
	c := new(expvar.Config)
	tcs := []struct {
		name  string
		input *bytes.Buffer
	}{
		{
			name: "timeout",
			input: bytes.NewBuffer([]byte(`
    recorders:
        recorder1:
                index_name: example_index
                timeout: abc
                interval: 1s
    `)),
		},
		{
			name: "bad interval",
			input: bytes.NewBuffer([]byte(`
    recorders:
        recorder1:
                index_name: example_index
                timeout: 1s
                interval: def
    `)),
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			v.ReadConfig(tc.input)
			err := expvar.WithViper(v, "recorder1", "recorders.recorder1")(c)
			if err == nil {
				t.Error("err = (nil); want (error)")
			}
		})
	}

	err := expvar.WithViper(&badMarshaller{}, "recorder1", "recorders.recorder1")(c)
	if err == nil {
		t.Error("err = (nil); want (error)")
	}
}

func TestNewConfig(t *testing.T) {
	log := tools.DiscardLogger()
	c, err := expvar.NewConfig(
		expvar.WithLogger(log),
	)
	if err != nil {
		t.Errorf("err = (%v); want (nil)", err)
	}
	if c == nil {
		t.Error("c = (nil); want (Config)")
	}
	if c.Mapper() != datatype.DefaultMapper() {
		t.Errorf("c.Mapper() = (%v); want (%v)", c.Mapper(), datatype.DefaultMapper())
	}
}

func TestNewConfigErrors(t *testing.T) {
	c, err := expvar.NewConfig(
		expvar.WithLogger(nil),
	)
	if err == nil {
		t.Error("err = (nil); want (error)")
	}
	if c != nil {
		t.Errorf("c = (%v); want (nil)", c)
	}
}

func TestWithMapFile(t *testing.T) {
	c := new(expvar.Config)
	err := expvar.WithMapFile("")(c)
	if err != nil {
		t.Errorf("err = (%v); want (nil)", err)
	}

	cwd, _ := os.Getwd()
	file, err := ioutil.TempFile(cwd, "yaml")
	if err != nil {
		panic(err)
	}
	oldName := file.Name() //required for viper
	newName := file.Name() + ".yml"
	os.Rename(oldName, newName)
	defer os.Remove(newName)

	err = expvar.WithMapFile(path.Base(file.Name()))(c)
	if err != nil {
		t.Errorf("err = (%v); want (nil)", err)
	}

	err = expvar.WithMapFile("this file does not exist")(c)
	if err == nil {
		t.Error("err = (nil); want (error)")
	}
}

func TestConfigReader(t *testing.T) {
	log := tools.DiscardLogger()
	c, err := expvar.NewConfig(
		expvar.WithLogger(log),
	)
	c.EXPName = "name"
	c.EXPTypeName = "name"
	c.EXPEndpoint = "http://localhost"
	c.ConfInterval = time.Second
	if err != nil {
		t.Fatalf("err = (%v); want (nil)", err)
	}
	e, err := c.Reader()
	if err == nil {
		t.Error("err = (nil); want (error)")
	}
	if e.(*expvar.Reader) != nil {
		t.Errorf("e = (%v); want (nil)", e)
	}
	c.ConfTimeout = time.Second
	e, err = c.Reader()
	if err != nil {
		t.Errorf("err = (%v); want (nil)", err)
	}
	if e.(*expvar.Reader) == nil {
		t.Error("e.(*expvar.Reader) = (nil); want (Reader)")
	}
}
