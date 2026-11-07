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

func TestArrayUnionRead(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/array_union.a"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatalf("Lower: %v", loweredError)
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
	node := onNode(test, path)
	machine, leaked, comparable := nativelyCheckingLeaks(test, program, true)
	if leaked != "" {
		test.Errorf("leaks: %s", leaked)
	}
	if !comparable {
		return
	}
	backends := map[string]run{"node": node, "native": machine, "javascript": onJavaScriptBackend(test, program)}
	if os.Getenv("WASI_SYSROOT") != "" {
		backends["wasi"] = onWASI(test, native.C(program))
	} else {
		test.Log("WASI not run: WASI_SYSROOT is unset")
	}
	for backend, got := range backends {
		if got.exitCode != 0 || len(got.stderr) != 0 {
			test.Errorf("%s: exit %d stderr %q", backend, got.exitCode, got.stderr)
		}
		if string(got.stdout) != strings.Join(want, "\n")+"\n" {
			test.Errorf("%s stdout bytes differ from pinned output", backend)
		}
		lines := strings.Split(strings.TrimSuffix(string(got.stdout), "\n"), "\n")
		if len(lines) != len(want) {
			test.Errorf("%s: want %d lines, got %q", backend, len(want), got.stdout)
			continue
		}
		for index, line := range want {
			if lines[index] != line {
				test.Errorf("%s printed line %d: got %q, want %q", backend, index+1, lines[index], line)
			}
		}
		test.Logf("%s stdout:\n%s", backend, got.stdout)
	}
}

func TestArrayUnionWriteRefused(test *testing.T) {
	test.Parallel()
	for _, row := range []struct{ name, diagnostic string }{
		{
			"write",
			":2:5: error TS2542: Index signature in type 'string[][] | readonly (string | readonly string[] | undefined)[]' only permits reading.",
		},
		{
			"mutate",
			":2:11: error TS2339: Property 'push' does not exist on type 'string[][] | readonly (string | readonly string[] | undefined)[]'.\n  Property 'push' does not exist on type 'readonly (string | readonly string[] | undefined)[]'.",
		},
	} {
		test.Run(row.name, func(test *testing.T) {
			path, absolutePathError := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/array_union_refused", row.name+".a"),
			)
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			_, loadError := load.Load([]string{path})
			var refused *load.CheckError
			if !errors.As(loadError, &refused) || loadError.Error() != path+row.diagnostic {
				test.Fatalf("want checker refusal %q, got %v", row.diagnostic, loadError)
			}
			test.Log(loadError)
		})
	}
}

func TestArrayUnionLayoutRefused(test *testing.T) {
	test.Parallel()
	for _, row := range []struct{ name, reason string }{
		{
			"mutable_write",
			"a value of type string[][] | (string | readonly string[] | undefined)[] where an array goes",
		},
		{
			"layout",
			"an array union read whose arm element number has a different slot representation from the union element",
		},
	} {
		test.Run(row.name, func(test *testing.T) {
			path, absolutePathError := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/array_union_refused", row.name+".a"),
			)
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			_, loweredError := lowered(test, path)
			var refused *lower.NotYet
			if !errors.As(loweredError, &refused) || refused.What != row.reason {
				test.Fatalf("want lowering refusal %q, got %v", row.reason, loweredError)
			}
			test.Log(loweredError)
		})
	}
}

// Source Node observes the replacement array. Adamic's established narrowing
// contract stops before reading it as the class (TestNarrowedUnionMemberCheck).
func TestArrayUnionClassNarrowingInvalidated(test *testing.T) {
	test.Parallel()
	source, readError := os.ReadFile(filepath.Join(repository, "internal/oracle/testdata/array_union.a"))
	if readError != nil {
		test.Fatal(readError)
	}
	path := filepath.Join(test.TempDir(), "array_union.a")
	source = append(source, []byte("\narrayUnionClassNarrowing();\n")...)
	if writeError := os.WriteFile(path, source, 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	const observation = "er-class-read-is-array:true\n"
	node := onNode(test, path)
	if node.exitCode != 0 || len(node.stderr) != 0 || !strings.HasSuffix(string(node.stdout), observation) {
		test.Fatalf("source Node: exit %d stdout %q stderr %q", node.exitCode, node.stdout, node.stderr)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	want := run{stdout: []byte(strings.TrimSuffix(string(node.stdout), observation)), exitCode: 70,
		stderr: []byte(
			"adamic: panic: union member where the checker narrowed it away: a call since the narrowing put it back\n",
		)}
	machine, _ := natively(test, program)
	for backend, got := range map[string]run{"native": machine, "javascript": onJavaScriptBackend(test, program)} {
		if difference := disagreement(want, got); difference != "" {
			test.Errorf(
				"%s: %s: exit %d stdout %q stderr %q",
				backend,
				difference,
				got.exitCode,
				got.stdout,
				got.stderr,
			)
		}
	}
	test.Logf("source Node: %s; native and JavaScript must stop before that class read", strings.TrimSpace(observation))
}

// Assigning an array gives the checker an array view while the declared union
// still carries the layouts that the narrowed local read must refuse.
func TestArrayUnionNarrowingRefused(test *testing.T) {
	test.Parallel()
	for _, name := range []string{"tuple_member", "object_member"} {
		test.Run(name, func(test *testing.T) {
			path, absolutePathError := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/array_union_refused", name+".a"),
			)
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			if node := onNode(test, path); disagreement(run{stdout: []byte("true\n")}, node) != "" {
				test.Fatalf("source Node: exit %d stdout %q stderr %q", node.exitCode, node.stdout, node.stderr)
			}
			_, loweredError := lowered(test, path)
			var refused *lower.NotYet
			const reason = "a narrowed union member whose object tag cannot be checked with typeof; keep differently held object kinds in separately typed variables"
			if !errors.As(loweredError, &refused) || refused.What != reason || refused.Where != path+":5:19" {
				test.Fatalf("want narrowing refusal %q, got %v", reason, loweredError)
			}
			test.Log(loweredError)
		})
	}
}
