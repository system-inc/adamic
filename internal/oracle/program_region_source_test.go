package oracle

import (
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	"os/exec"
	"testing"
)

func programRegionSource(t *testing.T) string {
	t.Helper()
	inputs := buildcache.Inputs{Name: "program-tsc-source", Files: []string{"internal/oracle/testdata/program_region/source.sh", "internal/oracle/testdata/program_region/parse_manifest.json"}}
	build := func(directory string) error {
		out, err := exec.Command("bash", "testdata/program_region/source.sh", directory).CombinedOutput()
		if err != nil {
			return fmt.Errorf("pinned TypeScript source: %w: %s", err, out)
		}
		return nil
	}
	return buildcache.Product(t, inputs, build)
}
func TestProduct_ProgramTscSource(t *testing.T) { t.Parallel(); programRegionSource(t) }
