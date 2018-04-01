// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package tools_test

import (
	"fmt"

	"github.com/arsham/expipe/tools"
)

// This example shows how to sanitise a URL.
func ExampleSanitiseURL() {
	res, err := tools.SanitiseURL("http localhost")
	fmt.Printf("Error: %v\n", err)
	fmt.Printf("Result: <%s>\n", res)

	res, err = tools.SanitiseURL("127.0.0.1")
	fmt.Printf("Error: %v\n", err)
	fmt.Printf("Result: <%s>\n", res)

	res, _ = tools.SanitiseURL("https://localhost.localdomain")
	fmt.Printf("Result: <%s>\n", res)

	// Output:
	// Error: invalid url: http localhost
	// Result: <>
	// Error: <nil>
	// Result: <http://127.0.0.1>
	// Result: <https://localhost.localdomain>
}
