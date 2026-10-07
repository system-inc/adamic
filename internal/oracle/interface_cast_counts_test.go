package oracle

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// The flag stays in a child compiler process so parallel default-off fixture runs cannot
// observe it. These rows use the same counted runtime and table as ordinary fixtures.
func interfaceCastCounts(t *testing.T) []string {
	t.Helper()
	rows := []string{}
	for _, path := range []string{"stage3/interface-downcasts/visitor.a", "stage3/interface-downcasts/wrong-kind.a"} {
		command := exec.Command("go", "run", "./cmd/adamic", "c", path)
		command.Dir = repository
		for _, variable := range os.Environ() {
			if !strings.HasPrefix(variable, "ADAMIC_INTERFACE_DOWNCASTS=") {
				command.Env = append(command.Env, variable)
			}
		}
		command.Env = append(command.Env, "ADAMIC_INTERFACE_DOWNCASTS=1")
		source, err := command.Output()
		if err != nil {
			t.Fatalf("flagged compiler for %s: %v", path, err)
		}
		binary := filepath.Join(t.TempDir(), "counted")
		if err := native.Build(string(source), binary, native.Options{Count: true}); err != nil {
			t.Fatal(err)
		}
		name, arguments := pinnedStack(binary)
		result := execute(t, name, arguments...)
		wantedExit := 0
		if strings.HasSuffix(path, "wrong-kind.a") {
			wantedExit = 70
		}
		if result.exitCode != wantedExit {
			t.Fatalf("%s counted exit %d stderr %q", path, result.exitCode, result.stderr)
		}
		match := countsLine.FindSubmatch(result.stderr)
		if match == nil {
			t.Fatalf("%s has no runtime counts: %q", path, result.stderr)
		}
		rows = append(rows, fmt.Sprintf("| %s | %s | %s | %s | %s | %s | %s |", path, match[1], match[2], match[3], match[4], match[5], match[6]))
	}
	return rows
}
