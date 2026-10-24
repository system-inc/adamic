package oracle

import (
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
	return buildcache.Product(t, buildcache.Inputs{Name: "program-tsc-node-api", Files: []string{"stage3/api/package-lock.json", "internal/oracle/program_region_node_records_test.go"}, Toolchain: []string{buildcache.Tool("tsc", "--version"), buildcache.Tool("node", "--version")}}, func(directory string) error {
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
		tsc, err := exec.LookPath("tsc")
		if err != nil {
			return err
		}
		tsc, err = filepath.EvalSymlinks(tsc)
		if err != nil {
			return err
		}
		library := filepath.Join(filepath.Dir(tsc), "../lib/typescript.js")
		version, err := exec.Command("node", "-e", "process.stdout.write(require(process.argv[1]).version)", library).Output()
		if err != nil {
			return fmt.Errorf("gate TypeScript API version: %w", err)
		}
		if string(version) != api.Version {
			return fmt.Errorf("gate TypeScript API version %q, lock requires %s", version, api.Version)
		}
		// npm's integrity hashes the compressed package archive, not typescript.js. The
		// traced product records the installed file's bytes instead, after checking its version.
		data, err = os.ReadFile(library)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Join(directory, "lib"), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(directory, "lib/typescript.js"), data, 0644); err != nil {
			return err
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
