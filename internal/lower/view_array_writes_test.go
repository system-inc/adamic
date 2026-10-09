package lower

import (
	"context"
	"errors"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"os"
	"path/filepath"
	"testing"
)

func TestViewArrayWritesNeedSourceCertificate(t *testing.T) {
	for _, write := range []string{"items(raw).values.push(true);", "items(raw).values[0] = true;"} {
		t.Run(write, func(t *testing.T) {
			source := "interface Base { readonly kind: 'items' | 'other'; }\ninterface Items extends Base { readonly kind: 'items'; readonly values: boolean[]; }\nfunction items(node: Base): Items { return node as Items; }\nconst raw = {kind: 'items' as const, values: [1]};\n" + write
			path := filepath.Join(t.TempDir(), "write.a")
			if err := os.WriteFile(path, []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
			program, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			_, err = Lower(context.Background(), program)
			var refusal *NotYet
			if !errors.As(err, &refusal) || refusal.What != "writing through an array cast without its source-slot type certificate" {
				t.Fatalf("wanted source-slot refusal, got %v", err)
			}
		})
	}
}

func TestViewArraySourceCertificateOmission(t *testing.T) {
	path := filepath.Join(t.TempDir(), "certificate.a")
	if err := os.WriteFile(path, []byte("console.log('control');"), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	p := &ir.Program{Source: path, ArrayViewEnabled: true, ArrayViewNeedsSourceCertificate: true, Main: []ir.Statement{ir.SetIndex{Array: ir.ArrayLiteral{Element: ir.Number}, Index: ir.NumberConstant{Value: 0}, Value: ir.BooleanConstant{Value: true}, Element: ir.Boolean, Site: 1}}}
	l := &lowering{program: program, result: p}
	control := l.activateViewArrayReads(nil, nil, false)
	if control == nil {
		t.Fatal("source write certificate guard missing")
	}
	p.ArrayViewNeedsSourceCertificate = false
	mutant := l.activateViewArrayReads(nil, nil, false)
	if mutant != nil {
		t.Fatalf("omission did not remove the refusal: %v", mutant)
	}
	t.Log("source-slot certificate guard omitted; the pinned refusal rejects the mutant")
}

// Array readers certify selected slots, not an untagged union's membership.
// V2 removes this boundary when its array matcher is installed.
func TestArrayMembershipAwaitingUnionMatcher(t *testing.T) {
	p := &ir.Program{ViewContracts: []ir.ViewContract{
		{Kind: ir.ViewScalar, Of: ir.Number},
		{Kind: ir.ViewArray, Of: ir.Array, Element: 1},
		{Kind: ir.ViewObject, Of: ir.Object, Fields: []ir.ViewFieldContract{{Name: "values", Contract: 2}}},
	}}
	root := ir.ViewContract{Kind: ir.ViewUnion, Of: ir.Object, Members: []ir.ViewContractID{3}}
	l := &lowering{result: p}
	if l.supportsUntaggedRead(root) {
		t.Fatal("awaits views-v2: array union membership dispatch")
	}
}
