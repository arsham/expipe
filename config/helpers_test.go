// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	"github.com/arsham/expipe/internal"
)

type FixtureTestCase struct {
	Name string
	Body *bytes.Buffer
	Info string
}

var secReg = regexp.MustCompile(`name:\s*([^\n\s]+)\s*\n>>>[\s\n]*([[:ascii:]]*)<<<[\s\n]*info:\s*([^\n\s]+)`)

// ReadFixtures reads the fixture file.
func ReadFixtures(filename string) ([]FixtureTestCase, error) {
	ret := make([]FixtureTestCase, 0)
	f, err := os.Open("testdata/" + filename)
	if err != nil {
		return nil, err
	}
	contents, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	parts := bytes.Split(contents, []byte("==="))
	for _, part := range parts {
		m := secReg.FindStringSubmatch(string(part))
		if len(m) < 4 {
			return nil, fmt.Errorf("error parsing info: %v", string(part))
		}
		m = m[1:]
		d := FixtureTestCase{Name: m[0], Body: bytes.NewBuffer([]byte(m[1])), Info: m[2]}
		ret = append(ret, d)
	}
	return ret, nil
}

// FixtureWithSection reads a section of the fixture file.
func FixtureWithSection(filename, name string) (FixtureTestCase, error) {
	tcs, err := ReadFixtures(filename)
	if err != nil {
		return FixtureTestCase{}, err
	}
	for _, tc := range tcs {
		if tc.Name == name {
			return tc, nil
		}
	}
	return FixtureTestCase{}, fmt.Errorf("section not found: %s", name)
}

func equalSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !internal.StringInSlice(a[i], b) {
			return false
		}
	}
	return true
}
