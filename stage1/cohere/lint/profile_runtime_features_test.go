package lint

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Use an existing oracle fixture so source Node decides the closure call convention.
func TestProfileRuntimeFeaturesMatchNode(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/arguments_length_value_count.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(lowered)
	if !strings.Contains(source, "#define ADAMIC_CLOSURE_CONVENTION 1\n") {
		t.Fatal("fixture does not enable closure convention")
	}
	binary := filepath.Join(t.TempDir(), "profiled")
	if err := compilationProfileBuild(source, binary, native.Options{}); err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	want := execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, path).output
	got := execute(t, "", binary).output
	if !bytes.Equal(got, want) {
		t.Fatalf("profile output %q; Node %q", got, want)
	}
	if len(want) == 0 {
		t.Fatal("empty Node control")
	}
	t.Logf("profile and Node agree: %q", got)
}
