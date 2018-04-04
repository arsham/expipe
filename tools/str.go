// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package tools

import "encoding/json"

// StringInSlice returns true if niddle is in the haystack
func StringInSlice(niddle string, haystack []string) bool {
	for _, b := range haystack {
		if b == niddle {
			return true
		}
	}
	return false
}

// StringInMapKeys returns true if niddle is in the haystack's keys
func StringInMapKeys(niddle string, haystack map[string]string) bool {
	for b := range haystack {
		if b == niddle {
			return true
		}
	}
	return false
}

// IsJSON checks is the input content is a valid JSON input.
func IsJSON(input []byte) bool {
	var js json.RawMessage
	return json.Unmarshal(input, &js) == nil
}
