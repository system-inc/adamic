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

func stableSplitProduct(t *testing.T, name string) string {
	t.Helper()
	path := "internal/oracle/testdata/" + name + ".a"
	inputs := buildcache.Inputs{Name: "stable identity split " + name, Files: append(compilerinputs.Lowering(), "internal/native/names_split_test.go", "oracle/node.mjs", path), Flags: native.Flags(native.Options{Split: true, Sanitize: true, Count: true, Jobs: 2}), Toolchain: []string{buildcache.Tool("clang", "--version")}}
	return buildcache.Product(t, inputs, func(directory string) error {
		absolute, err := filepath.Abs(filepath.Join("../..", path))
		if err != nil {
			return err
		}
		checked, err := load.Load([]string{absolute})
		if err != nil {
			return err
		}
		program, err := lower.Lower(context.Background(), checked)
		if err != nil {
			return err
		}
		if name == "async_methods" {
			found := false
			for _, function := range program.Functions {
				if function.Name == "Worker_work" {
					found = true
					if function.Source.Module != "async_methods.a" || len(function.Source.Declaration) != 2 || function.Source.Declaration[0] != "Worker" || function.Source.Declaration[1] != "work" {
						t.Fatal("async method lost source identity")
					}
				}
			}
			if !found {
				t.Fatal("missing async method identity witness")
			}
		} else {
			found := false
			for _, local := range program.Locals {
				if local.Global && local.Name == "alias" {
					found = true
					if local.Source.Module != "entries_record_alias.a" || len(local.Source.Declaration) == 0 {
						t.Fatal("record alias lost source identity")
					}
				}
			}
			if !found {
				t.Fatal("missing record alias identity witness")
			}
		}
		code := native.C(program)
		if err := os.WriteFile(filepath.Join(directory, "program.c"), []byte(code), 0600); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(directory, "program.mjs"), []byte(javascript.JavaScript(program)), 0600); err != nil {
			return err
		}
		return native.Build(code, filepath.Join(directory, "program"), native.Options{Split: true, Sanitize: true, Count: true, Jobs: 2})
	})
}

func stableSplitWitness(t *testing.T, name string) {
	t.Helper()
	directory := stableSplitProduct(t, name)
	run := func(command string, arguments ...string) []byte {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, command, arguments...)
		cmd.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
		output, err := cmd.Output()
		if err != nil {
			t.Fatalf("%s: %v", command, err)
		}
		return output
	}
	entry, err := filepath.Abs("../oracle/testdata/" + name + ".a")
	if err != nil {
		t.Fatal(err)
	}
	truth := run("node", "--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", entry)
	for _, actual := range [][]byte{run("node", "--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", filepath.Join(directory, "program.mjs")), run(filepath.Join(directory, "program"))} {
		if string(actual) != string(truth) {
			t.Fatalf("split/backend output %q, Node %q", actual, truth)
		}
	}
	code, err := os.ReadFile(filepath.Join(directory, "program.c"))
	if err != nil {
		t.Fatal(err)
	}
	if report := leakcheck.Report(t, string(code), filepath.Join(directory, "program")); report != "" {
		t.Fatal(report)
	}
	t.Logf("%s: split, Node, JavaScript, ASan/UBSan and leak check passed", name)
}
func TestProduct_StableIdentityAsyncMethods(t *testing.T) {
	t.Parallel()
	stableSplitProduct(t, "async_methods")
}
func TestProduct_StableIdentityEntriesRecordAlias(t *testing.T) {
	t.Parallel()
	stableSplitProduct(t, "entries_record_alias")
}
func TestStableIdentityAsyncMethodsSplit(t *testing.T) {
	t.Parallel()
	stableSplitWitness(t, "async_methods")
}
func TestStableIdentityEntriesRecordAliasSplit(t *testing.T) {
	t.Parallel()
	stableSplitWitness(t, "entries_record_alias")
}
