package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func init() {
	// Invalidated reads are pinned by TestRecordRefusalNamedInvalidated.
	for _, name := range []string{"options", "operations", "union", "intrinsics", "alias_unready"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/records_named_" + name + ".a", true, false})
	}
}

func TestNamedRecordTypeGuardMutant(test *testing.T) {
	test.Parallel()
	if !recordRefusalMutant(test, "named_invalidated") {
		return
	}
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/records_refused/records_named_invalidated.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	want := onJavaScriptBackend(test, program)
	if want.exitCode != 70 {
		test.Fatalf("declared type check did not fire: %+v", want)
	}
	changed := false
	for index, function := range program.Functions {
		if function.Name != "record_declared_member" {
			continue
		}
		body := []ir.Statement{}
		for _, statement := range function.Body {
			if check, ok := statement.(ir.If); ok && len(check.Then) == 1 {
				if _, ok := check.Then[0].(ir.Panic); ok {
					changed = true
					continue
				}
			}
			body = append(body, statement)
		}
		program.Functions[index].Body = body
	}
	if !changed {
		test.Fatal("mutant removed no type check")
	}
	got := onJavaScriptBackend(test, program)
	if got.exitCode != 0 || disagreement(onNode(test, path), got) != "" {
		test.Fatalf("guard-removal mutant did not run on as Node does: %+v", got)
	}
	test.Log("removed member-kind guard caught by exit 70 pin; mutant runs on exactly as Node")
}

func TestNamedRecordAliasReadinessMutant(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/records_named_alias_unready.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	changed := false
	for functionIndex, function := range program.Functions {
		for j, statement := range function.Body {
			declare, ok := statement.(ir.Declare)
			if !ok {
				continue
			}
			read, ok := declare.Value.(ir.Read)
			if !ok || !read.Checked {
				continue
			}
			read.Checked = false
			declare.Value = read
			program.Functions[functionIndex].Body[j] = declare
			changed = true
		}
	}
	if !changed {
		test.Fatal("mutant removed no alias-copy readiness check")
	}
	want := onNode(test, path)
	if want.exitCode != 70 || !strings.Contains(string(want.stderr), "ReferenceError") {
		test.Fatalf("Node did not stop in the dead zone: %+v", want)
	}
	native, _ := nativelyUncached(test, program)
	for name, got := range map[string]run{"native": native, "javascript": onJavaScriptBackend(test, program)} {
		if got.exitCode != 0 || disagreement(want, got) != "exit codes differ" {
			test.Fatalf("%s alias-copy readiness mutant survived: %+v", name, got)
		}
		test.Logf("%s alias-copy readiness mutant caught by main's normalized ReferenceError exit 70", name)
	}
}

func TestNamedRecordJavaScriptAgreesWithNode(test *testing.T) {
	test.Parallel()
	for _, name := range []string{"options", "operations", "union", "intrinsics", "alias_unready"} {
		test.Run(name, func(test *testing.T) {
			path, absolutePathError := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/records_named_"+name+".a"),
			)
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			program, loweredError := lowered(test, path)
			if loweredError != nil {
				test.Fatal(loweredError)
			}
			want, got := onNode(test, path), onJavaScriptBackend(test, program)
			if diff := disagreement(want, got); diff != "" {
				test.Fatalf("%s: Node %+v; backend %+v", diff, want, got)
			}
		})
	}
}

func TestNamedRecordAbsentEntryMutant(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/records_named_union.a"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	changed := false
	for index, statement := range program.Main {
		declaration, ok := statement.(ir.Declare)
		if !ok {
			continue
		}
		literal, ok := declaration.Value.(ir.RecordLiteral)
		if !ok {
			continue
		}
		key := len(program.Strings)
		program.Strings = append(program.Strings, "target")
		literal.Entries = append(
			literal.Entries,
			ir.RecordEntry{Key: ir.StringConstant{Index: key}, Value: ir.Box{Value: ir.Undefined{}}},
		)
		declaration.Value = literal
		program.Main[index] = declaration
		changed = true
	}
	if !changed {
		test.Fatal("mutant changed no dictionary literal")
	}
	want := onNode(test, path)
	native, _ := nativelyUncached(test, program)
	for name, got := range map[string]run{"native": native, "javascript": onJavaScriptBackend(test, program)} {
		if got.exitCode != 0 || disagreement(want, got) != "stdout differs" {
			test.Fatalf("%s absent-as-present mutant survived: Node %+v; mutant %+v", name, want, got)
		}
		test.Logf("%s absent-as-present mutant caught by Node stdout", name)
	}
}
