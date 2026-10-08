package oracle

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckedViewIntersectionOriginalIdentifier(t *testing.T) {
	declarations, manifest := intersectionOriginalInputs(t)
	assigned := manifest.AssignedIdentifier
	if assigned.ID != 9477 || assigned.Reads != 1 || assigned.Field != "escapedText" || assigned.Receiver != "LeftHandSideExpression & Identifier" || len(assigned.SourceSHA) != 64 {
		t.Fatal("assigned original Identifier provenance missing")
	}
	for _, variant := range []string{"good", "internal", "undefined", "wrong", "null", "missing", "direct", "alias-write", "helper-only", "helper-only-good"} {
		t.Run(variant, func(t *testing.T) {
			fixture := variant
			if variant == "helper-only-good" {
				fixture = "good"
			}
			if variant == "direct" || variant == "alias-write" || variant == "helper-only" {
				fixture = "wrong"
			}
			input, err := os.ReadFile(checkedViewFixturePath("../../stage3/interface-downcasts/lane4/original/lefthandsideexpression & identifier-" + fixture + ".a"))
			if err != nil {
				t.Fatal(err)
			}
			source := strings.Replace(string(input), "child: {kind: 80,", "child: {symbol: {flags: 16, escapedName: 'word', id: 1, mergeId: 0, constEnumOnlyModule: false}, pos: 0, end: 10, flags: 0, modifierFlagsCache: 0, transformFlags: 0, parent: {pos: 0, end: 11, kind: 308, flags: 0, modifierFlagsCache: 0, transformFlags: 0, parent: {}}, kind: 80,", 1)
			if variant == "direct" {
				source = strings.Replace(source, "helper(carrier.child);", "const direct: object = raw.child; helper(direct as LeftHandSideExpression & Identifier);", 1)
			}
			if variant == "alias-write" {
				source = strings.Replace(source, "escapedText: 42", "escapedText: ('word' + '-built') as string | number | object", 1)
				source = strings.Replace(source, "helper(carrier.child);", "const child = carrier.child; raw.child.escapedText = 42; helper(child);", 1)
			}
			if variant == "helper-only" || variant == "helper-only-good" {
				source = strings.Replace(source, "LeftHandSideExpression, Identifier", "LeftHandSideExpression, Identifier, SyntaxKind", 1)
				source = strings.Replace(source, "helper(carrier.child);", "interface DirectBase {readonly kind: SyntaxKind} const direct: DirectBase = raw.child; helper(direct as Identifier);", 1)
			}
			source = strings.Replace(source, "'original-tsc-types'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(declarations, "compiler/types.d.ts"))), 1)
			path := filepath.Join(t.TempDir(), "identifier-"+variant+".a")
			if err := os.WriteFile(path, []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
			output := map[string]string{"good": "word-built", "internal": "__importAttributes", "undefined": "undefined", "wrong": "42", "direct": "42", "alias-write": "42", "helper-only": "42", "helper-only-good": "word-built", "null": "null", "missing": "undefined"}[variant] + "\n"
			if difference := disagreement(run{stdout: []byte(output)}, onNode(t, path)); difference != "" {
				t.Fatal("Node: " + difference)
			}
			program, err := lowered(t, path)
			if variant == "direct" {
				t.Logf("direct admission result: %v", err)
				if err == nil || !strings.HasSuffix(err.Error(), "(adamic/no-unchecked-cast)") {
					t.Fatalf("expected named direct intersection cast refusal, got %v", err)
				}
				return
			}
			if variant == "alias-write" {
				suffix := "Adamic 0.1 refuses checked view read of field escapedText with unsupported representation conversion contract; prove or implement the representation conversion contract before reading this field"
				if err == nil || !strings.HasSuffix(err.Error(), suffix) {
					t.Fatalf("expected named wider-alias refusal, got %v", err)
				}
				t.Log(err)
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			requireIntersectionOriginalComplete(t, program, manifest, []string{"LeftHandSideExpression", "Identifier", "Node"})
			bounded := false
			for _, contract := range program.ViewContracts {
				if contract.Name == "LeftHandSideExpression & Identifier" && contract.IntersectionBounded && contract.Unsupported == "" {
					bounded = true
				}
			}
			if !bounded {
				t.Fatal("original intersected receiver not bounded")
			}
			if variant == "helper-only" || variant == "helper-only-good" {
				recordIntersectionOriginalCounts(t, path, "9477-"+variant)
			}
			if (variant == "wrong" || variant == "helper-only") && os.Getenv("ADAMIC_INTERSECTION_IDENTIFIER_MUTANT") != "" {
				index := len(program.Strings)
				program.Strings = append(program.Strings, "unchecked")
				replacement := ir.Concat{Parts: []ir.Expression{ir.StringConstant{Index: index}, ir.StringConstant{Index: index}}}
				if count := dropLane4NamedHelperView(program, "escapedText", replacement); count != 1 {
					t.Fatalf("expected one member-read mutation, got %d", count)
				}
			}
			want := run{stdout: []byte(output)}
			if variant == "wrong" || variant == "null" || variant == "helper-only" {
				found := map[string]string{"wrong": "number", "null": "null", "helper-only": "number"}[variant]
				expression := "carrier.child.escapedText"
				if variant == "helper-only" {
					expression = "value.escapedText"
				}
				if variant == "null" {
					expression = "carrier.child.escapedText"
				}
				want = run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: " + expression + " is not a __String; expected __String, found " + found + "\n")}
			}
			if variant == "missing" {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: carrier.child.escapedText is not initialized; expected __String, found missing\n")}
			}
			actual, binary := nativelyUncached(t, program)
			if want.exitCode == 0 {
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := viewReadDisagreement(want, got, program); difference != "" {
					t.Errorf("%s; got %#v", difference, got)
				}
			}
		})
	}
}

// Not parallel: update only this unit's original-declaration count rows.
func recordIntersectionOriginalCounts(t *testing.T, fixture, label string) {
	t.Helper()
	root, err := filepath.Abs(checkedViewFixturePath(repository))
	if err != nil {
		t.Fatal(err)
	}
	relative, err := filepath.Rel(root, fixture)
	if err != nil {
		t.Fatal(err)
	}
	row := strings.Replace(counted(t, checkedViewFixturePath(relative), false, nil, false, false), relative, "original/"+label+".a", 1)
	path := filepath.Join(repository, "stage3/interface-downcasts/lane7/counts.md")
	data, err := os.ReadFile(checkedViewFixturePath(path))
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if !*updateCounts {
		if !strings.Contains(string(data), row+"\n") {
			t.Fatalf("original Identifier counts moved: %s", row)
		}
		return
	}
	text := string(data)
	if text == "" {
		text = "# Lane 7 original-declaration counts\n\nScoped rows measured from generated .a controls with complete pinned original declarations. Labels identify generated cases, not repository file paths. Refresh using the original oracle with -args -update-counts.\n\n" + countsHeader[strings.Index(countsHeader, "| Fixture |"):]
	}
	lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
	key := strings.Split(row, " | ")[0] + " | "
	found := false
	for i, line := range lines {
		if strings.HasPrefix(line, key) {
			lines[i] = row
			found = true
			break
		}
	}
	if !found {
		lines = append(lines, row)
	}
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
}
