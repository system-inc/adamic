package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/rest_callable_generic.a",
		"internal/oracle/testdata/rest_callable_write.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}

// The generic not wrapper captures a callable that can reach its new closure.
// Keep the existing cycle refusal rather than erase the ownership contract.
func TestRestNotKeepsCycleRefusal(t *testing.T) {
	t.Parallel()
	path, pathErr := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/rest_callable_not.a"))
	if pathErr != nil {
		t.Fatal(pathErr)
	}
	source := onNode(t, path)
	if source.exitCode != 0 || string(source.stdout) != "true true false\nfalse\n2 3\n" || len(source.stderr) != 0 {
		t.Fatalf("source Node: %+v", source)
	}
	_, err := lowered(t, path)
	var refused *lower.Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "cycle reference counting can't free") {
		t.Fatalf("want the callable capture cycle refusal, got %v", err)
	}
}

func TestRestCallableViewsNeedAdapters(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ file, reason string }{
		{"rest_callable_view.a", "one keeps something weakly that the other keeps strongly"},
		{"rest_callable_context.a", "one keeps something weakly that the other keeps strongly"},
	} {
		t.Run(probe.file, func(t *testing.T) {
			path, pathErr := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", probe.file))
			if pathErr != nil {
				t.Fatal(pathErr)
			}
			source := onNode(t, path)
			if source.exitCode != 0 || string(source.stdout) != "1\n" || len(source.stderr) != 0 {
				t.Fatalf("source Node: %+v", source)
			}
			_, err := lowered(t, path)
			var stopped *lower.NotYet
			if !errors.As(err, &stopped) || !strings.Contains(err.Error(), probe.reason) {
				t.Fatalf("want an adapter boundary, got %v", err)
			}
		})
	}
}
