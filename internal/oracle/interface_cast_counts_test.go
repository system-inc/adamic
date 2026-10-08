package oracle

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// defaultInterfaceCount builds with the default compiler, as the old interface sweep did.
func defaultInterfaceCount(t *testing.T, path, compiler string) string {
	t.Helper()
	command := exec.Command(compiler, "c", path)
	command.Dir = repository
	source, err := command.Output()
	if err != nil {
		t.Fatalf("default compiler for %s: %v", path, err)
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
	return fmt.Sprintf("| %s | %s | %s | %s | %s | %s | %s |", path, match[1], match[2], match[3], match[4], match[5], match[6])
}
