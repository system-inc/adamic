package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// A probe, never submitted: run one explain driver the way a box without ADAMIC_BUILD_CACHE_DIR does, with and
// without the drivers' fresh XDG_CACHE_HOME, and report both. It always fails, so the report reaches the log.
func TestProbeExplainDriverCache(t *testing.T) {
	repository, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	var environment []string
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, "ADAMIC_BUILD_CACHE_DIR=") && !strings.HasPrefix(entry, "XDG_CACHE_HOME=") {
			environment = append(environment, entry)
		}
	}
	shared := t.TempDir()
	var report []string
	for _, c := range []struct{ name, xdg string }{{"fresh xdg", t.TempDir()}, {"shared xdg", shared}, {"shared xdg again", shared}} {
		build := exec.Command(os.Args[0], "-test.run=^TestProduct_CommandRuntimeSanitized$")
		build.Env = append(append([]string(nil), environment...), "XDG_CACHE_HOME="+shared)
		buildOut, buildErr := build.CombinedOutput()
		command := exec.Command(os.Args[0], "-test.run=^TestExplainChecksDriver$", "--", "build", "internal/lower/testdata/predicates/overload_erased.a", "-o", filepath.Join(t.TempDir(), "program"), "--sanitize", "--explain-checks")
		command.Dir = repository
		command.Env = append(append([]string(nil), environment...), "ADAMIC_EXPLAIN_TEST_DRIVER=1", "XDG_CACHE_HOME="+c.xdg)
		var diagnostic bytes.Buffer
		command.Stderr = &diagnostic
		err := command.Run()
		last := strings.TrimSpace(diagnostic.String())
		if index := strings.LastIndex(last, "\n"); index >= 0 {
			last = last[index+1:]
		}
		report = append(report, c.name+": product build "+errText(buildErr)+" ("+strings.TrimSpace(lastLine(string(buildOut)))+"); driver "+errText(err)+": "+last)
	}
	t.Fatalf("\n%s", strings.Join(report, "\n"))
}

func errText(err error) string {
	if err == nil {
		return "ok"
	}
	return err.Error()
}

func lastLine(text string) string {
	text = strings.TrimSpace(text)
	if index := strings.LastIndex(text, "\n"); index >= 0 {
		return text[index+1:]
	}
	return text
}
