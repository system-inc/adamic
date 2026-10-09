package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func TestViewIntersectionContracts(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name, source string
		members      int
		refused      bool
	}{
		{"interfaces", "interface A { readonly text: string } interface B { readonly count: number } type Target = A & B;", 2, false},
		{"nested", "interface A { readonly child: { readonly text: string } } type Target = A & { readonly child: { readonly count: number } };", 2, false},
		{"brand", "interface A { readonly text: string } type Target = A & { readonly __brand: void };", 1, false},
		{"runtime member", "interface A { readonly text: string } type Target = A & { readonly __brand: number };", 2, false},
		{"optional brand", "interface A { readonly text: string } type Target = A & { readonly __brand?: undefined };", 1, false},
		{"primitive brand lane", "type Target = string & { readonly __brand: void };", 0, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "contract.a")
			if err := os.WriteFile(path, []byte(test.source), 0600); err != nil {
				t.Fatal(err)
			}
			program, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			file := program.Files()[0]
			checked, release := program.Checker(context.Background(), file)
			defer release()
			l := &lowering{program: program, checker: checked, result: &ir.Program{}}
			node := file.Statements.Nodes[len(file.Statements.Nodes)-1]
			target := checked.GetTypeAtLocation(node.Name())
			calls := 0
			build := func(part *checker.Type) (ir.ViewContractID, error) {
				calls++
				l.result.ViewContracts = append(l.result.ViewContracts, ir.ViewContract{Kind: ir.ViewObject, Name: checked.TypeToString(part)})
				return ir.ViewContractID(len(l.result.ViewContracts)), nil
			}
			members, err := l.viewIntersectionContracts(node, target, build)
			if test.refused {
				if err == nil || calls != 0 {
					t.Fatalf("primitive family must remain separate: %v, calls %d", err, calls)
				}
				return
			}
			if err != nil || len(members) != test.members || calls != test.members {
				t.Fatalf("members %v, calls %d, error %v; want %d", members, calls, err, test.members)
			}
			if !l.structuralViewIntersection(target) {
				t.Fatal("structural intersection not recognized")
			}
			l.result = &ir.Program{}
			id, err := l.internStructuralViewIntersection(node, target)
			if err != nil {
				t.Fatal(err)
			}
			contract := l.result.ViewContracts[id-1]
			if !contract.Intersection || contract.Kind != ir.ViewObject || len(contract.Members) != test.members {
				t.Fatalf("lost conjunction: %#v", contract)
			}
			if test.name == "brand" || test.name == "optional brand" {
				for _, field := range contract.Fields {
					if field.Name == "__brand" {
						t.Fatal("phantom introduced runtime field")
					}
				}
			}
			sentinel := errors.New("member unavailable")
			members, err = l.viewIntersectionContracts(node, target, func(*checker.Type) (ir.ViewContractID, error) { return 0, sentinel })
			if !errors.Is(err, sentinel) || members != nil {
				t.Fatalf("lost constituent refusal: %v, %v", members, err)
			}
			members, err = l.viewIntersectionContracts(node, target, func(*checker.Type) (ir.ViewContractID, error) { return 0, nil })
			if err == nil || members != nil {
				t.Fatal("unknown member certified intersection")
			}
		})
	}
}

func TestViewIntersectionAncestorProof(t *testing.T) {
	t.Parallel()
	for _, sample := range []struct {
		name, source string
		inherits     bool
	}{
		{"declared", "interface A { readonly value: number } interface B extends A { readonly other: string } type Target = B;", true},
		{"structural only", "interface A { readonly value: number } interface B { readonly value: number; readonly other: string } type Target = B;", false},
		{"intersection", "interface A { readonly value: number } interface B extends A { readonly other: string } type Target = B & { readonly more: number };", true},
		{"generic ancestry", "interface A { readonly value: number } interface B<T> extends A { readonly other: T } type Target = B<string>;", true},
	} {
		t.Run(sample.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "ancestor.a")
			if err := os.WriteFile(path, []byte(sample.source), 0600); err != nil {
				t.Fatal(err)
			}
			program, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			file := program.Files()[0]
			checked, release := program.Checker(context.Background(), file)
			defer release()
			l := &lowering{program: program, checker: checked, result: &ir.Program{}}
			ancestor := checked.GetTypeAtLocation(file.Statements.Nodes[0].Name())
			source := checked.GetTypeAtLocation(file.Statements.Nodes[len(file.Statements.Nodes)-1].Name())
			if got := l.viewIntersectionInherits(source, ancestor, map[*checker.Type]bool{}); got != sample.inherits {
				t.Fatalf("declared ancestry %v, want %v", got, sample.inherits)
			}
		})
	}
}

func TestViewIntersectionUnsupportedAtCreation(t *testing.T) {
	t.Parallel()
	source := `interface A { readonly count: number } interface B { readonly hidden: unknown } interface Box { readonly child: object } interface Target extends Box { readonly child: A & B } const raw: Box = {child: {count: 1}}; const view = raw as Target; console.log("unread");`
	_, err := lowerSource(t, source)
	if err == nil || !strings.Contains(err.Error(), "unsupported member") || !strings.Contains(err.Error(), ".hidden") {
		t.Fatalf("unread unsupported member must refuse at creation with its path: %v", err)
	}
	var refused *Refused
	if !errors.As(err, &refused) || refused.Fix == "" {
		t.Fatalf("creation refusal needs a fix: %v", err)
	}
}

func TestViewIntersectionOptionalDescendantAtCreation(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `interface A { readonly text: string } interface B { readonly next?: B; readonly code: 1 } interface Box { readonly child: object } interface Target extends Box { readonly child: A & B } const raw: Box = {child: {text: "x", code: 1}}; const view = raw as Target; console.log("unread");`)
	if err == nil || !strings.Contains(err.Error(), "unsupported member") || !strings.Contains(err.Error(), ".next") {
		t.Fatalf("optional recursive descendant needs a complete handoff at creation: %v", err)
	}
}
func TestViewIntersectionDistributedUnionAtCreation(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `interface A { readonly code: 1 } interface B { readonly code: 2 } interface Extra { readonly text: string } interface Box { readonly child: object } interface Target extends Box { readonly child: A | (B & Extra) } const raw: Box = {child: {code: 2}}; const view = raw as Target; console.log("unread");`)
	if err == nil || !strings.Contains(err.Error(), "unsupported member") || !strings.Contains(err.Error(), ".child") {
		t.Fatalf("distributed intersection cannot lose its obligation at creation: %v", err)
	}
}
