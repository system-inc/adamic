package lint

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var lintMutantComponents = map[string]string{
	"oracle":           "oracle/adapter/no-var",
	"capture":          "capture/filter",
	"go-output":        "go-output/manifest",
	"port":             "port/modules",
	"node":             "node/modules",
	"javascript":       "javascript/modules",
	"javascript-build": "javascript-build/modules",
}

func lintChange(t *testing.T, path, from, to string) {
	t.Helper()
	data := lintBytes(t, path)
	if strings.Count(data, from) != 1 {
		t.Fatalf("cache probe anchor count for %s", path)
	}
	if err := os.WriteFile(path, []byte(strings.Replace(data, from, to, 1)), 0600); err != nil {
		t.Fatal(err)
	}
}

// Real inputs change behind a populated cache. The expected answer is always
// obtained from the original uncached implementation, independently of its key.
func TestLintCacheInvalidation(t *testing.T) {
	t.Setenv("GOCACHE", strings.TrimSpace(lintTool(t, "go", "env", "GOCACHE")))
	t.Setenv("ADAMIC_GATE_UNCACHED", "0")
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	asked := os.Getenv("ADAMIC_LINT_CACHE_PROBE")
	if drop := os.Getenv("ADAMIC_LINT_CACHE_DROP"); drop != "" {
		lintKeyDrop = drop
		defer func() { lintKeyDrop = "" }()
	}
	for _, kind := range []string{"oracle", "capture", "go-output", "port", "node", "javascript", "javascript-build"} {
		if asked != "" && asked != kind {
			continue
		}
		t.Run(kind, func(t *testing.T) {
			d := selectedDescriptor(t, "no-var")
			directory := selectedPort(t, d)
			source := filepath.Join(t.TempDir(), "input.ts")
			lintPublish(t, source, []byte("var value = 1;\n"), 0600)
			path := manifest(t, []string{source + "\tno-var"})
			var before, got, want []byte
			switch kind {
			case "oracle":
				before = execute(t, "", goOracleFrom(t, directory), "--manifest", path).output
				lintChange(t, filepath.Join(directory, "rules/no-var/oracle.go"), "return rules.NoVar", "subject := rules.NoVar; subject.Run = func(ctx rule.Context, options any) rule.Listeners { return nil }; return subject")
				got = execute(t, "", goOracleFrom(t, directory), "--manifest", path).output
				want = executeUncached(t, "", goOracleFromUncached(t, directory), "--manifest", path).output
			case "capture":
				d.UpstreamTest = "TestNoVarFires"
				before = lintCaptureObservation(t, lintCapture(t, ".", d))
				d.UpstreamTest = "TestNoVarStaysSilent"
				got = lintCaptureObservation(t, lintCapture(t, ".", d))
				want = lintCaptureObservation(t, upstreamSelection(t, ".", &d))
			case "go-output":
				oracle := goOracle(t)
				before = execute(t, "", oracle, "--manifest", path).output
				lintPublish(t, source, []byte("let value = 1;\n"), 0600)
				got = execute(t, "", oracle, "--manifest", path).output
				want = executeUncached(t, "", oracle, "--manifest", path).output
			case "port":
				before = execute(t, "", buildPort(t, directory, true), "--manifest", path).output

				// Add an observable change; both versions still compile and run cleanly.
				main := filepath.Join(directory, "main.ts")
				lintPublish(t, main, []byte(lintBytes(t, main)+"\nconsole.log('cache changed');\n"), 0600)
				got = execute(t, "", buildPort(t, directory, true), "--manifest", path).output
				want = execute(t, "", buildPortUncached(t, directory, true), "--manifest", path).output
			case "node":
				before = node(t, directory, path, false).output
				main := filepath.Join(directory, "main.ts")
				lintPublish(t, main, []byte(lintBytes(t, main)+"\nconsole.log('cache changed');\n"), 0600)
				got = node(t, directory, path, false).output
				want = nodeUncached(t, directory, path, false).output
			case "javascript":
				module := emittedJavaScript(t, directory)
				before = runJavaScript(t, module, path, false).output
				edited := filepath.Join(t.TempDir(), "edited.mjs")
				lintPublish(t, edited, []byte(lintBytes(t, module)+"\nconsole.log('cache changed');\n"), 0600)
				got = runJavaScript(t, edited, path, false).output
				want = runJavaScriptUncached(t, edited, path, false).output
			case "javascript-build":
				before = []byte(lintBytes(t, emittedJavaScript(t, directory)))
				main := filepath.Join(directory, "main.ts")
				lintPublish(t, main, []byte(lintBytes(t, main)+"\nconsole.log('cache changed');\n"), 0600)
				got = []byte(lintBytes(t, emittedJavaScript(t, directory)))
				want = []byte(lintBytes(t, emittedJavaScriptUncached(t, directory)))
			}
			if bytes.Equal(before, want) {
				t.Fatalf("%s probe did not change the uncached answer", kind)
			}
			if !bytes.Equal(got, want) {
				t.Fatalf("stale %s answer after guarded input changed; cached_matches_before=%t; before=%d cached=%d fresh=%d bytes", kind, bytes.Equal(got, before), len(before), len(got), len(want))
			}
			t.Logf("%s changed real input and invalidated its observation", kind)
		})
	}
}

