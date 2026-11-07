package buildcache

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/toolchain"
)

// compilerReportBuild is a real clang and a build platform whose triple isn't that clang's default, linux on the arch
// this host isn't, with the names that triple's config takes: as the table spells it and as clang normalizes it, since
// clang searches by the triple it computes. Only a report read under the build's triple finds a config by either name;
// a bare --version looks for this host's. It skips where there's no clang, or the host isn't linux/amd64 or arm64.
func compilerReportBuild(t *testing.T) (clang string, environment []string, configs []string) {
	t.Helper()
	other := map[string]string{"amd64": "arm64", "arm64": "amd64"}[runtime.GOARCH]
	if runtime.GOOS != "linux" || other == "" {
		t.Skipf("a cross triple's config is probed from linux/amd64 or linux/arm64, and this is %s/%s", runtime.GOOS, runtime.GOARCH)
	}
	found, err := toolchain.Path("clang")
	if err != nil {
		t.Skipf("no clang: %v", err)
	}
	clang, err = filepath.EvalSymlinks(found)
	if err != nil {
		t.Fatal(err)
	}
	triple, err := toolchain.Triple("linux", other)
	if err != nil {
		t.Fatal(err)
	}
	architecture, system, _ := strings.Cut(triple, "-")
	return clang, []string{"GOOS=linux", "GOARCH=" + other}, []string{triple + ".cfg", architecture + "-unknown-" + system + ".cfg"}
}

// compilerReportMoves reads a compiler report for environment's build with clang and extra arguments before
// CompilerReport's, once with the configs in directory holding one define and once another, and fails unless the
// report read a config and moved with the edit.
func compilerReportMoves(t *testing.T, clang string, extra []string, environment []string, directory string, configs []string) {
	t.Helper()
	report := func(define string) string {
		t.Helper()
		for _, config := range configs {
			write(t, directory, config, "-DADAMIC_CONFIG_PROBE="+define+"\n")
		}
		// toolValue, not Tool: Tool runs a command once per process, and the edit has to be read again.
		arguments := append(append([]string{}, extra...), compilerReportArguments(environment)...)
		return toolValue(clang, "clang "+strings.Join(arguments, " "), arguments, toolchain.Env())
	}
	before, after := report("1"), report("2")
	if !strings.Contains(before, "Configuration file: sha256 ") {
		t.Errorf("the report read no config of %q in %s:\n%s", configs, directory, before)
	}
	if before == after {
		t.Errorf("an edited config of %q moved no report:\n%s", configs, before)
	}
}

// A compiler's report for a key is read under the build's triple (Node's must on #5xsd2ha): clang picks its config
// files by target, so a config for the triple a cross build compiles for, beside clang, moves the key when it's edited.
// clang searches beside the program it resolves itself to, so a link to clang wouldn't do: clang is hard linked
// beside the config, or copied where the two file systems differ. Mutant: compilerReportArguments without --target,
// which reads this host's triple's configs, and the report reads none here.
func TestACompilerReportReadsTheBuildsTriplesConfigBesideClang(t *testing.T) {
	t.Parallel()
	clang, environment, configs := compilerReportBuild(t)
	directory := t.TempDir()
	beside := filepath.Join(directory, "clang")
	if err := os.Link(clang, beside); err != nil {
		copyExecutable(t, clang, beside)
	}
	compilerReportMoves(t, beside, nil, environment, directory, configs)
}

// The same as TestACompilerReportReadsTheBuildsTriplesConfigBesideClang, with the config in a directory clang is told
// is its system config directory (--config-system-dir), so the real clang is read where it's installed. Mutant: as
// that test's.
func TestACompilerReportReadsTheBuildsTriplesConfigFromItsSystemDirectory(t *testing.T) {
	t.Parallel()
	clang, environment, configs := compilerReportBuild(t)
	directory := t.TempDir()
	compilerReportMoves(t, clang, []string{"--config-system-dir=" + directory}, environment, directory, configs)
}

// copyExecutable copies the program at from to to, executable.
func copyExecutable(t *testing.T, from, to string) {
	t.Helper()
	source, err := os.Open(from)
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	destination, err := os.OpenFile(to, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o755)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(destination, source); err != nil {
		destination.Close()
		t.Fatal(err)
	}
	if err := destination.Close(); err != nil {
		t.Fatal(err)
	}
}
