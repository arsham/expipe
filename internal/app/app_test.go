// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package app_test

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"syscall"
	"testing"
	"time"

	rdt "github.com/arsham/expipe/reader/testing"
	rct "github.com/arsham/expipe/recorder/testing"

	"github.com/arsham/expipe/internal/app"
	"github.com/arsham/expipe/reader"
	"github.com/arsham/expipe/recorder"
	"github.com/arsham/expipe/tools"
	"github.com/arsham/expipe/tools/config"
	flags "github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
)

// returns the file base name and tear down function
func setup(content []byte) (string, func()) {
	cwd, _ := os.Getwd()
	file, err := ioutil.TempFile(cwd, "yaml")
	if err != nil {
		panic(err)
	}
	oldName := file.Name() //required for viper
	newName := file.Name() + ".yml"
	os.Rename(oldName, newName)
	file.Write(content)
	return path.Base(file.Name()), func() {
		os.Remove(newName)
	}
}

// readFixtures reads the fixture file.
func readFixtures(t *testing.T, filename string) [][]byte {
	f, err := os.Open("testdata/" + filename)
	if err != nil {
		t.Fatalf("reading test fixtures: %v", err)
	}
	contents, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatalf("reading test fixtures: %v", err)
	}

	ret := bytes.Split(contents, []byte("==="))
	if len(ret) < 1 {
		t.Fatalf("reading test fixtures: %v", err)
	}
	return ret
}

func TestConfigLogLevel(t *testing.T) {
	p := flags.NewParser(&app.Opts, flags.IgnoreUnknown)
	p.Parse()
	log, _, _ := app.Config()
	if log.Level != tools.InfoLevel {
		t.Errorf("want (info), got (%s)", log.Level)
	}

	os.Setenv("LOGLEVEL", "warn")
	p = flags.NewParser(&app.Opts, flags.IgnoreUnknown)
	p.Parse()

	log, conf, err := app.Config()
	if log.Level != tools.WarnLevel {
		t.Errorf("want (warn), got (%s)", log.Level)
	}
	if errors.Cause(err) != nil {
		t.Errorf("err = (%v); want (nil)", err)
	}
	if conf == nil {
		t.Error("want (Config), got (nil)")
	}
}

func TestConfigFileDoesNotExists(t *testing.T) {
	os.Setenv("CONFIG", "thisfiledoesnotexists")
	p := flags.NewParser(&app.Opts, flags.IgnoreUnknown)
	p.Parse()

	_, conf, err := app.Config()
	if errors.Cause(err) == nil {
		t.Error("err = (nil); want (error)")
	}
	if conf != nil {
		t.Errorf("want (nil), got (%v)", conf)
	}
}

func TestMainAndFromConfigFileErrors(t *testing.T) {
	tcs := readFixtures(t, "main_and_from_config_file_errors.txt")
	for i, tc := range tcs {
		name := fmt.Sprintf("fromFlagsCase_%d", i)
		t.Run(name, func(t *testing.T) {
			filename, teardown := setup(tc)
			defer teardown()
			defer func() {
				os.Unsetenv("CONFIG")
			}()

			os.Setenv("CONFIG", filename)
			p := flags.NewParser(&app.Opts, flags.IgnoreUnknown)
			p.Parse()

			_, result, err := app.Config()
			if err == nil {
				t.Error("want error, got nothing")
			}
			if result != nil {
				t.Errorf("want nil, got (%v)", result)
			}
		})
	}
}

func TestMainAndFromConfigFilePasses(t *testing.T) {
	filename, teardown := setup(readFixtures(t, "main_and_from_config_file_passes.txt")[0])
	defer teardown()
	defer func() {
		os.Unsetenv("CONFIG")
	}()

	os.Setenv("CONFIG", filename)
	p := flags.NewParser(&app.Opts, flags.IgnoreUnknown)
	p.Parse()

	_, result, err := app.Config()
	if err != nil {
		t.Errorf("want nil, got (%v)", err)
	}
	if reflect.TypeOf(result) != reflect.TypeOf(&config.ConfMap{}) {
		t.Errorf("want config.ConfMap, got (%v)", result)
	}
}

func TestMainAndFromFlagsErrors(t *testing.T) {
	os.Unsetenv("CONFIG")
	fakeDuration, _ := time.ParseDuration("dfdfdf")
	tcs := []struct {
		recorder  string
		timeout   time.Duration
		indexName string
		reader    string
	}{
		{"", 0, "", ""},
		{"localhost:9200", fakeDuration, "", ""},
		{"localhost:9200", time.Second, "", ""},
		{"localhost:9200", time.Second, "", ""},
		{"localhost:9200", time.Second, "222", ""},
		{"localhost:9200", time.Second, "222", "sss"},
		{"localhost:9200", time.Second, "222", "sss"},
		{"localhost:9200", time.Second, "", "sss"},
		{"localhost:9200", time.Second, "222", "localhost6/dev"},
		{"localhost:9200", fakeDuration, "222", "localhost7/dev"},
	}
	for i, tc := range tcs {
		os.Unsetenv("CONFIG")
		os.Setenv("RECORDER", tc.recorder)
		os.Setenv("TIMEOUT", tc.timeout.String())
		os.Setenv("INDEXNAME", tc.indexName)
		os.Setenv("READER", tc.reader)
		name := fmt.Sprintf("fromFlagsCase_%d", i)
		t.Run(name, func(t *testing.T) {
			defer func() {
				os.Unsetenv("CONFIG")
				os.Unsetenv("RECORDER")
				os.Unsetenv("TIMEOUT")
				os.Unsetenv("INDEXNAME")
				os.Unsetenv("READER")
			}()

			p := flags.NewParser(&app.Opts, flags.IgnoreUnknown)
			p.Parse()

			_, result, err := app.Config()
			if err == nil {
				t.Error("want error, got nothing")
			}
			if result != nil {
				t.Errorf("want nil, got (%v)", result)
			}
		})
	}
}