// Not parallel: mutant processes intentionally do expensive builds in sequence.
// Only the omitted key component kills each check; all inputs remain valid.
func TestLintCacheMutants(t *testing.T) {
	for _, kind := range []string{"oracle", "capture", "go-output", "port", "node", "javascript", "javascript-build"} {
		t.Run(kind, func(t *testing.T) {
			log := filepath.Join(t.TempDir(), "mutant.log")
			output, err := os.Create(log)
			if err != nil {
				t.Fatal(err)
			}
			command := exec.Command(os.Args[0], "-test.run=^TestLintCacheInvalidation$", "-test.v", "-test.timeout=30m")
			command.Env = append(os.Environ(), "ADAMIC_LINT_CACHE_PROBE="+kind, "ADAMIC_LINT_CACHE_DROP="+lintMutantComponents[kind])
			command.Stdout, command.Stderr = output, output
			runError := command.Run()
			if err := output.Close(); err != nil {
				t.Fatal(err)
			}
			data := []byte(lintBytes(t, log))
			if runError == nil || (!bytes.Contains(data, []byte("stale "+kind+" answer after guarded input changed")) || !bytes.Contains(data, []byte("cached_matches_before=true"))) {
				t.Fatalf("wrong %s cache mutant failure: %v\n%s", kind, runError, data)
			}
			t.Logf("dropped %s; real input changed; invalidation test rejected stale answer:\n%s", lintMutantComponents[kind], data)
		})
	}
}
func TestLintCacheBypassAndIntegrity(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	t.Setenv("ADAMIC_GATE_UNCACHED", "0")
	key := lintKey("probe", map[string]string{"input": "fixed"})
	runs := 0
	run := func() []byte { runs++; return []byte{0, 255, byte(runs)} }
	first := lintCachedBytes(t, "probe", key, run)
	if !bytes.Equal(first, lintCachedBytes(t, "probe", key, run)) || runs != 1 {
		t.Fatal("valid evidence not reused byte for byte")
	}
	t.Setenv("ADAMIC_GATE_UNCACHED", "1")
	fresh := lintCachedBytes(t, "probe", key, run)
	if runs != 2 || bytes.Equal(first, fresh) {
		t.Fatal("bypass reused old evidence")
	}
	t.Setenv("ADAMIC_GATE_UNCACHED", "0")
	if !bytes.Equal(first, lintCachedBytes(t, "probe", key, run)) {
		t.Fatal("uncached mode published evidence")
	}
	lintPublish(t, filepath.Join(lintCacheRoot(t, "probe"), key+".json"), []byte("partial"), 0600)
	lintCachedBytes(t, "probe", key, run)
	if runs != 3 {
		t.Fatal("corrupt entry was trusted")
	}
}
