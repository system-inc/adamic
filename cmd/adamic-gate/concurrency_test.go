package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestConcurrencyBudgetAndResumeIdentity(t *testing.T) {
	t.Parallel()
	for _, setting := range []string{"auto", "4x1", "2x2"} {
		c, err := resolveConcurrency(setting, 0)
		if err != nil || validateConcurrency(c) != nil {
			t.Fatal(c, err)
		}
		if setting != "auto" && c.Budget != 4 {
			t.Fatal(c)
		}
		args := concurrencyArgs(testArgs("pkg", "^TestParent$/^child$"), c.Parallel)
		if args[len(args)-1] != "pkg" || args[6] != "^TestParent$/^child$" {
			t.Fatal("selector changed", args)
		}
		if (setting == "auto") != !strings.Contains(strings.Join(args, " "), "-parallel=") {
			t.Fatal("parallel flag not exact", args)
		}
	}
	for _, setting := range []string{"", "0x1", "2x0", "4x-1", "1x1025", "1x2x3", "garbage"} {
		if _, err := resolveConcurrency(setting, 0); err == nil {
			t.Fatal("invalid budget accepted", setting)
		}
	}
	if _, err := resolveConcurrency("2x2", 4); err == nil {
		t.Fatal("conflicting jobs accepted")
	}
	if schedulingIdentity("inputs", 4, 1, 1) == schedulingIdentity("inputs", 4, 2, 2) {
		t.Fatal("parallel missing from resume identity")
	}
	if schedulingIdentity("inputs", 4, 0, 5) == schedulingIdentity("inputs", 4, 5, 5) {
		t.Fatal("explicit parallel flag missing from resume identity")
	}
	if schedulingIdentity("inputs", 4, 0, 4) == schedulingIdentity("inputs", 4, 0, 2) {
		t.Fatal("effective automatic parallel missing from resume identity")
	}
	c, _ := resolveConcurrency("2x2", 0)
	c.Budget = 16
	if validateConcurrency(c) == nil {
		t.Fatal("wrong recorded budget accepted")
	}
	if !reflect.DeepEqual(concurrencyArgs(testArgs("pkg", "^Test$"), 0), testArgs("pkg", "^Test$")) {
		t.Fatal("auto settings changed today's Go arguments")
	}
}

func TestConcurrencyCheckpointMatchesActualInvocation(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := plan{Count: 1, Units: []unit{{Package: "pkg", Test: "TestPass"}}}
	e := packageEvidence{Package: "pkg", TestParallel: 1, Invocations: []invocation{{Package: "pkg", Args: concurrencyArgs(testArgs("pkg", "^TestPass$"), 1), Uncached: true}}}
	log := `{"Action":"run","Package":"pkg","Test":"TestPass"}
{"Action":"pass","Package":"pkg","Test":"TestPass"}
{"Action":"pass","Package":"pkg"}
`
	if err := os.WriteFile(filepath.Join(dir, "test.jsonl"), []byte(log), 0600); err != nil {
		t.Fatal(err)
	}
	if err := completePackage(dir, e, p, 0, "pkg"); err != nil {
		t.Fatal(err)
	}
	e.TestParallel = 2
	if err := completePackage(dir, e, p, 0, "pkg"); err == nil {
		t.Fatal("different parallel invocation accepted")
	}
}
