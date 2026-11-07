package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// installed-layout's answer is measured, not remembered: tsc 6.0.3's require("source-map-support"),
// from lib/_tsc.js in a typescript package with nothing installed beside it, run on the Node the
// oracle runs. The error's name, code and first line are the constants native's rows are pinned from
// (backend_divergences_test.go), so this fails if Node's answer ever changes. The rest is the
// declared difference: a Require stack tail and a requireStack naming the requiring file.
func TestInstalledLayoutIsNodes(test *testing.T) {
	test.Parallel()
	library := filepath.Join(test.TempDir(), "typescript", "lib")
	if createDirectoryError := os.MkdirAll(library, 0o755); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	// The package's own package.json, without a type, as npm installs it: _tsc.js is CommonJS whatever
	// a directory above the temporary one says.
	if writeError := os.WriteFile(
		filepath.Join(library, "..", "package.json"),
		[]byte(`{"name": "typescript", "version": "6.0.3"}`+"\n"),
		0o644,
	); writeError != nil {
		test.Fatal(writeError)
	}
	tsc := filepath.Join(library, "_tsc.js")
	script := `try { require("source-map-support"); console.log("installed"); } catch (e) { const lines = e.message.split("\n"); console.log([e.name, e.code, lines[0], lines[1], Object.keys(e).join(",")].join("\n")); }` +
		"\n"
	if writeError := os.WriteFile(tsc, []byte(script), 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	// NODE_PATH would add a directory to where Node looks, which the installed layout doesn't have.
	measured := executeWith(test, []string{"NODE_PATH="}, "node", tsc)
	if measured.exitCode != 0 || len(measured.stderr) != 0 {
		test.Fatalf("node: exit %d, stdout %q, stderr %q", measured.exitCode, measured.stdout, measured.stderr)
	}
	lines := strings.Split(strings.TrimSuffix(string(measured.stdout), "\n"), "\n")
	want := []string{
		installedLayoutName,
		installedLayoutCode,
		installedLayoutMessage,
		"Require stack:",
		"code,requireStack",
	}
	if strings.Join(lines, "\n") != strings.Join(want, "\n") {
		test.Fatalf("Node's answer for an installed tsc: got %q, want %q", lines, want)
	}
}
