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
func inventoryEngineDeadline(t *testing.T, binary, directory string, cases []string) {
	t.Helper()
	started := time.Now()
	parent := context.Background()
	cancelParent := func() {}
	if deadline, ok := t.Deadline(); ok {
		parent, cancelParent = context.WithDeadline(parent, deadline.Add(-time.Second))
	}
	defer cancelParent()
	ctx, cancel := context.WithTimeout(parent, 75*time.Second)
	defer cancel()
	defer func() {
		t.Logf("shard time %.6fs cooked=%t", time.Since(started).Seconds(), ctx.Err() == context.DeadlineExceeded)
	}()
	for _, name := range cases {
		command, cancelCommand := inventoryEngineCommand(ctx, binary, "-test.run=^"+regexp.QuoteMeta(name)+"$", "-test.count=1", "-test.v", "-test.timeout=75s")
		command.Dir = directory
		output, err := command.CombinedOutput()
		cancelCommand()
		t.Logf("%s", output)
		if ctx.Err() == context.DeadlineExceeded {
			t.Fatalf("COOKED: shard/test deadline reached (75s shard limit) while running %s; split smaller before rerunning", name)
		}
		if err != nil {
			t.Fatalf("case %s: %v", name, err)
		}
	}
	if elapsed := time.Since(started); elapsed > 60*time.Second {
		t.Fatalf("OVER BUDGET: shard took %s, budget 60s; split smaller before rerunning", elapsed)
	}
}
