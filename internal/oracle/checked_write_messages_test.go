package oracle

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func checkedWriteMessageProgram(t *testing.T, name string) (*ir.Program, string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/checked-write-messages", name+"_out.a"))
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	// The source's .a refusal remains intact. This temporary frontend adapter is
	// the same TypeScript-mode boundary exercised by the upstream fixture driver.
	_, refusal := lowered(t, path)
	if refusal == nil || !strings.Contains(refusal.Error(), "Adamic 0.1 refuses") || !strings.Contains(refusal.Error(), "readonly") {
		t.Fatalf("want the unchanged .a refusal and fix, got %v", refusal)
	}
	t.Logf(".a refusal: %v", refusal)
	adapted := filepath.Join(t.TempDir(), name+"_out.ts")
	if err := os.WriteFile(adapted, source, 0600); err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, adapted)
	if err != nil {
		t.Fatal(err)
	}
	return program, path
}

func checkedWriteMessage(t *testing.T, name string, mutant bool) {
	t.Helper()
	program, path := checkedWriteMessageProgram(t, name)

	truth := onNode(t, path)
	if truth.exitCode != 0 || !strings.HasPrefix(string(truth.stdout), "before write\n") {
		t.Fatalf("Node: %#v", truth)
	}
	stderr, err := os.ReadFile(path + ".stderr")
	if err != nil {
		t.Fatal(err)
	}
	want := run{exitCode: 70, stdout: []byte("before write\n"), stderr: stderr}
	if mutant {
		// Change only diagnostic declarations, leaving kinds, allowed values,
		// allocation IDs and directional proof tables intact.
		types := map[string][2]string{
			"02_shared-empty":          {"never", "number"},
			"03_flow-node":             {"BinaryExpression", "BinaryExpression | BindingElement"},
			"13_declaration-array":     {"Declaration", "Node"},
			"15_detached-diagnostic":   {"undefined", "SourceFile | undefined"},
			"16_flow-assignment-union": {"BinaryExpression", "BinaryExpression | BindingElement"},
		}[name]
		from, to := types[0], types[1]
		changed := 0
		for _, contract := range program.WriteContracts {
			if contract.Declared == from {
				contract.Declared = to
				changed++
			}
		}
		if changed == 0 {
			t.Fatal("message mutant changed no declaration")
		}
	}
	sanitized, _ := nativelyUncached(t, program)
	for backend, got := range map[string]run{"native sanitized": sanitized, "native release": releasedUncached(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
		difference := disagreement(want, got)
		if mutant {
			if got.exitCode != 70 || string(got.stdout) != "before write\n" || difference == "" {
				t.Fatalf("%s: diagnostic mutant was not caught: %#v", backend, got)
			}
			t.Logf("%s message mutant caught by exact stderr: %q", backend, got.stderr)
		} else if difference != "" {
			t.Fatalf("%s: %s; got %#v", backend, difference, got)
		}
	}
}

func TestCheckedWriteMessage02(t *testing.T) {
	t.Parallel()
	checkedWriteMessage(t, "02_shared-empty", false)
}
func TestCheckedWriteMessage03(t *testing.T) {
	t.Parallel()
	checkedWriteMessage(t, "03_flow-node", false)
}
func TestCheckedWriteMessage13(t *testing.T) {
	t.Parallel()
	checkedWriteMessage(t, "13_declaration-array", false)
}
func TestCheckedWriteMessage15(t *testing.T) {
	t.Parallel()
	checkedWriteMessage(t, "15_detached-diagnostic", false)
}
func TestCheckedWriteMessage16(t *testing.T) {
	t.Parallel()
	checkedWriteMessage(t, "16_flow-assignment-union", false)
}
func TestCheckedWriteMessageMutant03(t *testing.T) {
	t.Parallel()
	checkedWriteMessage(t, "03_flow-node", true)
}
func TestCheckedWriteMessageMutant13(t *testing.T) {
	t.Parallel()
	checkedWriteMessage(t, "13_declaration-array", true)
}
func TestCheckedWriteMessageMutant15(t *testing.T) {
	t.Parallel()
	checkedWriteMessage(t, "15_detached-diagnostic", true)
}
func TestCheckedWriteMessageMutant16(t *testing.T) {
	t.Parallel()
	checkedWriteMessage(t, "16_flow-assignment-union", true)
}

func init() {
	additionalFixtureCounts = append(additionalFixtureCounts, func(t *testing.T) []string {
		rows := []string{}
		for _, name := range []string{"02_shared-empty", "03_flow-node", "13_declaration-array", "15_detached-diagnostic", "16_flow-assignment-union"} {
			program, path := checkedWriteMessageProgram(t, name)
			binary := filepath.Join(t.TempDir(), "counted-message")
			if err := native.Build(native.C(program), binary, native.Options{Count: true}); err != nil {
				t.Fatal(err)
			}
			command, arguments := pinnedStack(binary)
			result := execute(t, command, arguments...)
			expected, err := os.ReadFile(path + ".stderr")
			if err != nil {
				t.Fatal(err)
			}
			match := countsLine.FindSubmatch(result.stderr)
			if result.exitCode != 70 || match == nil || !strings.HasPrefix(string(result.stderr), string(expected)) {
				t.Fatalf("counted %s: %#v", name, result)
			}
			rows = append(rows, fmt.Sprintf("| internal/oracle/testdata/checked-write-messages/%s_out.a (TypeScript adapter) | %s | %s | %s | %s | %s | %s |", name, match[1], match[2], match[3], match[4], match[5], match[6]))
		}
		return rows
	})
}
