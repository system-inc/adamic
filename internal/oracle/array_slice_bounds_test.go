package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

func TestArraySliceBounds(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/array_slice_bounds.a"))
	if err != nil {
		test.Fatal(err)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatalf("Lower: %v", err)
	}
	want := []string{
		"absent-start:1,2,3", "absent-end:2,3", "both-absent:1,2,3",
		"undefined-start:1,2", "undefined-end:2,3", "both-undefined:1,2,3",
		"optional-passed:2", "optional-omitted:1,2,3", "maybe-local:1,2,3",
		"negative-one:3", "below-length:1,2,3", "past-length:",
		"fraction-positive:2,3", "fraction-negative:1,2,3",
		"end-negative:1,2", "end-below-length:", "end-past-length:1,2,3",
		"end-fraction-positive:1", "end-fraction-negative:",
		"nan-start:1,2,3", "nan-end:",
		"order-result:1,2,3", "order-log:receiver,start,end", "string-optional:abc/b",
	}
	node := onNode(test, path)
	native, leaked, comparable := nativelyCheckingLeaks(test, program, true)
	if leaked != "" {
		test.Errorf("leaks: %s", leaked)
	}
	if !comparable {
		return
	}
	for backend, got := range map[string]run{
		"node":       node,
		"native":     native,
		"javascript": onJavaScriptBackend(test, program),
	} {
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
		for i, line := range want {
			if lines[i] != line {
				test.Errorf("%s printed line %d: got %q, want %q", backend, i+1, lines[i], line)
			}
		}
		test.Logf("%s stdout:\n%s", backend, got.stdout)
	}
}

func TestArraySliceBoundsNullRefused(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/array_slice_refused/null_end.a"))
	if err != nil {
		test.Fatal(err)
	}
	truth := execute(
		test,
		"node",
		"--disable-warning=ExperimentalWarning",
		filepath.Join(repository, "oracle", "node.mjs"),
		path,
	)
	if truth.exitCode != 0 || string(truth.stdout) != "\n" || len(truth.stderr) != 0 {
		test.Fatalf("Node null end: %+v", truth)
	}
	_, err = load.Load([]string{path})
	var refused *load.CheckError
	const diagnostic = ":1:77: error TS2345: Argument of type 'number | null' is not assignable to parameter of type 'number | undefined'.\n  Type 'null' is not assignable to type 'number | undefined'."
	if !errors.As(err, &refused) || err.Error() != path+diagnostic {
		test.Fatalf("want null bound refused with %q, got %v", diagnostic, err)
	}
	test.Log(err)
}
