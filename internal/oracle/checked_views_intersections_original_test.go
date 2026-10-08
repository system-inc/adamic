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
	Commit             string              `json:"upstream_commit"`
	Declarations       map[string]string   `json:"declarations"`
	Fields             map[string][]string `json:"fields"`
	AssignedIdentifier struct {
		ID        int    `json:"receiver_type_id"`
		Receiver  string `json:"type"`
		Field     string `json:"field"`
		Reads     int    `json:"reads"`
		SourceSHA string `json:"source_sha256"`
	} `json:"assigned_identifier"`
	Pairs []struct {
		ID             int               `json:"type_id"`
		Declared       string            `json:"declared_type"`
		Type           string            `json:"type"`
		Field          string            `json:"field"`
		Reads          int               `json:"read_count"`
		Sites          []json.RawMessage `json:"sites"`
		ReceiverFields []string          `json:"receiver_fields"`
		PresentFields  []string          `json:"present_fields"`
	} `json:"pairs"`
}

func intersectionOriginalInputs(t *testing.T) (string, intersectionOriginalManifest) {
	t.Helper()
	declarations := os.Getenv("ADAMIC_INTERSECTION_ORIGINAL_DECLS")
	if declarations == "" {
		t.Skip("set ADAMIC_INTERSECTION_ORIGINAL_DECLS to pinned declaration output")
	}
	data, err := os.ReadFile(checkedViewFixturePath(filepath.Join(declarations, "intersection-manifest.json")))
	if err != nil {
		t.Fatal(err)
	}
	var manifest intersectionOriginalManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	reads := map[int]int{10236: 1, 7612: 4, 9476: 9, 9474: 11, 9485: 9, 9475: 4, 9454: 1, 8883: 4, 8882: 3, 7642: 10, 36241: 1, 7644: 1, 92175: 1, 9657: 1}
	if manifest.Commit != "050880ce59e30b356b686bd3144efe24f875ebc8" || len(manifest.Declarations) != 78 || len(manifest.Pairs) != len(reads) || manifest.Pairs[0].ID != 10236 {
		t.Fatal("original provenance changed")
	}
	for _, pair := range manifest.Pairs {
		if reads[pair.ID] != pair.Reads || len(pair.Sites) != pair.Reads {
			t.Fatalf("original read sites changed for %d.%s", pair.ID, pair.Field)
		}
	}
	for name, digest := range manifest.Declarations {
		data, err := os.ReadFile(checkedViewFixturePath(filepath.Join(declarations, name)))
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
	input, err := os.ReadFile(checkedViewFixturePath("../../stage3/interface-downcasts/lane7/original/" + name + ".a"))
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
		{"tracker-wrong", "true\n", "cast failed: field read failed: tracker?.moduleResolverHost.useCaseSensitiveFileNames is not a () => boolean; expected () => boolean, found string"},
		{"tracker-helpers-wrong", "true\n", "cast failed: field read failed: tracker?.moduleResolverHost.useCaseSensitiveFileNames is not a () => boolean; expected () => boolean, found string"},
		{"tracker-missing", "true\n", "cast failed: field read failed: tracker?.moduleResolverHost.useCaseSensitiveFileNames is not initialized; expected () => boolean, found missing"},
		{"tracker-root-wrong", "true\n", "cast failed: field read failed: tracker?.moduleResolverHost matches no member of (ModuleSpecifierResolutionHost & { getCommonSourceDirectory(): string; }) | undefined; expected (ModuleSpecifierResolutionHost & { getCommonSourceDirectory(): string; }) | undefined, found string"},
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
	requireIntersectionOriginalComplete(t, program, manifest, []string{"Identifier", "Node", "Symbol", "PropertyAccessEntityNameExpression", "ElementAccessExpression"}, ids...)
}

// Every named original interface keeps its complete field set, and each pair's
// declared contract is admitted whole by the bounded walk.
func requireIntersectionOriginalComplete(t *testing.T, program *ir.Program, manifest intersectionOriginalManifest, names []string, ids ...int) {
	t.Helper()
	sorted := func(c ir.ViewContract) []string {
		fields := []string{}
		for _, f := range c.Fields {
			fields = append(fields, f.Name)
		}
		slices.Sort(fields)
		return fields
	}
	for _, name := range names {
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
		{"bindable-static-left-wrong", "true\n", "cast failed: field read failed: node.left.name.symbol is not initialized; expected Symbol, found missing", static[:1]},
		{"bindable-static-helpers-wrong", "true\n", "cast failed: field read failed: node.left.name.symbol is not initialized; expected Symbol, found missing", static[:1]},
		{"bindable-static-left-kind", "true\n", "cast failed: field read failed: node.left.kind expected SyntaxKind.PropertyAccessExpression | SyntaxKind.ElementAccessExpression, found number 999", static[:1]},
		{"bindable-static-expression-wrong", "true\n", "cast failed: field read failed: node.left.expression.expression.symbol is not initialized; expected Symbol, found missing", static},
		{"bindable-static-expression-this", "true\n", "cast failed: field read failed: node.left.expression.expression.kind expected SyntaxKind.Identifier | SyntaxKind.PropertyAccessExpression, found number 110", static},
		{"bindable-element-argument-wrong", "true\n", "cast failed: field read failed: node.left.argumentExpression.kind expected SyntaxKind.NumericLiteral | SyntaxKind.StringLiteral | SyntaxKind.NoSubstitutionTemplateLiteral, found number 80", static},
		{"bindable-access-good", "true\n", "", access},
		{"bindable-access-chain-good", "true\n", "", access},
		{"bindable-access-left-wrong", "true\n", "cast failed: field read failed: node.left.expression.kind expected SyntaxKind.Identifier | SyntaxKind.PropertyAccessExpression | SyntaxKind.ElementAccessExpression, found number 110", access[:1]},
		{"bindable-access-left-symbol", "true\n", "cast failed: field read failed: node.left.expression.symbol is not initialized; expected Symbol, found missing", access[:1]},
		{"bindable-access-left-kind", "true\n", "cast failed: field read failed: node.left.kind expected SyntaxKind.PropertyAccessExpression | SyntaxKind.ElementAccessExpression, found number 999", access[:1]},
		{"bindable-access-expression-wrong", "true\n", "cast failed: field read failed: node.left.expression.expression.symbol is not initialized; expected Symbol, found missing", access},
		{"bindable-access-expression-this", "true\n", "cast failed: field read failed: node.left.expression.expression.kind expected SyntaxKind.Identifier | SyntaxKind.PropertyAccessExpression, found number 110", access},
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

// Each mutant removes one obligation from one pair's read: skip drops the read's
// conjunctive dispatch (a union falls back to its discriminant), shape accepts a
// wrong value, nested forgets a descendant field, presence makes a required field
// optional.
func intersectionOriginalBindableMutant(t *testing.T, program *ir.Program, kind string) {
	t.Helper()
	roots := map[string]string{"left": "BindableStaticAccessExpression", "expression": "EntityNameExpression | (LeftHandSideExpression & BindableStaticNameExpression)", "access": "BindableAccessExpression",
		"emit": "EmitNode & { autoGenerate: AutoGenerateInfo; }", "class": "ExpressionWithTypeArguments & { readonly expression: Identifier | PropertyAccessEntityNameExpression; }", "jsdoc": "HasJSDoc", "root": "HasJSDoc"}
	tags := map[string][2]any{"left": {"SyntaxKind.PropertyAccessExpression", 999.0}, "expression": {"SyntaxKind.Identifier", 110.0}, "access": {"SyntaxKind.ElementAccessExpression", 999.0}, "class": {"SyntaxKind.ExpressionWithTypeArguments", 999.0}, "jsdoc": {"SyntaxKind.EmptyStatement", 999.0}, "root": {"SyntaxKind.EmptyStatement", 999.0}}
	pair, change, _ := strings.Cut(kind, "-")
	changed := 0
	for i := range program.ViewContracts {
		c := &program.ViewContracts[i]
		switch {
		case change == "skip":
			if c.Name == roots[pair] && c.IntersectionBounded {
				c.IntersectionBounded, c.IntersectionTag, c.Intersection = false, "", false
				changed++
			}
		case change == "shape" && pair == "emit":
			if c.Name != "AutoGenerateInfo" {
				continue
			}
			for j, f := range c.Fields {
				if f.Name == "id" {
					child := program.ViewContracts[f.Contract-1]
					child.Of, child.Name = ir.String, "string"
					c.Fields = slices.Clone(c.Fields)
					c.Fields[j].Contract = ir.ViewContractID(len(program.ViewContracts) + 1)
					program.ViewContracts = append(program.ViewContracts, child)
					c = &program.ViewContracts[i]
					changed++
				}
			}
		case change == "shape":
			if c.Name == tags[pair][0] && c.Kind == ir.ViewScalar {
				c.Allowed = append(slices.Clone(c.Allowed), ir.ViewLiteral{Of: ir.Number, Number: tags[pair][1].(float64)})
				changed++
			}
		case change == "nested" && pair == "jsdoc":
			if c.Name == "FunctionDeclaration" {
				c.Fields = slices.DeleteFunc(slices.Clone(c.Fields), func(f ir.ViewFieldContract) bool { return f.Name == "parameters" })
				changed++
			}
		case change == "nested" && pair == "emit":
			if c.Name == "SourceMapRange" {
				c.Fields = slices.DeleteFunc(slices.Clone(c.Fields), func(f ir.ViewFieldContract) bool { return f.Name == "pos" })
				changed++
			}
		case change == "nested":
			if c.Name == "Identifier" || c.Name == "LeftHandSideExpression & Identifier" {
				c.Fields = slices.DeleteFunc(slices.Clone(c.Fields), func(f ir.ViewFieldContract) bool { return f.Name == "symbol" })
				changed++
			}
		case change == "presence":
			if c.Name == roots[pair] {
				for j, f := range c.Fields {
					if f.Name == "autoGenerate" {
						c.Fields = slices.Clone(c.Fields)
						c.Fields[j].Optional = true
						changed++
					}
				}
			}
		}
	}
	if changed == 0 {
		t.Fatal("original mutant found no obligation: " + kind)
	}
}

func TestCheckedViewIntersectionOriginalNodes(t *testing.T) {
	declarations, manifest := intersectionOriginalInputs(t)
	emit := []string{"EmitNode", "AutoGenerateInfo", "GeneratedIdentifier", "SourceMapRange"}
	class := []string{"Identifier", "Symbol", "ExpressionWithTypeArguments"}
	jsdoc := []string{"JSDoc", "Node"}
	for _, test := range []struct {
		name, source, diagnostic string
		names                    []string
		pairs                    []int
	}{
		{"emit-original-good", "0\n", "", emit, []int{7612}},
		{"emit-original-ranges-good", "0\n", "", emit, []int{7612}},
		{"emit-original-missing", "0\n", "cast failed: field read failed: node.emitNode.autoGenerate is not initialized; expected AutoGenerateInfo, found missing", emit, []int{7612}},
		{"emit-original-wrong", "0\n", "cast failed: field read failed: node.emitNode.autoGenerate.id is not a number; expected number, found string", emit, []int{7612}},
		{"emit-original-range-wrong", "0\n", "cast failed: field read failed: node.emitNode.sourceMapRange.pos is not a number; expected number, found string", emit, []int{7612}},
		{"class-augments-good", "true true\n", "", append(slices.Clone(class), "JSDocAugmentsTag"), []int{8883}},
		{"class-implements-good", "true true\n", "", append(slices.Clone(class), "JSDocImplementsTag"), []int{8882}},
		{"class-augments-kind", "true\n", "cast failed: field read failed: node.class.kind expected SyntaxKind.ExpressionWithTypeArguments, found number 999", class, []int{8883}},
		{"class-augments-name", "true\n", "cast failed: field read failed: node.class.expression.kind expected SyntaxKind.Identifier | SyntaxKind.PropertyAccessExpression, found number 110", class, []int{8883}},
		{"class-augments-symbol", "true\n", "cast failed: field read failed: node.class.expression.symbol is not initialized; expected Symbol, found missing", class, []int{8883}},
		{"class-implements-kind", "true\n", "cast failed: field read failed: n.class.kind expected SyntaxKind.ExpressionWithTypeArguments, found number 999", class, []int{8882}},
		{"class-implements-symbol", "true\n", "cast failed: field read failed: n.class.expression.symbol is not initialized; expected Symbol, found missing", class, []int{8882}},
		{"jsdoc-parent-good", "true\n", "", jsdoc, []int{7642}},
		{"jsdoc-parent-function-good", "true\n", "", append(slices.Clone(jsdoc), "FunctionDeclaration"), []int{7642}},
		{"jsdoc-parent-kind", "true\n", "cast failed: field read failed: jsDoc.parent.kind expected SyntaxKind.EndOfFileToken | SyntaxKind.Identifier | SyntaxKind.TypeParameter | SyntaxKind.Parameter | SyntaxKind.PropertySignature | SyntaxKind.PropertyDeclaration | SyntaxKind.MethodSignature | ... 58 more ... | SyntaxKind.JSDocSignature, found number 999", jsdoc, []int{7642}},
		{"jsdoc-parent-parameters", "true\n", "cast failed: field read failed: jsDoc.parent.parameters is not initialized; expected NodeArray<ParameterDeclaration>, found missing", jsdoc, []int{7642}},
		{"jsdoc-parent-symbol", "true\n", "cast failed: field read failed: jsDoc.parent.name.symbol is not initialized; expected Symbol, found missing", jsdoc, []int{7642}},
		{"jsdoc-root-good", "true false\n", "", jsdoc, []int{36241}},
		{"jsdoc-root-symbol", "true\n", "cast failed: field read failed: root(found)?.parent.name.symbol is not initialized; expected Symbol, found missing", jsdoc, []int{36241}},
		{"jsdoc-root-kind", "true\n", "cast failed: field read failed: root(found)?.parent.kind expected SyntaxKind.EndOfFileToken | SyntaxKind.Identifier | SyntaxKind.TypeParameter | SyntaxKind.Parameter | SyntaxKind.PropertySignature | SyntaxKind.PropertyDeclaration | SyntaxKind.MethodSignature | ... 58 more ... | SyntaxKind.JSDocSignature, found number 999", jsdoc, []int{36241}},
		{"class-implements-arguments", "true\n", "cast failed: field read failed: n.class.typeArguments is not a NodeArray<TypeNode> | undefined; expected NodeArray<TypeNode> | undefined, found string", class, []int{8882}},
	} {
		t.Run(test.name, func(t *testing.T) {
			program, _ := intersectionOriginalProgram(t, declarations, test.name, test.source)
			requireIntersectionOriginalComplete(t, program, manifest, test.names, test.pairs...)
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
