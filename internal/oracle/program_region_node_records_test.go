package oracle

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

func programNodeAPI(t *testing.T) string {
	t.Helper()
	return buildcache.Product(t, buildcache.Inputs{Name: "program-tsc-node-api", Files: []string{"stage3/api/package-lock.json", "internal/oracle/program_region_node_records_test.go"}}, func(directory string) error {
		data, err := os.ReadFile("../../stage3/api/package-lock.json")
		if err != nil {
			return err
		}
		var lock struct {
			Packages map[string]struct{ Version, Resolved, Integrity string }
		}
		if err := json.Unmarshal(data, &lock); err != nil {
			return err
		}
		api := lock.Packages["node_modules/typescript"]
		if api.Version != "6.0.3" {
			return fmt.Errorf("Node witness requires TypeScript 6.0.3, got %s", api.Version)
		}
		archive := filepath.Join(t.TempDir(), "typescript.tgz")
		out, err := exec.Command("curl", "--fail", "--location", "--silent", "--show-error", "--max-time", "35", api.Resolved, "-o", archive).CombinedOutput()
		if err != nil {
			return fmt.Errorf("fetch pinned Node API: %w: %s", err, out)
		}
		data, err = os.ReadFile(archive)
		if err != nil {
			return err
		}
		digest := sha512.Sum512(data)
		if "sha512-"+base64.StdEncoding.EncodeToString(digest[:]) != api.Integrity {
			return fmt.Errorf("TypeScript Node API integrity mismatch")
		}
		out, err = exec.Command("tar", "-xzf", archive, "-C", directory, "--strip-components=1", "package/lib/typescript.js").CombinedOutput()
		if err != nil {
			return fmt.Errorf("extract pinned Node API: %w: %s", err, out)
		}
		return nil
	})
}
func TestProduct_ProgramTscNodeAPI(t *testing.T) { t.Parallel(); programNodeAPI(t) }
func TestProgramRegionScalarContainersOnNode(t *testing.T) {
	t.Parallel()
	api := filepath.Join(programNodeAPI(t), "lib/typescript.js")
	result := execute(t, "node", "testdata/program_region/node_records.cjs", api)
	if result.exitCode != 0 || !strings.Contains(string(result.stdout), "K60 ") || strings.Count(string(result.stdout), "K144 ") != 3 {
		t.Fatalf("scalar container Node witness: %d %s %s", result.exitCode, result.stdout, result.stderr)
	}
	t.Log(string(result.stdout))
}
