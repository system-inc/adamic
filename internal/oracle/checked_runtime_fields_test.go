package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

func checkedRuntimeFieldProgram(t *testing.T, name string) (*ir.Program, string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/checked-runtime-fields", name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	_, refusal := lowered(t, path)
	if refusal == nil || !strings.Contains(refusal.Error(), "Adamic 0.1 refuses") || !strings.Contains(refusal.Error(), "readonly") {
		t.Fatalf("want unchanged wider writable view refusal and fix, got %v", refusal)
	}
	t.Logf(".a refusal: %v", refusal)
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	adapted := filepath.Join(t.TempDir(), name+".ts")
	if err := os.WriteFile(adapted, source, 0600); err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, adapted)
	if err != nil {
		t.Fatal(err)
	}
	if !program.CheckedElements {
		t.Fatal("TypeScript adapter did not enable structural element checks")
	}
	return program, path
}

func checkedRuntimeFields(t *testing.T, name string) {
	t.Helper()
	program, path := checkedRuntimeFieldProgram(t, name)
	truth := onNode(t, path)
	want := map[string]string{"regex": "a\na\ndone\n", "map": "alpha\ndone\n", "set": "alpha\ndone\n", "fs": "alpha\n", "directory": "alpha.txt\n", "status": "file 6 false\n"}[name]
	if truth.exitCode != 0 || string(truth.stdout) != want {
		t.Fatalf("Node: %#v", truth)
	}
	sanitized, binary := nativelyUncached(t, program)
	for backend, got := range map[string]run{"native sanitized": sanitized, "native release": releasedUncached(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(truth, got); difference != "" {
			t.Fatalf("%s: %s; got %#v", backend, difference, got)
		}
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
	t.Logf("%s: Node exit 0 stdout %q; TypeScript-mode structural contracts, native sanitized/release and JavaScript agree", path, truth.stdout)
}

func TestCheckedRuntimeRegexFields(t *testing.T) { t.Parallel(); checkedRuntimeFields(t, "regex") }
func TestCheckedRuntimeMapFields(t *testing.T)   { t.Parallel(); checkedRuntimeFields(t, "map") }
func TestCheckedRuntimeSetFields(t *testing.T)   { t.Parallel(); checkedRuntimeFields(t, "set") }
func TestCheckedRuntimeFSFields(t *testing.T)    { t.Parallel(); checkedRuntimeFields(t, "fs") }
func TestCheckedRuntimeDirectoryFields(t *testing.T) {
	t.Parallel()
	checkedRuntimeFields(t, "directory")
}
func TestCheckedRuntimeStatusFields(t *testing.T) { t.Parallel(); checkedRuntimeFields(t, "status") }

func init() {
	additionalFixtureCounts = append(additionalFixtureCounts, func(t *testing.T) []string {
		rows := []string{}
		for _, name := range []string{"regex", "map", "set", "fs", "directory", "status"} {
			program, _ := checkedRuntimeFieldProgram(t, name)
			binary := filepath.Join(t.TempDir(), "counted")
			if err := native.Build(native.C(program), binary, native.Options{Count: true}); err != nil {
				t.Fatal(err)
			}
			nameCommand, arguments := pinnedStack(binary)
			result := execute(t, nameCommand, arguments...)
			match := countsLine.FindSubmatch(result.stderr)
			if result.exitCode != 0 || match == nil {
				t.Fatalf("counted %s: %#v", name, result)
			}
			rows = append(rows, fmt.Sprintf("| internal/oracle/testdata/checked-runtime-fields/%s.a (TypeScript adapter) | %s | %s | %s | %s | %s | %s |", name, match[1], match[2], match[3], match[4], match[5], match[6]))
		}
		return rows
	})
}

// An internal writer has no checker source type. After a runtime reference array
// is admitted, its installed element contract must still catch a later bad push.
func TestCheckedRuntimeReferenceArrayLaterWriter(t *testing.T) {
	t.Parallel()
	program, path := checkedRuntimeFieldProgram(t, "regex")
	truth := onNode(t, path)
	local := -1
	for index, variable := range program.Locals {
		if variable.Name == "results" && variable.Global {
			local = index
			break
		}
	}
	if local < 0 {
		t.Fatal("missing result array")
	}
	receiver := ir.Property{Object: ir.ArrayIndex{Array: ir.Read{Local: local, Of: ir.Array}, Index: ir.NumberConstant{Value: 0}, Element: ir.Object}, Name: "value", Of: ir.Array}
	program.Main = append(program.Main, ir.Evaluate{Value: ir.ArrayPush{Array: receiver, Value: ir.ObjectLiteral{}, Element: ir.String, WriteOrigin: ir.WriteCheck{Expression: "later[]"}}})
	want := run{exitCode: 70, stdout: truth.stdout, stderr: []byte("adamic: panic: write failed: later[] expects string | undefined, got object\n")}
	sanitized, _ := nativelyUncached(t, program)
	for backend, got := range map[string]run{"native sanitized": sanitized, "native release": releasedUncached(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Fatalf("%s: %s; got %#v", backend, difference, got)
		}
	}
}
