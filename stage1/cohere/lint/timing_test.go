package lint

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// Not parallel: the throughput switch is process-wide. PIDs prove two real
// processes ran; the source-Node probe reads an input outside the lint corpus.
func TestTimedObservationsUncached(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	t.Setenv("ADAMIC_GATE_UNCACHED", "0")
	t.Setenv("ADAMIC_LINT_BENCH", "1")
	path := manifest(t, nil)
	t.Run("Go", func(t *testing.T) {
		binary := filepath.Join(t.TempDir(), "oracle")
		lintPublish(t, binary, []byte("#!/bin/sh\nprintf '%s\\n' \"$$\"\n"), 0700)
		lintOracles.Store(binary, "timing-probe")
		defer lintOracles.Delete(binary)
		a := execute(t, "", binary, "--manifest", path).output
		b := execute(t, "", binary, "--manifest", path).output
		if bytes.Equal(a, b) {
			t.Fatal("timed Go observation replayed")
		}
	})
	t.Run("JavaScript", func(t *testing.T) {
		module := filepath.Join(t.TempDir(), "probe.mjs")
		lintPublish(t, module, []byte("console.log(process.pid);\n"), 0600)
		a := runJavaScript(t, module, path, false).output
		b := runJavaScript(t, module, path, false).output
		if bytes.Equal(a, b) {
			t.Fatal("timed JavaScript observation replayed")
		}
	})
	t.Run("Node", func(t *testing.T) {
		directory := selectedPort(t, selectedDescriptor(t, "no-var"))
		input := filepath.Join(t.TempDir(), "clock.txt")
		main := filepath.Join(directory, "main.ts")
		lintPublish(t, main, []byte(lintBytes(t, main)+fmt.Sprintf("\nconst clock = readTextFile(%q); if(clock.kind === 'Ok') { console.log(clock.text); }\n", input)), 0600)
		lintPublish(t, input, []byte("first sample"), 0600)
		first := node(t, directory, path, false).output
		lintPublish(t, input, []byte("next sample"), 0600)
		next := node(t, directory, path, false).output
		if bytes.Equal(first, next) || !bytes.Contains(next, []byte("next sample")) {
			t.Fatal("timed Node observation replayed")
		}
	})
	if os.Getenv("ADAMIC_GATE_UNCACHED") != "0" {
		t.Fatal("timing probe accidentally used the integration bypass")
	}
}
