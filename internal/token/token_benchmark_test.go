// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package token_test

import (
	"context"
	"testing"

	"github.com/arsham/expvastic/internal/token"
)

var result *token.Context

func BenchmarkGenerateTokens(b *testing.B) {
	var t *token.Context
	ctx := context.Background()

	for n := 0; n < b.N; n++ {
		t = token.New(ctx)
	}

	result = t
}
