package main

import (
	"context"
	"regexp"
	"testing"
	"time"
)

// A whole shard shares the 75-second kill deadline; its performance budget is
// 60 seconds. Each subprocess inherits the remaining shard deadline, so case
// growth cannot multiply the limit. A deadline kill is reported as COOKED.
func inventoryEngineDeadline(test *testing.T, binary, directory string, cases []string) {
	test.Helper()
	started := time.Now()
	parent := context.Background()
	cancelParent := func() {}
	if deadline, ok := test.Deadline(); ok {
		parent, cancelParent = context.WithDeadline(parent, deadline.Add(-time.Second))
	}
	defer cancelParent()
	timeoutContext, cancel := context.WithTimeout(parent, 75*time.Second)
	defer cancel()
	defer func() {
		test.Logf(
			"shard time %.6fs cooked=%t",
			time.Since(started).Seconds(),
			timeoutContext.Err() == context.DeadlineExceeded,
		)
	}()
	for _, name := range cases {
		command, commandContext, cancelCommand := inventoryEngineCommand(
			timeoutContext,
			binary,
			"-test.run=^"+regexp.QuoteMeta(name)+"$",
			"-test.count=1",
			"-test.v",
			"-test.timeout=75s",
		)
		command.Dir = directory
		output, inventoryEngineOutputError := inventoryEngineOutput(commandContext, command)
		cancelCommand()
		test.Logf("%s", output)
		if timeoutContext.Err() == context.DeadlineExceeded {
			test.Fatalf(
				"COOKED: shard/test deadline reached (75s shard limit) while running %s; split smaller before rerunning",
				name,
			)
		}
		if inventoryEngineOutputError != nil {
			test.Fatalf("case %s: %v", name, inventoryEngineOutputError)
		}
	}
	if elapsed := time.Since(started); elapsed > 60*time.Second {
		test.Fatalf("OVER BUDGET: shard took %s, budget 60s; split smaller before rerunning", elapsed)
	}
}
