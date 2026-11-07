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

// A directory a config names is taken inside the compiler's own installation, the directory above the bin/ the config
// sits in, as wasi-sdk 27's bin/clang.cfg names `--sysroot=<CFGDIR>/../share/wasi-sysroot`: the release its version
// report names holds it, and the config's lines name no path. A directory outside it, `..` leading out, or a config that
// sits in no bin/, is refused by name as before. Mutants: insideInstallation true for every directory; `..` left
// unresolved, so <CFGDIR>/../../outside reads as inside; the next word left unread for a flag apart from its directory.
func TestAConfigMayNameADirectoryInsideItsOwnInstallation(t *testing.T) {
	t.Parallel()
	for contents, refused := range map[string]string{
		"--sysroot=<CFGDIR>/../share/wasi-sysroot\n": "",
		"-isysroot <CFGDIR>/../share/wasi-sysroot\n": "",
		"-B <CFGDIR>/../lib\n-fno-plt\n":             "",
		"--sysroot=/usr/lib\n":                       "--sysroot=",
		"--sysroot=<CFGDIR>/../../outside\n":         "--sysroot=",
		"--sysroot <CFGDIR>/../..\n":                 "--sysroot=",
		"-B <CFGDIR>/../../elsewhere\n":              "-B",
		"--gcc-toolchain=<CFGDIR>/../../../usr\n":    "--gcc-toolchain=",
		"--sysroot=relative/sysroot\n":               "--sysroot=",
	} {
		installation := filepath.Join(t.TempDir(), "wasi-sdk")
		config := filepath.Join(installation, "bin", "clang.cfg")
		if err := os.MkdirAll(filepath.Dir(config), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(config, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
		lines, err := ConfigurationFile(config)
		if refused == "" {
			if err != nil || len(lines) != 1 || !strings.HasPrefix(lines[0], "Configuration file: sha256 ") {
				t.Errorf("a config of %q inside its installation gave %q, %v; want its own line alone", contents, lines, err)
			}
			continue
		}
		want := "a clang config naming a directory, " + refused + ", isn't keyed until its archive digest is pinned (#5xsd2ha)"
		if err == nil || err.Error() != want {
			t.Errorf("a config of %q gave %q, %v; want %q", contents, lines, err, want)
		}
	}
	// A config in no bin/ has no installation of its own.
	elsewhere := filepath.Join(t.TempDir(), "wasi-sdk", "etc", "clang.cfg")
	if err := os.MkdirAll(filepath.Dir(elsewhere), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(elsewhere, []byte("--sysroot=<CFGDIR>/../share/wasi-sysroot\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if lines, err := ConfigurationFile(elsewhere); err == nil {
		t.Errorf("a config in no bin/ naming a directory gave %q, want it refused", lines)
	}
}
