package oracle

import (
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/records_named_invalidated.a", true, true})
	for _, name := range []string{"options", "operations", "union", "intrinsics", "alias_unready"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/records_named_" + name + ".a", true, false})
	}
}

func TestNamedRecordTypeGuardMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/records_named_invalidated.a"))
	if err != nil {
		t.Fatal(err)
	}
	p, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := onJavaScriptBackend(t, p)
	if want.exitCode != 70 {
		t.Fatalf("declared type check did not fire: %+v", want)
	}
	changed := false
	for i, f := range p.Functions {
		if f.Name != "record_declared_member" {
			continue
		}
		body := []ir.Statement{}
		for _, s := range f.Body {
			if check, ok := s.(ir.If); ok && len(check.Then) == 1 {
				if _, ok := check.Then[0].(ir.Panic); ok {
					changed = true
					continue
				}
			}
			body = append(body, s)
		}
		p.Functions[i].Body = body
	}
	if !changed {
		t.Fatal("mutant removed no type check")
	}
	got := onJavaScriptBackend(t, p)
	if got.exitCode != 0 || disagreement(onNode(t, path), got) != "" {
		t.Fatalf("guard-removal mutant did not run on as Node does: %+v", got)
	}
	t.Log("removed member-kind guard caught by exit 70 pin; mutant runs on exactly as Node")
}

func TestNamedRecordAliasReadinessMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/records_named_alias_unready.a"))
	if err != nil {
		t.Fatal(err)
	}
	p, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := false
	for i, f := range p.Functions {
		for j, s := range f.Body {
			d, ok := s.(ir.Declare)
			if !ok {
				continue
			}
			r, ok := d.Value.(ir.Read)
			if !ok || !r.Checked {
				continue
			}
			r.Checked = false
			d.Value = r
			p.Functions[i].Body[j] = d
			changed = true
		}
	}
	if !changed {
		t.Fatal("mutant removed no alias-copy readiness check")
	}
	want := onNode(t, path)
	if want.exitCode != 70 {
		t.Fatalf("Node did not stop in the dead zone: %+v", want)
	}
	native, _ := nativelyUncached(t, p)
	for name, got := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, p)} {
		if got.exitCode != 0 || disagreement(want, got) != "exit codes differ" {
			t.Fatalf("%s alias-copy readiness mutant survived: %+v", name, got)
		}
		t.Logf("%s alias-copy readiness mutant caught by Node's exit 70", name)
	}
}

func TestNamedRecordJavaScriptAgreesWithNode(t *testing.T) {
	for _, name := range []string{"options", "operations", "union", "intrinsics", "alias_unready"} {
		t.Run(name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/records_named_"+name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			p, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			want, got := onNode(t, path), onJavaScriptBackend(t, p)
			if diff := disagreement(want, got); diff != "" {
				t.Fatalf("%s: Node %+v; backend %+v", diff, want, got)
			}
		})
	}
}

func TestNamedRecordAbsentEntryMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/records_named_union.a"))
	if err != nil {
		t.Fatal(err)
	}
	p, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := false
	for i, statement := range p.Main {
		declaration, ok := statement.(ir.Declare)
		if !ok {
			continue
		}
		literal, ok := declaration.Value.(ir.RecordLiteral)
		if !ok {
			continue
		}
		key := len(p.Strings)
		p.Strings = append(p.Strings, "target")
		literal.Entries = append(literal.Entries, ir.RecordEntry{Key: ir.StringConstant{Index: key}, Value: ir.Box{Value: ir.Undefined{}}})
		declaration.Value = literal
		p.Main[i] = declaration
		changed = true
	}
	if !changed {
		t.Fatal("mutant changed no dictionary literal")
	}
	want := onNode(t, path)
	native, _ := nativelyUncached(t, p)
	for name, got := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, p)} {
		if got.exitCode != 0 || disagreement(want, got) != "stdout differs" {
			t.Fatalf("%s absent-as-present mutant survived: Node %+v; mutant %+v", name, want, got)
		}
		t.Logf("%s absent-as-present mutant caught by Node stdout", name)
	}
}