func TestMainAndFromFlagsPasses(t *testing.T) {
	os.Unsetenv("CONFIG")
	os.Setenv("READER", "localhost1:222/dev")
	os.Setenv("RECORDER", "localhost2:9200")
	os.Setenv("TIMEOUT", time.Second.String())
	os.Setenv("INDEX", "222")
	os.Setenv("TYPE", "222")
	defer func() {
		os.Unsetenv("CONFIG")
		os.Unsetenv("READER")
		os.Unsetenv("RECORDER")
		os.Unsetenv("TIMEOUT")
		os.Unsetenv("INDEX")
		os.Unsetenv("TYPE")
	}()
	p := flags.NewParser(&app.Opts, flags.IgnoreUnknown)
	p.Parse()

	_, result, err := app.Config()
	if err != nil {
		t.Errorf("want nil, got (%v)", err)
	}
	if reflect.TypeOf(result) != reflect.TypeOf(&config.ConfMap{}) {
		t.Errorf("want config.ConfMap, got (%v)", result)
	}
}

func TestConfig(t *testing.T) {
	os.Unsetenv("CONFIG")
	os.Setenv("READER", "localhost1:222/dev")
	os.Setenv("RECORDER", "localhost2:9200")
	os.Setenv("TIMEOUT", time.Second.String())
	os.Setenv("INDEX", "222")
	os.Setenv("TYPE", "222")
	defer func() {
		os.Unsetenv("CONFIG")
		os.Unsetenv("READER")
		os.Unsetenv("RECORDER")
		os.Unsetenv("TIMEOUT")
		os.Unsetenv("INDEX")
		os.Unsetenv("TYPE")
	}()
	p := flags.NewParser(&app.Opts, flags.IgnoreUnknown)
	p.Parse()

	_, result, err := app.Config()
	if err != nil {
		t.Errorf("want nil, got (%v)", err)
	}
	if reflect.TypeOf(result) != reflect.TypeOf(&config.ConfMap{}) {
		t.Errorf("want config.ConfMap, got (%v)", result)
	}
}

func TestCaptureSignals(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal)
	exitCh := make(chan int)
	exit := func(code int) {
		exitCh <- code
	}
	timeout := 10 * time.Millisecond
	app.CaptureSignals(cancel, sigCh, exit, timeout)
	sigCh <- syscall.SIGINT
	select {
	case <-ctx.Done():
	case <-time.After(timeout):
		t.Error("context wasn't cancelled")
	}
	select {
	case code := <-exitCh:
		if code != 130 {
			t.Errorf("want to exit with code (130), got (%d)", code)
		}
	case <-time.After(timeout * 2):
		t.Error("exit function wasn't called")
	}
}

func TestConfigReadSampleYAML(t *testing.T) {
	filename, teardown := setup(readFixtures(t, "config_read_sample_yaml.txt")[0])
	defer teardown()
	defer func() {
		os.Unsetenv("CONFIG")
	}()

	os.Setenv("CONFIG", filename)
	p := flags.NewParser(&app.Opts, flags.IgnoreUnknown)
	p.Parse()

	_, result, _ := app.Config()
	if len(result.Readers) != 3 {
		t.Errorf("len(result.Readers) = (%d); want (3)", len(result.Readers))
	}
	if len(result.Routes) != 3 {
		t.Errorf("len(result.Routes) = (%d); want (3)", len(result.Routes))
	}
	// for each routes

}

type logger struct {
	tools.FieldLogger
	FatalfFunc func(string, ...interface{})
	FatalFunc  func(...interface{})
}

func (l logger) Fatalf(f string, a ...interface{}) { l.FatalfFunc(f, a) }
func (l logger) Fatal(a ...interface{})            { l.FatalfFunc("", a) }

func TestBootstrapFatal(t *testing.T) {
	if testing.Short() {
		return
	}

	called := make(chan struct{})
	conf := &config.ConfMap{
		Readers:   map[string]reader.DataReader{"red1": nil},
		Recorders: map[string]recorder.DataRecorder{},
		Routes:    map[string][]string{"red1": nil},
	}
	ctx := context.Background()
	log := &logger{
		FieldLogger: tools.StandardLogger(),
		FatalfFunc: func(f string, a ...interface{}) {
			close(called)
		},
	}

	go func() {
		app.Bootstrap(ctx, log, conf)
	}()

	select {
	case <-called:
	case <-time.After(time.Second):
		t.Error("Fatalf() wasn't called")
	}
}

func TestBootstrap(t *testing.T) {
	if testing.Short() {
		return
	}

	conf := &config.ConfMap{
		Readers: map[string]reader.DataReader{"red1": &rdt.Reader{
			MockName:     "name",
			MockInterval: time.Second,
			Pinged:       true,
		}},
		Recorders: map[string]recorder.DataRecorder{"rec1": &rct.Recorder{
			MockName: "name",
			Pinged:   true,
		}},
		Routes: map[string][]string{"red1": {"rec1", "rec2"}},
	}
	ctx, cancel := context.WithCancel(context.Background())
	log := tools.DiscardLogger()
	step := make(chan struct{})

	go func() {
		app.Bootstrap(ctx, log, conf)
		step <- struct{}{}
	}()

	select {
	case <-step:
		t.Error("Bootstrap() finished unexpectedly")
	case <-time.After(time.Second):
	}
	cancel()
	select {
	case <-step:
	case <-time.After(time.Second * 3):
		t.Error("Bootstrap() didn't quit")
	}
}
