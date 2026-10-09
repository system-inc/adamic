//go:build lintoracle

package main

import "testing"

func TestNoCallerOptionsPayload(t *testing.T) {
	t.Parallel()
	for _, input := range []string{`{}`, `{"ignored": true}`, `[]`} {
		fields := []string{"", "no-caller", "", "", "", input}
		if oracleNoCallerOptions(fields) == nil {
			t.Fatalf("explicit options %s were dropped", input)
		}
	}
	if oracleNoCallerOptions([]string{"", "no-caller"}) != nil {
		t.Fatal("absent options must retain the upstream default")
	}
	defer func() {
		if recover() == nil {
			t.Fatal("malformed options were silently accepted")
		}
	}()
	oracleNoCallerOptions([]string{"", "no-caller", "", "", "", "{"})
}
