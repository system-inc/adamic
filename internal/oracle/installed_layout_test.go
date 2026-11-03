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
func TestInstalledLayoutIsNodes(t *testing.T) {
	t.Parallel()
	library := filepath.Join(t.TempDir(), "typescript", "lib")
	if err := os.MkdirAll(library, 0o755); err != nil {
		t.Fatal(err)
	}
	// The package's own package.json, without a type, as npm installs it: _tsc.js is CommonJS whatever
	// a directory above the temporary one says.
	if err := os.WriteFile(filepath.Join(library, "..", "package.json"), []byte(`{"name": "typescript", "version": "6.0.3"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	tsc := filepath.Join(library, "_tsc.js")
	script := `try { require("source-map-support"); console.log("installed"); } catch (e) { const lines = e.message.split("\n"); console.log([e.name, e.code, lines[0], lines[1], Object.keys(e).join(",")].join("\n")); }` + "\n"
	if err := os.WriteFile(tsc, []byte(script), 0o644); err != nil {
		t.Fatal(err)
	}
	// NODE_PATH would add a directory to where Node looks, which the installed layout doesn't have.
	measured := executeWith(t, []string{"NODE_PATH="}, "node", tsc)
	if measured.exitCode != 0 || len(measured.stderr) != 0 {
		t.Fatalf("node: exit %d, stdout %q, stderr %q", measured.exitCode, measured.stdout, measured.stderr)
	}
	lines := strings.Split(strings.TrimSuffix(string(measured.stdout), "\n"), "\n")
	want := []string{installedLayoutName, installedLayoutCode, installedLayoutMessage, "Require stack:", "code,requireStack"}
	if strings.Join(lines, "\n") != strings.Join(want, "\n") {
		t.Fatalf("Node's answer for an installed tsc: got %q, want %q", lines, want)
	}
}
