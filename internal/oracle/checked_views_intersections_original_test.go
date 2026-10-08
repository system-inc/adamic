package oracle

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

type intersectionOriginalManifest struct {
	Commit       string              `json:"upstream_commit"`
	Declarations map[string]string   `json:"declarations"`
	Fields       map[string][]string `json:"fields"`
	Pairs        []struct {
		ID            int               `json:"type_id"`
		Declared      string            `json:"declared_type"`
		Type          string            `json:"type"`
		Field         string            `json:"field"`
		Reads         int               `json:"read_count"`
		Sites         []json.RawMessage `json:"sites"`
		PresentFields []string          `json:"present_fields"`
	} `json:"pairs"`
}

func intersectionOriginalInputs(t *testing.T) (string, intersectionOriginalManifest) {
	t.Helper()
	declarations := os.Getenv("ADAMIC_INTERSECTION_ORIGINAL_DECLS")
	if declarations == "" {
		t.Skip("set ADAMIC_INTERSECTION_ORIGINAL_DECLS to pinned declaration output")
	}
	data, err := os.ReadFile(filepath.Join(declarations, "intersection-manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest intersectionOriginalManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	reads := map[int]int{10236: 1, 7612: 4, 9476: 9, 9474: 11, 9485: 9, 9475: 4, 9454: 1}
	if manifest.Commit != "050880ce59e30b356b686bd3144efe24f875ebc8" || len(manifest.Declarations) != 78 || len(manifest.Pairs) != len(reads) || manifest.Pairs[0].ID != 10236 {
		t.Fatal("original provenance changed")
	}
	for _, pair := range manifest.Pairs {
		if reads[pair.ID] != pair.Reads || len(pair.Sites) != pair.Reads {
			t.Fatalf("original read sites changed for %d.%s", pair.ID, pair.Field)
		}
	}
	for name, digest := range manifest.Declarations {
		data, err := os.ReadFile(filepath.Join(declarations, name))
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprintf("%x", sha256.Sum256(data)) != digest {
			t.Fatal("declaration drift: " + name)
		}
	}
	return declarations, manifest
}

func intersectionOriginalProgram(t *testing.T, declarations, name, source string) (*ir.Program, string) {
	t.Helper()
	input, err := os.ReadFile("../../stage3/interface-downcasts/lane7/original/" + name + ".a")
	if err != nil {
		t.Fatal(err)
	}
	bound := strings.Replace(string(input), "'original-tsc-types'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(declarations, "compiler/types.d.ts"))), 1)
	file := filepath.Join(t.TempDir(), name+".a")
	if err := os.WriteFile(file, []byte(bound), 0600); err != nil {
		t.Fatal(err)
	}
	if difference := disagreement(run{stdout: []byte(source)}, onNode(t, file)); difference != "" {
		t.Fatal("Node: " + difference)
	}
	program, err := lowered(t, file)
	if err != nil {
		t.Fatal(err)
	}
	return program, file
}

func requireIntersectionOriginalFields(t *testing.T, program *ir.Program, manifest intersectionOriginalManifest) {
	t.Helper()
	for _, name := range []string{"SymbolTracker", "ModuleSpecifierResolutionHost"} {
		complete := false
		for _, c := range program.ViewContracts {
			if c.Name != name {
				continue
			}
			fields := []string{}
			for _, f := range c.Fields {
				fields = append(fields, f.Name)
			}
			slices.Sort(fields)
			complete = complete || len(fields) > 0 && slices.Equal(fields, manifest.Fields[name])
		}
		if !complete {
			t.Fatal("original field set was reduced: " + name)
		}
	}
	complete := false
	for _, c := range program.ViewContracts {
		if !c.Intersection || c.Unsupported != "" {
			continue
		}
		fields := []string{}
		for _, f := range c.Fields {
			fields = append(fields, f.Name)
		}
		slices.Sort(fields)
		complete = complete || slices.Equal(fields, manifest.Pairs[0].PresentFields)
	}
	if !complete {
		t.Fatal("original intersection obligations missing")
	}
}

func TestCheckedViewIntersectionOriginalPairs(t *testing.T) {
	declarations, manifest := intersectionOriginalInputs(t)
	for _, test := range []struct{ name, source, diagnostic string }{
		{"tracker-good", "true\n", ""}, {"tracker-helpers-good", "true\n", ""},
		{"tracker-absent", "false\n", ""}, {"tracker-undefined", "false\n", ""},
		{"tracker-wrong", "true\n", "field read failed: tracker?.moduleResolverHost.useCaseSensitiveFileNames is not a () => boolean; expected () => boolean, found string"},
		{"tracker-helpers-wrong", "true\n", "field read failed: tracker?.moduleResolverHost.useCaseSensitiveFileNames is not a () => boolean; expected () => boolean, found string"},
		{"tracker-missing", "true\n", "field read failed: tracker?.moduleResolverHost.useCaseSensitiveFileNames is not initialized; expected () => boolean, found missing"},
		{"tracker-root-wrong", "true\n", "field read failed: tracker?.moduleResolverHost matches no member of (ModuleSpecifierResolutionHost & { getCommonSourceDirectory(): string; }) | undefined; expected (ModuleSpecifierResolutionHost & { getCommonSourceDirectory(): string; }) | undefined, found string"},
	} {
		t.Run(test.name, func(t *testing.T) {
			program, _ := intersectionOriginalProgram(t, declarations, test.name, test.source)
			requireIntersectionOriginalFields(t, program, manifest)
			if kind := os.Getenv("ADAMIC_INTERSECTION_ORIGINAL_MUTANT"); kind != "" {
				intersectionOriginalMutant(t, program, kind)
			}
			want := run{stdout: []byte(test.source)}
			if test.diagnostic != "" {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: " + test.diagnostic + "\n")}
			}
			actual, binary := nativelyUncached(t, program)
			if want.exitCode == 0 {
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Errorf("%s; got %#v", difference, got)
				}
			}
		})
	}
}

func intersectionOriginalMutant(t *testing.T, program *ir.Program, kind string) {
	t.Helper()
	changed := 0
	for i := range program.ViewContracts {
		c := &program.ViewContracts[i]
		if kind == "skip" && c.Intersection {
			c.Intersection = false
			changed++
		}
		for j, f := range c.Fields {
			if f.Name != "useCaseSensitiveFileNames" {
				continue
			}
			switch kind {
			case "nested":
				c.Fields = append(c.Fields[:j:j], c.Fields[j+1:]...)
				changed++
			case "shape":
				child := program.ViewContracts[f.Contract-1]
				child.Kind = ir.ViewScalar
				child.Of = ir.String
				child.Name = "string"
				c.Fields[j].Contract = ir.ViewContractID(len(program.ViewContracts) + 1)
				program.ViewContracts = append(program.ViewContracts, child)
				changed++
			case "presence":
				c.Fields[j].Optional = true
				changed++
			}
			break
		}
	}
	if kind == "outer" || kind == "absence" {
		changed = changeObjectPrimitiveRead(program, func(p ir.Property) bool { return p.Name == "moduleResolverHost" }, func(p ir.Property) ir.Property {
			if kind == "outer" {
				p.View = ""
			} else {
				p.Absent = false
			}
			return p
		})
	}
	if changed == 0 {
		t.Fatal("original mutant found no obligation")
	}
}

// Complete original node declarations require a Node parent and a Symbol on
// every declaration. Each read checks every descriptor once per path, selecting
// tagged arms; a descriptor already entered keeps its fields for its own reads.
func requireIntersectionOriginalBindable(t *testing.T, program *ir.Program, manifest intersectionOriginalManifest, ids ...int) {
	t.Helper()
	sorted := func(c ir.ViewContract) []string {
		fields := []string{}
		for _, f := range c.Fields {
			fields = append(fields, f.Name)
		}
		slices.Sort(fields)
		return fields
	}
	for _, name := range []string{"Identifier", "Node", "Symbol", "PropertyAccessEntityNameExpression", "ElementAccessExpression"} {
		complete := false
		for _, c := range program.ViewContracts {
			complete = complete || c.Name == name && len(c.Fields) > 0 && slices.Equal(sorted(c), manifest.Fields[name])
		}
		if !complete {
			t.Fatal("original field set was reduced: " + name)
		}
	}
	for _, pair := range manifest.Pairs {
		if !slices.Contains(ids, pair.ID) {
			continue
		}
		bounded := false
		for _, c := range program.ViewContracts {
			bounded = bounded || c.Name == pair.Declared && c.IntersectionBounded && c.Unsupported == "" && slices.Equal(sorted(c), pair.PresentFields)
		}
		if !bounded {
			t.Fatalf("original %d.%s read is not held by its complete bounded contract", pair.ID, pair.Field)
		}
	}
}

func TestCheckedViewIntersectionOriginalBindable(t *testing.T) {
	declarations, manifest := intersectionOriginalInputs(t)
	static, access := []int{9476, 9474}, []int{9485, 9475}
	for _, test := range []struct {
		name, source, diagnostic string
		pairs                    []int
	}{
		{"bindable-static-good", "true\n", "", static},
		{"bindable-static-chain-good", "true\n", "", static},
		{"bindable-static-helpers-good", "true\n", "", static[:1]},
		{"bindable-element-good", "true\n", "", append(slices.Clone(static), 9454)},
		{"bindable-static-left-wrong", "true\n", "field read failed: node.left.name.symbol is not initialized; expected Symbol, found missing", static[:1]},
		{"bindable-static-helpers-wrong", "true\n", "field read failed: node.left.name.symbol is not initialized; expected Symbol, found missing", static[:1]},
		{"bindable-static-left-kind", "true\n", "field read failed: node.left.kind expected SyntaxKind.PropertyAccessExpression | SyntaxKind.ElementAccessExpression, found number 999", static[:1]},
		{"bindable-static-expression-wrong", "true\n", "field read failed: node.left.expression.expression.symbol is not initialized; expected Symbol, found missing", static},
		{"bindable-static-expression-this", "true\n", "field read failed: node.left.expression.expression.kind expected SyntaxKind.Identifier | SyntaxKind.PropertyAccessExpression, found number 110", static},
		{"bindable-element-argument-wrong", "true\n", "field read failed: node.left.argumentExpression.kind expected SyntaxKind.NumericLiteral | SyntaxKind.StringLiteral | SyntaxKind.NoSubstitutionTemplateLiteral, found number 80", static},
		{"bindable-access-good", "true\n", "", access},
		{"bindable-access-chain-good", "true\n", "", access},
		{"bindable-access-left-wrong", "true\n", "field read failed: node.left.expression.kind expected SyntaxKind.Identifier | SyntaxKind.PropertyAccessExpression | SyntaxKind.ElementAccessExpression, found number 110", access[:1]},
		{"bindable-access-left-symbol", "true\n", "field read failed: node.left.expression.symbol is not initialized; expected Symbol, found missing", access[:1]},
		{"bindable-access-left-kind", "true\n", "field read failed: node.left.kind expected SyntaxKind.PropertyAccessExpression | SyntaxKind.ElementAccessExpression, found number 999", access[:1]},
		{"bindable-access-expression-wrong", "true\n", "field read failed: node.left.expression.expression.symbol is not initialized; expected Symbol, found missing", access},
		{"bindable-access-expression-this", "true\n", "field read failed: node.left.expression.expression.kind expected SyntaxKind.Identifier | SyntaxKind.PropertyAccessExpression, found number 110", access},
	} {
		t.Run(test.name, func(t *testing.T) {
			program, _ := intersectionOriginalProgram(t, declarations, test.name, test.source)
			requireIntersectionOriginalBindable(t, program, manifest, test.pairs...)
			if kind := os.Getenv("ADAMIC_INTERSECTION_ORIGINAL_MUTANT"); kind != "" {
				intersectionOriginalBindableMutant(t, program, kind)
			}
			want := run{stdout: []byte(test.source)}
			if test.diagnostic != "" {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: " + test.diagnostic + "\n")}
			}
			actual, binary := nativelyUncached(t, program)
			if want.exitCode == 0 {
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Errorf("%s; got %#v", difference, got)
				}
			}
		})
	}
}

// Each mutant removes one obligation from one pair's read: skip falls back to
// the plain union's discriminant, shape accepts a wrong tag, nested forgets an
// Identifier's Symbol.
func intersectionOriginalBindableMutant(t *testing.T, program *ir.Program, kind string) {
	t.Helper()
	roots := map[string]string{"left": "BindableStaticAccessExpression", "expression": "EntityNameExpression | (LeftHandSideExpression & BindableStaticNameExpression)", "access": "BindableAccessExpression"}
	tags := map[string][2]any{"left": {"SyntaxKind.PropertyAccessExpression", 999.0}, "expression": {"SyntaxKind.Identifier", 110.0}, "access": {"SyntaxKind.ElementAccessExpression", 999.0}}
	pair, change, _ := strings.Cut(kind, "-")
	changed := 0
	for i := range program.ViewContracts {
		c := &program.ViewContracts[i]
		switch change {
		case "skip":
			if c.Name == roots[pair] && c.IntersectionBounded {
				c.IntersectionBounded, c.IntersectionTag = false, ""
				changed++
			}
		case "shape":
			if c.Name == tags[pair][0] && c.Kind == ir.ViewScalar {
				c.Allowed = append(slices.Clone(c.Allowed), ir.ViewLiteral{Of: ir.Number, Number: tags[pair][1].(float64)})
				changed++
			}
		case "nested":
			if c.Name == "Identifier" || c.Name == "LeftHandSideExpression & Identifier" {
				c.Fields = slices.DeleteFunc(slices.Clone(c.Fields), func(f ir.ViewFieldContract) bool { return f.Name == "symbol" })
				changed++
			}
		}
	}
	if changed == 0 {
		t.Fatal("original bindable mutant found no obligation: " + kind)
	}
}
