package worker

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func TestWasmPurityRefusesHostOperations(t *testing.T) {
	t.Parallel()
	builds := 0
	err := BuildWith("testdata/wasm/host_cwd.a", filepath.Join(t.TempDir(), "worker"), Options{
		Wasm:      []string{"hostDirectory"},
		BuildWasm: func(string, string, string, []string) error { builds++; return fmt.Errorf("unexpected Wasm build") },
	})
	for _, want := range []string{"host_cwd.a:3:", "hostDirectory", "impure: host operation host_cwd", "fix:"} {
		if err == nil || !strings.Contains(err.Error(), want) || builds != 0 {
			t.Fatalf("expected located host refusal before Wasm build (%s), builds=%d: %v", want, builds, err)
		}
	}
}
