package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"path/filepath"
	"testing"
)

func init() {
	for _, name := range []string{"accessor", "union_logical", "scanner", "typed_array", "array", "field", "local", "index_once", "receiver_once", "logical", "read_missing", "array_missing", "typed_array_missing"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/non_null_write_" + name + ".a", true, name == "read_missing" || name == "array_missing" || name == "typed_array_missing"})
	}
}

func TestNonNullWriteReadHalfMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/non_null_write_read_missing.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := run{stdout: []byte("before\n"), stderr: []byte("adamic: panic: non-null assertion failed: box.value! is null or undefined\n"), exitCode: 70}
	baseline, _ := nativelyUncached(t, program)
	if difference := disagreement(want, baseline); difference != "" {
		t.Fatal(difference)
	}
	if difference := disagreement(want, onJavaScriptBackend(t, program)); difference != "" {
		t.Fatal(difference)
	}
	changes := 0
	program.Main = mutateReadiness(program.Main, func(node any) any {
		if check, ok := node.(ir.Coalesce); ok && check.Panic != nil {
			check.Panic = nil
			check.Fallback = ir.NumberConstant{}
			changes++
			return check
		}
		return node
	})
	if changes != 1 {
		t.Fatalf("read-half mutant changed %d checks", changes)
	}
	mutant, _ := nativelyUncached(t, program)
	if mutant.exitCode != 0 || disagreement(want, mutant) == "" {
		t.Fatalf("read-half check mutant escaped pinned output: %#v", mutant)
	}
	t.Logf("read-half check removed: exit %d stdout %q, caught by pinned panic", mutant.exitCode, mutant.stdout)
}

func TestNonNullWriteIndexOnceMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/non_null_write_index_once.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	baseline, _ := nativelyUncached(t, program)
	if difference := disagreement(want, baseline); difference != "" {
		t.Fatal(difference)
	}
	indices := map[int]ir.Expression{}
	program.Main = mutateReadiness(program.Main, func(node any) any {
		if declaration, ok := node.(ir.Declare); ok && program.Locals[declaration.Local].Name == "assertion_index" {
			indices[declaration.Local] = declaration.Value
		}
		return node
	})
	changes := 0
	program.Main = mutateReadiness(program.Main, func(node any) any {
		if store, ok := node.(ir.SetIndex); ok {
			if read, ok := store.Index.(ir.Read); ok && indices[read.Local] != nil {
				store.Index = indices[read.Local]
				changes++
				return store
			}
		}
		return node
	})
	if changes != 1 {
		t.Fatalf("index-twice mutant changed %d stores", changes)
	}
	mutant, _ := nativelyUncached(t, program)
	if mutant.exitCode != 0 || disagreement(want, mutant) != "stdout differs" {
		t.Fatalf("index-twice mutant escaped Node output: %#v", mutant)
	}
	t.Logf("index evaluated twice: exit %d stdout %q versus Node %q", mutant.exitCode, mutant.stdout, want.stdout)
}
