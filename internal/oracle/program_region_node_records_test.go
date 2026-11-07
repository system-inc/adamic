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

func programNodeAPI(test *testing.T) string {
	test.Helper()
	return buildcache.Product(test, buildcache.Inputs{
		Name:      "program-tsc-node-api",
		Files:     []string{"stage3/api/package-lock.json", "internal/oracle/program_region_node_records_test.go"},
		Toolchain: []string{buildcache.Tool("tsc", "--version"), buildcache.Tool("node", "--version")},
	}, func(directory string) error {
		data, readError := os.ReadFile("../../stage3/api/package-lock.json")
		if readError != nil {
			return readError
		}
		var lock struct {
			Packages map[string]struct{ Version, Resolved, Integrity string }
		}
		if decodeError := json.Unmarshal(data, &lock); decodeError != nil {
			return decodeError
		}
		api := lock.Packages["node_modules/typescript"]
		if api.Version != "6.0.3" {
			return fmt.Errorf("Node witness requires TypeScript 6.0.3, got %s", api.Version)
		}
		tsc, lookPathError := exec.LookPath("tsc")
		if lookPathError != nil {
			return lookPathError
		}
		tsc, resolveSymlinksError := filepath.EvalSymlinks(tsc)
		if resolveSymlinksError != nil {
			return resolveSymlinksError
		}
		library := filepath.Join(filepath.Dir(tsc), "../lib/typescript.js")
		version, outputError := exec.Command(
			"node",
			"-e",
			"process.stdout.write(require(process.argv[1]).version)",
			library,
		).Output()
		if outputError != nil {
			return fmt.Errorf("gate TypeScript API version: %w", outputError)
		}
		if string(version) != api.Version {
			return fmt.Errorf("gate TypeScript API version %q, lock requires %s", version, api.Version)
		}
		// npm's integrity hashes the compressed package archive, not typescript.js. The
		// traced product records the installed file's bytes instead, after checking its version.
		data, readError = os.ReadFile(library)
		if readError != nil {
			return readError
		}
		if createDirectoryError := os.MkdirAll(filepath.Join(directory, "lib"), 0755); createDirectoryError != nil {
			return createDirectoryError
		}
		if writeError := os.WriteFile(filepath.Join(directory, "lib/typescript.js"), data, 0644); writeError != nil {
			return writeError
		}
		return nil
	})
}
func TestProduct_ProgramTscNodeAPI(test *testing.T) { test.Parallel(); programNodeAPI(test) }
func TestProgramRegionScalarContainersOnNode(test *testing.T) {
	test.Parallel()
	api := filepath.Join(programNodeAPI(test), "lib/typescript.js")
	result := execute(test, "node", "testdata/program_region/node_records.cjs", api)
	if result.exitCode != 0 ||
		!strings.Contains(string(result.stdout), "K60 ") ||
		strings.Count(string(result.stdout), "K144 ") != 3 {
		test.Fatalf("scalar container Node witness: %d %s %s", result.exitCode, result.stdout, result.stderr)
	}
	test.Log(string(result.stdout))
}
