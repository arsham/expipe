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
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/arsham/expipe/internal/app"
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
		backoff   int
		indexName string
		reader    string
	}{
		{"", 0, 0, "", ""},
		{"localhost:9200", fakeDuration, 0, "", ""},
		{"localhost:9200", time.Second, 0, "", ""},
		{"localhost:9200", time.Second, 20, "", ""},
		{"localhost:9200", time.Second, 20, "222", ""},
		{"localhost:9200", time.Second, 20, "222", "sss"},
		{"localhost:9200", time.Second, 2, "222", "sss"},
		{"localhost:9200", time.Second, 20, "", "sss"},
		{"localhost:9200", time.Second, 2, "222", "localhost6/dev"},
		{"localhost:9200", fakeDuration, 20, "222", "localhost7/dev"},
	}
	for i, tc := range tcs {
		os.Unsetenv("CONFIG")
		os.Setenv("RECORDER", tc.recorder)
		os.Setenv("TIMEOUT", tc.timeout.String())
		os.Setenv("BACKOFF", strconv.Itoa(tc.backoff))
		os.Setenv("INDEXNAME", tc.indexName)
		os.Setenv("READER", tc.reader)
		name := fmt.Sprintf("fromFlagsCase_%d", i)
		t.Run(name, func(t *testing.T) {
			defer func() {
				os.Unsetenv("CONFIG")
				os.Unsetenv("RECORDER")
				os.Unsetenv("TIMEOUT")
				os.Unsetenv("BACKOFF")
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
	os.Setenv("BACKOFF", strconv.Itoa(20))
	os.Setenv("INDEX", "222")
	os.Setenv("TYPE", "222")
	defer func() {
		os.Unsetenv("CONFIG")
		os.Unsetenv("READER")
		os.Unsetenv("RECORDER")
		os.Unsetenv("TIMEOUT")
		os.Unsetenv("BACKOFF")
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
	os.Setenv("BACKOFF", strconv.Itoa(20))
	os.Setenv("INDEX", "222")
	os.Setenv("TYPE", "222")
	defer func() {
		os.Unsetenv("CONFIG")
		os.Unsetenv("READER")
		os.Unsetenv("RECORDER")
		os.Unsetenv("TIMEOUT")
		os.Unsetenv("BACKOFF")
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
