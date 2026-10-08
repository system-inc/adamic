package oracle

import (
	"context"
	"github.com/system-inc/adamic/internal/flow"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"path/filepath"
	"strings"
	"testing"
)

// Register this unit without editing the shared oracle's compiler-owned fixture list.
func init() {
	for _, name := range []string{"catch_callback", "finally_callback", "rethrow", "finally_completion", "liveness"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{
			"internal/oracle/testdata/step21_" + name + ".a", true, false,
		})
	}
}

// Proposals retain their current barriers; source Node independently establishes their meaning.
func TestStep21ProposalOutcomes(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, diagnostic, stdout string }{
		{"any_value", "refuses throwing a undefined", "undefined\n"},
		{"saved_error", "throwing an Error that isn't made", "saved1\n"},
		{"error_subclass", "base", "true\n"},
		{"unknown_read", "TS18046", "unknown\n"},
		{"library_failure", "a try around repeat", "caught\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "docs/step-21-exceptions/proposals", probe.name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			loaded, err := load.Load([]string{path})
			if err == nil {
				_, err = lower.Lower(context.Background(), loaded)
			}
			if err == nil || !strings.Contains(err.Error(), probe.diagnostic) {
				t.Fatalf("want diagnostic containing %q, got %v", probe.diagnostic, err)
			}
			t.Log(err)
			source := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path)
			if source.exitCode != 0 || string(source.stdout) != probe.stdout || len(source.stderr) != 0 {
				t.Fatalf("source Node: exit %d stdout %q stderr %q", source.exitCode, source.stdout, source.stderr)
			}
		})
	}
}

// A real executed handler/finalizer gains one output line after lowering. The outside
// source oracle must reject it in each backend; warnings and sanitizers cannot kill it.
func TestStep21FixtureMutants(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name    string
		finally bool
	}{
		{"catch_callback", false}, {"finally_callback", true}, {"rethrow", false},
		{"finally_completion", true}, {"liveness", false},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/step21_"+probe.name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			original := onNode(t, path)
			marker := ir.WriteLine{Value: ir.StringConstant{Index: len(program.Strings)}}
			program.Strings = append(program.Strings, "mutated handler")
			changed := step21MutateHandler(program.Main, probe.finally, marker)
			for index := range program.Functions {
				if !changed {
					changed = step21MutateHandler(program.Functions[index].Body, probe.finally, marker)
				}
			}
			if !changed {
				t.Fatal("no handler selected")
			}
			backend := onJavaScriptBackend(t, program)
			compiled, sanitized := natively(t, program)
			release := released(t, program)
			for label, result := range map[string]run{"JavaScript": backend, "sanitized native": compiled, "release native": release} {
				if result.exitCode != 0 || len(result.stderr) != 0 || !strings.Contains(string(result.stdout), "mutated handler\n") {
					t.Fatalf("%s mutant did not execute cleanly: exit %d stdout %q stderr %q", label, result.exitCode, result.stdout, result.stderr)
				}
				if disagreement(original, result) == "" {
					t.Fatalf("%s mutant survived", label)
				}
			}
			if leaked := leaks(t, program, sanitized); leaked != "" {
				t.Fatal(leaked)
			}
			t.Log("mutant caught by stdout in both backends and release native; sanitizer and leak checks clean")
		})
	}
}

func step21MutateHandler(statements []ir.Statement, finally bool, marker ir.Statement) bool {
	for index, statement := range statements {
		switch statement := statement.(type) {
		case ir.Try:
			if finally && statement.HasFinally {
				statement.Finally = append([]ir.Statement{marker}, statement.Finally...)
				statements[index] = statement
				return true
			}
			if !finally && statement.HasCatch {
				statement.Catch = append([]ir.Statement{marker}, statement.Catch...)
				statements[index] = statement
				return true
			}
			if step21MutateHandler(statement.Body, finally, marker) || step21MutateHandler(statement.Catch, finally, marker) || step21MutateHandler(statement.Finally, finally, marker) {
				return true
			}
		case ir.Block:
			if step21MutateHandler(statement.Body, finally, marker) {
				return true
			}
		case ir.If:
			if step21MutateHandler(statement.Then, finally, marker) || step21MutateHandler(statement.Else, finally, marker) {
				return true
			}
		case ir.ForOf:
			if step21MutateHandler(statement.Body, finally, marker) {
				return true
			}
		case ir.Loop:
			if step21MutateHandler(statement.Body, finally, marker) {
				return true
			}
		}
	}
	return false
}

// Source Node reads the original array after the callback assignment throws.
// This pins the old value's lifetime before the earlier consuming grow call.
func TestStep21ThrowPathLiveness(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/step21_liveness.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	for function, body := range program.Functions {
		if body.Name != "read" {
			continue
		}
		text := -1
		for local, value := range program.Locals {
			if value.Function == function && value.Name == "text" {
				text = local
			}
		}
		if text < 0 {
			t.Fatal("missing text local")
		}
		graph := flow.Build(program, function)
		live := flow.LiveOut(graph)
		for _, instruction := range graph.Instructions {
			if _, ok := (*instruction.At).(ir.WriteLine); ok {
				if !live[instruction.Id][flow.DeclarationId(text+1)] {
					t.Fatal("old text is dead before a callback that throws into a catch reading it")
				}
				return
			}
		}
		t.Fatal("missing consuming statement")
	}
	t.Fatal("missing read function")
}
