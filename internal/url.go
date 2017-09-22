// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package internal

import (
	"fmt"
	"strings"

	"github.com/asaskevich/govalidator"
)

type errInvalidURL string

// InvalidURL is the behaviour of an errInvalidURL error.
func (errInvalidURL) InvalidURL() {}

// Error returns the string representation of the error.
func (i errInvalidURL) Error() string { return fmt.Sprintf("invalid url: %s", string(i)) }

// SanitiseURL prepends a protocol to the url if not defined, and checks if it's a valid url.
func SanitiseURL(url string) (string, error) {
	if govalidator.IsURL(url) {
		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			url = "http://" + url
		}
		return url, nil
	}

	return "", errInvalidURL(url)
}
