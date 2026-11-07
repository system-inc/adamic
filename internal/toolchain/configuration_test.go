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
// that can't be read is refused, naming it. Mutants: comments read as words; the include's seen set dropped (the cycle
// never ends); a quoted path left with its quotes, so it names no file.
func TestAConfigsLinesAreWhatItAndItsFilesHold(t *testing.T) {
	t.Parallel()
	directory, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for name, contents := range map[string]string{
		"target.cfg":  "# --ld-path=<CFGDIR>/commented\n--ld-path=<CFGDIR>/ld.lld \\\n  \"<CFGDIR>/quoted file\" @back.cfg\n# --sysroot=<CFGDIR>\n",
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
	if report, err := ConfigurationsByContent("clang version 20.1.8\nConfiguration file: " + filepath.Join(directory, "absent.cfg")); err == nil || !strings.Contains(err.Error(), "reading clang's config "+filepath.Join(directory, "absent.cfg")) {
		t.Errorf("an absent config reported %q, %v; want it refused by its path", report, err)
	}
}

// A config naming a directory is refused by the flag it names it with, joined or with the directory after it, until
// wave 2b keys the directory by its archive's digest (#5xsd2ha): nothing keys what a sysroot, a -B prefix or a GCC
// installation holds yet. A config that includes one is refused too. Mutants: directoryFlag finding nothing; a flag
// read only joined to its directory.
func TestAConfigNamingADirectoryIsRefusedByName(t *testing.T) {
	t.Parallel()
	for contents, flag := range map[string]string{
		"--sysroot=<CFGDIR>/sysroot\n":        "--sysroot=",
		"--sysroot <CFGDIR>/sysroot\n":        "--sysroot=",
		"-B<CFGDIR>/bin\n":                    "-B",
		"-B <CFGDIR>/bin\n":                   "-B",
		"--prefix=<CFGDIR>/bin\n":             "--prefix=",
		"--gcc-toolchain=/usr\n":              "--gcc-toolchain=",
		"-isysroot <CFGDIR>/sdk\n":            "-isysroot",
		"-fno-plt\n@includes-a-sysroot.cfg\n": "--sysroot=",
	} {
		directory := t.TempDir()
		for name, text := range map[string]string{"target.cfg": contents, "includes-a-sysroot.cfg": "--sysroot=/\n"} {
			if err := os.WriteFile(filepath.Join(directory, name), []byte(text), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		want := "a clang config naming a directory, " + flag + ", isn't keyed until its archive digest is pinned (#5xsd2ha)"
		if lines, err := ConfigurationFile(filepath.Join(directory, "target.cfg")); err == nil || err.Error() != want {
			t.Errorf("a config of %q gave %q, %v; want %q", contents, lines, err, want)
		}
	}
}
