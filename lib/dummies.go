// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package lib

// DummyReadCloser implements io.ReadCloser that does nothing
type DummyReadCloser struct{}

// Close is a noop
func (DummyReadCloser) Close() error                     { return nil }
func (DummyReadCloser) Read(p []byte) (n int, err error) { return 0, nil }
