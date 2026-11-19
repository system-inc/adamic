package native

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/native_runtime"
)

// Count is checked in the linked runtime itself, not merely in a comparison of keys. Build an
// uncounted library first: a key missing ADAMIC_COUNT would reuse it and silently lose the report.
func TestRuntimeCacheKeepsCountFlags(t *testing.T) {
	t.Parallel()
	const source = "int main(void) { return 0; }\n"
	for _, count := range []bool{false, true, false} {
		binary := filepath.Join(t.TempDir(), "main")
		if err := Build(source, binary, native_runtime.Options{Count: count}); err != nil {
			t.Fatal(err)
		}
		output, err := exec.Command(binary).CombinedOutput()
		if err != nil {
			t.Fatalf("run: %v\n%s", err, output)
		}
		want := ""
		if count {
			want = "adamic: counts: allocations 0 frees 0 retains 0 releases 0 peak 0 regions 0\n"
		}
		if string(output) != want {
			t.Errorf("count %v: got %q, want %q", count, output, want)
		}
	}
}
