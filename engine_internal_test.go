// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package expvastic_test

// func TestInspectResult(t *testing.T) {
// 	buf := ioutil.NopCloser(strings.NewReader(`{"key": 6.6}`))
// 	r := reader.ReadJobResult{
// 		Res:  buf,
// 		Time: time.Now(),
// 	}

// 	res := datatype.JobResultDataTypes(r.Res)
// 	if res.Error() != nil {
// 		t.Errorf("expected no errors, got: %s", res.Error())
// 	}
// 	if res.Len() == 0 {
// 		t.Error("expected results, got nothing")
// 	}

// 	buf = ioutil.NopCloser(strings.NewReader(`{"key: 6.6}`))
// 	r = reader.ReadJobResult{
// 		Res:  buf,
// 		Time: time.Now(),
// 	}

// 	res = datatype.JobResultDataTypes(r.Res)
// 	if res.Error() == nil {
// 		t.Error("expected an error, got nothing")
// 	}

// 	if res.Len() != 0 {
// 		t.Errorf("expected no results, got %s", res)
// 	}

// }
