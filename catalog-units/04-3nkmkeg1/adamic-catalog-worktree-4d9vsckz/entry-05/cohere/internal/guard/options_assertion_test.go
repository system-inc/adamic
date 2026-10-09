package guard

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"
)

// A rule reads its options through rule.OptionsAs, never through a bare type assertion, and this is
// what makes that structural rather than remembered.
//
// Run takes `options any`, so the compiler never compares the type a registration's Decode returns
// with the type the rule asserts, and a comma-ok assertion cannot fail loudly. `unified-signatures`
// asserted a pointer over the value `DecodeOptionsInto` returns: the assertion missed on every
// configured run, both options read false whatever the config said, and every fixture stayed green
// because the fixtures handed Run the pointer directly (#qmvkf83). rule.OptionsAs keeps the comma-ok
// shape for the one honest miss, no options configured, and panics on any other mismatch, so a
// disagreement becomes a crashed file and a failing test instead of a silent default.
//
// The check reads syntax: a type assertion or type switch on a function's options parameter, under
// any name, in a rule package. A parameter is an options parameter when it is typed `any` and sits
// beside a rule.Context, which is Run's shape, or when it is named `options` and typed `any`, which
// is the shape of a helper that resolves settings for its rule.
func TestRulesReadOptionsThroughOptionsAs(t *testing.T) {
	t.Parallel()

	ruleFiles := ruleSourceFiles(t)
	if len(ruleFiles) == 0 {
		t.Fatal("found no rule source files, so this test proved nothing")
	}

	sawOptionsReader := false
	for _, path := range ruleFiles {
		facts := readOptionsFacts(t, path)
		if facts.readsThroughOptionsAs {
			sawOptionsReader = true
		}
		for _, position := range facts.bareAssertions {
			t.Errorf("%s asserts its options with a bare type assertion; read them with "+
				"rule.OptionsAs[T](options), so a decoder handing over a different type fails loudly "+
				"instead of every configured option being ignored", position)
		}
	}

	if !sawOptionsReader {
		// About two hundred rules read options today. Zero would mean the search is looking in the
		// wrong place, or at the wrong call, rather than that no rule takes options.
		t.Fatal("no rule source calls rule.OptionsAs, which is implausible: this check is looking " +
			"at the wrong files")
	}
}

// The guard sees the defect it exists for, in each shape it can take, and passes the repair.
//
// Each source is written to a file and read the way the guard reads the tree, so the guard is shown
// able to fail rather than only shown passing.
func TestReadOptionsFactsSeesBareAssertions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		source      string
		bare        int
		optionsAsOk bool
	}{
		{
			// unified-signatures before #qmvkf83, verbatim in shape.
			name: "a pointer asserted over a value decoder",
			source: `package probe
var Probe = rule.Rule{Run: func(ctx rule.Context, options any) rule.Listeners {
	settings, _ := options.(*ProbeOptions)
	_ = settings
	return nil
}}`,
			bare: 1,
		},
		{
			name: "a value assertion",
			source: `package probe
var Probe = rule.Rule{Run: func(ctx rule.Context, options any) rule.Listeners {
	if settings, ok := options.(ProbeOptions); ok { _ = settings }
	return nil
}}`,
			bare: 1,
		},
		{
			// A rule may call its options anything; what makes it the options is Run's shape.
			name: "an options parameter under another name",
			source: `package probe
var Probe = rule.Rule{Run: func(c rule.Context, raw any) rule.Listeners {
	settings, _ := raw.(ProbeOptions)
	_ = settings
	return nil
}}`,
			bare: 1,
		},
		{
			// no_use_before_define's shape: a helper resolving settings, with a type switch.
			name: "a type switch in a settings helper",
			source: `package probe
func resolveProbe(options any) int {
	switch options.(type) {
	case ProbeOptions:
		return 1
	}
	return 0
}`,
			bare: 1,
		},
		{
			name: "the repair",
			source: `package probe
var Probe = rule.Rule{Run: func(ctx rule.Context, options any) rule.Listeners {
	settings, ok := rule.OptionsAs[ProbeOptions](options)
	_, _ = settings, ok
	return nil
}}`,
			optionsAsOk: true,
		},
		{
			// An assertion on some other value is not an options read.
			name: "an assertion on a node",
			source: `package probe
func probe(node any) bool { _, ok := node.(string); return ok }`,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "probe.go")
			if err := os.WriteFile(path, []byte(testCase.source), 0o644); err != nil {
				t.Fatal(err)
			}
			facts := readOptionsFacts(t, path)
			if len(facts.bareAssertions) != testCase.bare || facts.readsThroughOptionsAs != testCase.optionsAsOk {
				t.Errorf("got %d bare assertions and OptionsAs %v, want %d and %v",
					len(facts.bareAssertions), facts.readsThroughOptionsAs, testCase.bare, testCase.optionsAsOk)
			}
		})
	}
}

// optionsFacts is how one rule file reads its options, from its syntax.
type optionsFacts struct {
	// bareAssertions are the positions of type assertions and type switches on an options parameter.
	bareAssertions []string
	// readsThroughOptionsAs is a call of rule.OptionsAs.
	readsThroughOptionsAs bool
}

// readOptionsFacts parses one file and reports how it reads its options.
func readOptionsFacts(t *testing.T, path string) optionsFacts {
	t.Helper()

	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, path, nil, 0)
	if err != nil {
		t.Fatalf("parsing %s: %v", path, err)
	}

	// Options parameter names, collected per function signature: an `any` parameter beside a
	// rule.Context, or one named options.
	optionsNames := map[string]bool{}
	ast.Inspect(file, func(node ast.Node) bool {
		function, isFunction := node.(*ast.FuncType)
		if !isFunction || function.Params == nil {
			return true
		}
		takesContext := false
		for _, field := range function.Params.List {
			if isRuleSelector(field.Type, "Context") {
				takesContext = true
			}
		}
		for _, field := range function.Params.List {
			typeName, isIdentifier := field.Type.(*ast.Ident)
			if !isIdentifier || (typeName.Name != "any" && typeName.Name != "interface{}") {
				continue
			}
			for _, name := range field.Names {
				if takesContext || name.Name == "options" {
					optionsNames[name.Name] = true
				}
			}
		}
		return true
	})

	facts := optionsFacts{}
	ast.Inspect(file, func(node ast.Node) bool {
		switch typed := node.(type) {
		case *ast.TypeAssertExpr:
			// A type switch's `x.(type)` is a TypeAssertExpr with no Type, and it is counted too:
			// its default arm defaults just as silently.
			if receiver, isIdentifier := typed.X.(*ast.Ident); isIdentifier && optionsNames[receiver.Name] {
				facts.bareAssertions = append(facts.bareAssertions, fileSet.Position(typed.Pos()).String())
			}
		case *ast.IndexExpr:
			if isRuleSelector(typed.X, "OptionsAs") {
				facts.readsThroughOptionsAs = true
			}
		}
		return true
	})
	return facts
}
