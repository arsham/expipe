// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package communication_test

import (
	"context"
	"fmt"

	"github.com/arsham/expvastic/communication"
)

// This example shows how to create a new job from a context.
func ExampleNewReadJob() {
	job := communication.NewReadJob(context.Background())
	_ = job
}

// This example shows how to pass a jobID around and how to
// get the ID back.
func ExampleJobValue_fromJob() {
	job := communication.NewReadJob(context.Background())
	// pass the job around....
	jobID := communication.JobValue(job)
	fmt.Println(jobID)
}
