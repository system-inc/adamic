package oracle

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Also exercise macOS's value-balance predicate on Linux, where globally
// reachable allocator storage would otherwise hide an unreleased value.
func TestScout33CountedBalance(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/scout_33/main.a"))
	if err != nil {
		t.Fatal(err)
	}
	source, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), source)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "counted")
	if err := native.Build(native.C(program), binary, native.Options{Count: true}); err != nil {
		t.Fatal(err)
	}
	result := executeWith(t, nil, binary)
	if report := leakcheck.Unbalanced(leakRun(result)); report != "" {
		t.Fatal(report)
	}
	t.Logf("%s", result.stderr)
}

// These are executable source shapes, not a ruling about disposing arbitrary
// checker cycles. Each semantic mutant finishes cleanly and differs from the
// original source on Node; the ownership mutant keeps the same observable bytes.
func TestScout33CacheEvictionAndMutants(t *testing.T) {
	root, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/scout_33"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range []string{"main.a", "cache.a"} {
		t.Run(entry, func(t *testing.T) {
			path := filepath.Join(root, entry)
			truth := onNode(t, path)
			if truth.exitCode != 0 {
				t.Fatalf("Node did not finish: %s", truth.stderr)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			code := native.C(program)
			binary := filepath.Join(t.TempDir(), "sanitized")
			if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			got := executeWith(t, []string{leakSanitizer()}, binary)
			if report := disagreement(truth, got); report != "" {
				t.Fatal(report)
			}
			if report := leakChecked(t, code, binary); report != "" {
				t.Fatal(report)
			}
			counted := filepath.Join(t.TempDir(), "counted")
			if err := native.Build(code, counted, native.Options{Count: true}); err != nil {
				t.Fatal(err)
			}
			counts := executeWith(t, nil, counted)
			if report := leakcheck.Unbalanced(leakRun(counts)); report != "" {
				t.Fatal(report)
			}
			if !bytes.Equal(counts.stdout, truth.stdout) {
				t.Fatal("counted output differs from Node")
			}
			t.Logf("Node-held, sanitized, shared leak check and counted balance pass:\n%s", counts.stderr)
		})
	}
	cases := []struct{ name, entry, file, from, to string }{
		{"checker callback publication", "main.a", "main.a", "checker.read = read;", "checker.read = (): string => 'unpublished';"},
		{"literal cache hit loses identity", "cache.a", "getOrUpdate.a", "return existing;", "return callback();"},
		{"cached undefined treated as missing", "cache.a", "getOrUpdate.a", "if (map.has(key))", "if (false)"},
	}
	for _, one := range cases {
		t.Run(one.name, func(t *testing.T) {
			directory := t.TempDir()
			for _, name := range []string{"main.a", "state.a", "cache.a", "getOrUpdate.a"} {
				data, err := os.ReadFile(filepath.Join(root, name))
				if err != nil {
					t.Fatal(err)
				}
				if name == one.file {
					if strings.Count(string(data), one.from) != 1 {
						t.Fatal("mutant site is not unique")
					}
					data = []byte(strings.Replace(string(data), one.from, one.to, 1))
				}
				if err := os.WriteFile(filepath.Join(directory, name), data, 0644); err != nil {
					t.Fatal(err)
				}
			}
			truth := onNode(t, filepath.Join(root, one.entry))
			mutantPath := filepath.Join(directory, one.entry)
			changedNode := onNode(t, mutantPath)
			if report := disagreement(truth, changedNode); report != "stdout differs" {
				t.Fatalf("source mutant caught by %q", report)
			}
			program, err := lowered(t, mutantPath)
			if err != nil {
				t.Fatal(err)
			}
			code := native.C(program)
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			got := executeWith(t, []string{leakSanitizer()}, binary)
			if report := disagreement(changedNode, got); report != "" {
				t.Fatalf("mutant native/source differ: %s", report)
			}
			if report := disagreement(truth, got); report != "stdout differs" {
				t.Fatalf("mutant caught by %q", report)
			}
			if report := leakChecked(t, code, binary); report != "" {
				t.Fatalf("semantic mutant leaked: %s", report)
			}
			t.Logf("mutant caught by original Node stdout; mutated Node/native agree and are leak-clean: %q", got.stdout)
		})
	}
	t.Run("extra cache value retain", func(t *testing.T) {
		path := filepath.Join(root, "cache.a")
		truth := onNode(t, path)
		program, err := lowered(t, path)
		if err != nil {
			t.Fatal(err)
		}
		original := native.C(program)
		changed := strings.ReplaceAll(original, "adamic_map_set(", "scout_leaky_map_set(")
		if changed == original {
			t.Fatal("ownership mutant changed nothing")
		}
		code := `#include "adamic.h"
static void scout_leaky_map_set(adamic_map *map, adamic_value key, adamic_value value) {
 if (map->reference_values && value.reference != NULL) { adamic_retain(value.reference); }
 adamic_map_set(map, key, value);
}
` + changed
		binary := filepath.Join(t.TempDir(), "leaky")
		if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
			t.Fatal(err)
		}
		got := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary)
		if report := disagreement(truth, got); report != "" {
			t.Fatalf("leak mutant changed semantics: %s", report)
		}
		report := leakChecked(t, code, binary)
		if report == "" {
			t.Fatal("extra retain escaped the shared leak check")
		}
		if runtime.GOOS == "linux" && !strings.Contains(report, "LeakSanitizer") {
			t.Fatalf("wrong leak failure: %s", report)
		}
		t.Logf("same Node bytes; shared check catches ownership mutant:\n%s", report)
		counted := filepath.Join(t.TempDir(), "leaky-counted")
		if err := native.Build(code, counted, native.Options{Count: true}); err != nil {
			t.Fatal(err)
		}
		counts := executeWith(t, nil, counted)
		if !bytes.Equal(counts.stdout, truth.stdout) {
			t.Fatal("counted mutant changed semantics")
		}
		report = leakcheck.Unbalanced(leakRun(counts))
		if report == "" {
			t.Fatal("extra retain escaped counted balance")
		}
		t.Logf("counted predicate catches ownership mutant: %s", report)
	})
}
