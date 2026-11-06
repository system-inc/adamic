package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func TestArrayUnionRead(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/array_union.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatalf("Lower: %v", err)
	}
	want := []string{
		"em-is-array:true",
		"em-is-array:true",
		"em-flat:a",
		"em-flat:b",
		"em-first:object",
		"em-missing:true",
		"em-missing-is-array:false",
		"er-is-array:false",
		"er-is-array:true",
		"er-is-array:false",
		"er-flat:c",
		"er-flat:d",
		"er-first:object",
		"er-missing:true",
		"er-missing-is-array:false",
		"sm-is-array:true",
		"sm-is-array:true",
		"sm-flat:a",
		"sm-flat:b",
		"sm-first:object",
		"sm-first-is-c:false",
		"sm-missing:true",
		"sm-missing-is-array:false",
		"sr-is-array:false",
		"sr-is-array:true",
		"sr-is-array:false",
		"sr-flat:c",
		"sr-flat:d",
		"sr-first:string",
		"sr-first-is-c:true",
		"sr-missing:true",
		"sr-missing-is-array:false",
	}
	node := onNode(t, path)
	machine, leaked, comparable := nativelyCheckingLeaks(t, program, true)
	if leaked != "" {
		t.Errorf("leaks: %s", leaked)
	}
	if !comparable {
		return
	}
	backends := map[string]run{"node": node, "native": machine, "javascript": onJavaScriptBackend(t, program)}
	if os.Getenv("WASI_SYSROOT") != "" {
		backends["wasi"] = onWASI(t, native.C(program))
	} else {
		t.Log("WASI not run: WASI_SYSROOT is unset")
	}
	for backend, got := range backends {
		if got.exitCode != 0 || len(got.stderr) != 0 {
			t.Errorf("%s: exit %d stderr %q", backend, got.exitCode, got.stderr)
		}
		if string(got.stdout) != strings.Join(want, "\n")+"\n" {
			t.Errorf("%s stdout bytes differ from pinned output", backend)
		}
		lines := strings.Split(strings.TrimSuffix(string(got.stdout), "\n"), "\n")
		if len(lines) != len(want) {
			t.Errorf("%s: want %d lines, got %q", backend, len(want), got.stdout)
			continue
		}
		for i, line := range want {
			if lines[i] != line {
				t.Errorf("%s printed line %d: got %q, want %q", backend, i+1, lines[i], line)
			}
		}
		t.Logf("%s stdout:\n%s", backend, got.stdout)
	}
}

func TestArrayUnionWriteRefused(t *testing.T) {
	t.Parallel()
	for _, row := range []struct{ name, diagnostic string }{
		{"write", ":2:5: error TS2542: Index signature in type 'string[][] | readonly (string | readonly string[] | undefined)[]' only permits reading."},
		{"mutate", ":2:11: error TS2339: Property 'push' does not exist on type 'string[][] | readonly (string | readonly string[] | undefined)[]'.\n  Property 'push' does not exist on type 'readonly (string | readonly string[] | undefined)[]'."},
	} {
		t.Run(row.name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/array_union_refused", row.name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			_, err = load.Load([]string{path})
			var refused *load.CheckError
			if !errors.As(err, &refused) || err.Error() != path+row.diagnostic {
				t.Fatalf("want checker refusal %q, got %v", row.diagnostic, err)
			}
			t.Log(err)
		})
	}
}

func TestArrayUnionLayoutRefused(t *testing.T) {
	t.Parallel()
	for _, row := range []struct{ name, reason string }{
		{"mutable_write", "a value of type string[][] | (string | readonly string[] | undefined)[] where an array goes"},
		{"layout", "an array union read whose arm element number has a different slot representation from the union element"},
	} {
		t.Run(row.name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/array_union_refused", row.name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowered(t, path)
			var refused *lower.NotYet
			if !errors.As(err, &refused) || refused.What != row.reason {
				t.Fatalf("want lowering refusal %q, got %v", row.reason, err)
			}
			t.Log(err)
		})
	}
}

// Source Node observes the replacement array. Adamic's established narrowing
// contract stops before reading it as the class (TestNarrowedUnionMemberCheck).
func TestArrayUnionClassNarrowingInvalidated(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile(filepath.Join(repository, "internal/oracle/testdata/array_union.a"))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "array_union.a")
	source = append(source, []byte("\narrayUnionClassNarrowing();\n")...)
	if err := os.WriteFile(path, source, 0o644); err != nil {
		t.Fatal(err)
	}
	const observation = "er-class-read-is-array:true\n"
	node := onNode(t, path)
	if node.exitCode != 0 || len(node.stderr) != 0 || !strings.HasSuffix(string(node.stdout), observation) {
		t.Fatalf("source Node: exit %d stdout %q stderr %q", node.exitCode, node.stdout, node.stderr)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := run{stdout: []byte(strings.TrimSuffix(string(node.stdout), observation)), exitCode: 70,
		stderr: []byte("adamic: panic: union member where the checker narrowed it away: a call since the narrowing put it back\n")}
	machine, _ := natively(t, program)
	for backend, got := range map[string]run{"native": machine, "javascript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Errorf("%s: %s: exit %d stdout %q stderr %q", backend, difference, got.exitCode, got.stdout, got.stderr)
		}
	}
	t.Logf("source Node: %s; native and JavaScript must stop before that class read", strings.TrimSpace(observation))
}

// Assigning an array gives the checker an array view while the declared union
// still carries the layouts that the narrowed local read must refuse.
func TestArrayUnionNarrowingRefused(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"tuple_member", "object_member"} {
		t.Run(name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/array_union_refused", name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			if node := onNode(t, path); disagreement(run{stdout: []byte("true\n")}, node) != "" {
				t.Fatalf("source Node: exit %d stdout %q stderr %q", node.exitCode, node.stdout, node.stderr)
			}
			_, err = lowered(t, path)
			var refused *lower.NotYet
			const reason = "a narrowed union member whose object tag cannot be checked with typeof; keep differently held object kinds in separately typed variables"
			if !errors.As(err, &refused) || refused.What != reason || refused.Where != path+":5:19" {
				t.Fatalf("want narrowing refusal %q, got %v", reason, err)
			}
			t.Log(err)
		})
	}
}
