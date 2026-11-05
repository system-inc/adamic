package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

func TestArraySliceBounds(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/array_slice_bounds.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatalf("Lower: %v", err)
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
	node := onNode(t, path)
	native, leaked, comparable := nativelyCheckingLeaks(t, program, true)
	if leaked != "" {
		t.Errorf("leaks: %s", leaked)
	}
	if !comparable {
		return
	}
	for backend, got := range map[string]run{"node": node, "native": native, "javascript": onJavaScriptBackend(t, program)} {
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

func TestArraySliceBoundsNullRefused(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/array_slice_refused/null_end.a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle", "node.mjs"), path)
	if truth.exitCode != 0 || string(truth.stdout) != "\n" || len(truth.stderr) != 0 {
		t.Fatalf("Node null end: %+v", truth)
	}
	_, err = load.Load([]string{path})
	var refused *load.CheckError
	const diagnostic = ":1:77: error TS2345: Argument of type 'number | null' is not assignable to parameter of type 'number | undefined'.\n  Type 'null' is not assignable to type 'number | undefined'."
	if !errors.As(err, &refused) || err.Error() != path+diagnostic {
		t.Fatalf("want null bound refused with %q, got %v", diagnostic, err)
	}
	t.Log(err)
}
