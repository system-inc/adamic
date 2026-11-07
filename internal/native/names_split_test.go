package native_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/compilerinputs"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func stableSplitProduct(test *testing.T, name string) string {
	test.Helper()
	path := "internal/oracle/testdata/" + name + ".a"
	inputs := buildcache.Inputs{
		Name:      "stable identity split " + name,
		Files:     append(compilerinputs.Lowering(), "internal/native/names_split_test.go", "oracle/node.mjs", path),
		Flags:     native.Flags(native.Options{Split: true, Sanitize: true, Count: true, Jobs: 2}),
		Toolchain: []string{buildcache.Tool("clang", "--version")},
	}
	return buildcache.Product(test, inputs, func(directory string) error {
		absolute, absolutePathError := filepath.Abs(filepath.Join("../..", path))
		if absolutePathError != nil {
			return absolutePathError
		}
		checked, loadError := load.Load([]string{absolute})
		if loadError != nil {
			return loadError
		}
		program, lowerError := lower.Lower(context.Background(), checked)
		if lowerError != nil {
			return lowerError
		}
		if name == "async_methods" {
			found := false
			for _, function := range program.Functions {
				if function.Name == "Worker_work" {
					found = true
					if function.Source.Module != "async_methods.a" ||
						len(function.Source.Declaration) != 2 ||
						function.Source.Declaration[0] != "Worker" ||
						function.Source.Declaration[1] != "work" {
						test.Fatal("async method lost source identity")
					}
				}
			}
			if !found {
				test.Fatal("missing async method identity witness")
			}
		} else {
			found := false
			for _, local := range program.Locals {
				if local.Global && local.Name == "alias" {
					found = true
					if local.Source.Module != "entries_record_alias.a" || len(local.Source.Declaration) == 0 {
						test.Fatal("record alias lost source identity")
					}
				}
			}
			if !found {
				test.Fatal("missing record alias identity witness")
			}
		}
		code := native.C(program)
		if writeError := os.WriteFile(filepath.Join(directory, "program.c"), []byte(code), 0600); writeError != nil {
			return writeError
		}
		if writeError := os.WriteFile(
			filepath.Join(directory, "program.mjs"),
			[]byte(javascript.JavaScript(program)),
			0600,
		); writeError != nil {
			return writeError
		}
		return native.Build(
			code,
			filepath.Join(directory, "program"),
			native.Options{Split: true, Sanitize: true, Count: true, Jobs: 2},
		)
	})
}

func stableSplitWitness(test *testing.T, name string) {
	test.Helper()
	directory := stableSplitProduct(test, name)
	run := func(command string, arguments ...string) []byte {
		test.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, command, arguments...)
		cmd.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
		output, outputError := cmd.Output()
		if outputError != nil {
			test.Fatalf("%s: %v", command, outputError)
		}
		return output
	}
	entry, absolutePathError := filepath.Abs("../oracle/testdata/" + name + ".a")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	truth := run("node", "--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", entry)
	for _, actual := range [][]byte{run(
		"node",
		"--disable-warning=ExperimentalWarning",
		"../../oracle/node.mjs",
		filepath.Join(directory, "program.mjs"),
	), run(filepath.Join(directory, "program"))} {
		if string(actual) != string(truth) {
			test.Fatalf("split/backend output %q, Node %q", actual, truth)
		}
	}
	code, readError := os.ReadFile(filepath.Join(directory, "program.c"))
	if readError != nil {
		test.Fatal(readError)
	}
	if report := leakcheck.Report(test, string(code), filepath.Join(directory, "program")); report != "" {
		test.Fatal(report)
	}
	test.Logf("%s: split, Node, JavaScript, ASan/UBSan and leak check passed", name)
}
func TestProduct_StableIdentityAsyncMethods(test *testing.T) {
	test.Parallel()
	stableSplitProduct(test, "async_methods")
}
func TestProduct_StableIdentityEntriesRecordAlias(test *testing.T) {
	test.Parallel()
	stableSplitProduct(test, "entries_record_alias")
}
func TestStableIdentityAsyncMethodsSplit(test *testing.T) {
	test.Parallel()
	stableSplitWitness(test, "async_methods")
}
func TestStableIdentityEntriesRecordAliasSplit(test *testing.T) {
	test.Parallel()
	stableSplitWitness(test, "entries_record_alias")
}
