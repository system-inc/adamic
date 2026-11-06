package toolchain

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// A config's lines name no path, and each file it names by an absolute path, quoted or after an option's =, is keyed
// once by its bytes, as is a config it includes; a word in a comment names nothing, and an include cycle ends. A config
// that can't be read is named by its error. Mutants: comments read as words; the include's seen set dropped (the cycle
// never ends); a quoted path left with its quotes, so it names no file.
func TestAConfigsLinesAreWhatItAndItsFilesHold(t *testing.T) {
	t.Parallel()
	directory, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for name, contents := range map[string]string{
		"target.cfg":  "# --ld-path=<CFGDIR>/commented\n--ld-path=<CFGDIR>/ld.lld \\\n  \"<CFGDIR>/quoted file\" @back.cfg\n--sysroot=<CFGDIR>\n",
		"back.cfg":    "@target.cfg\n-fuse-ld=lld\n",
		"ld.lld":      "lld\n",
		"quoted file": "quoted\n",
		"commented":   "commented\n",
	} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	lines, err := ConfigurationFile(filepath.Join(directory, "target.cfg"))
	if err != nil {
		t.Fatal(err)
	}
	kinds := []string{}
	for _, line := range lines {
		kind, _, _ := strings.Cut(line, "sha256 ")
		kinds = append(kinds, kind)
		if strings.Contains(line, directory) {
			t.Errorf("a line names where the config sits: %q", line)
		}
	}
	want := []string{"Configuration file: ", "Configuration names: ", "Configuration names: ", "Configuration file: "}
	if !slices.Equal(kinds, want) {
		t.Errorf("the config's lines are %q, want the config, ld.lld, the quoted file and back.cfg:\n%s", kinds, strings.Join(lines, "\n"))
	}
	report := ConfigurationsByContent("clang version 20.1.8\nConfiguration file: " + filepath.Join(directory, "absent.cfg"))
	if !strings.Contains(report, "Configuration file: unreadable (") {
		t.Errorf("an absent config reported %q", report)
	}
}
