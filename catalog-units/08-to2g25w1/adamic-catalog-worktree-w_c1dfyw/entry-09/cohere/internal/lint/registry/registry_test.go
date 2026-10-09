package registry

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
)

// Every rule declared in the rules packages must be in the live catalog.
//
// A rule that exists on disk and never registers compiles, tests green in its own package, and
// never runs. That is the worst defect this tool can carry: the file says the rule is enforced, the
// gate says the tree is clean, and neither is lying about what it knows.
//
// This was not hypothetical. Two boundary rules were written, tested, and left out of the registry;
// nothing failed until this test existed.
//
// The comparison is against `rule.Registered()` rather than against the text of `registry.go`. It
// used to read that file for the identifiers a hand-maintained list mentioned, which was the only
// thing available when the list was written by hand. Reading the live catalog is strictly stronger:
// a mention proved a name appeared in a file, while a registration proves the rule is loaded into
// the process that will run it. It also catches the failure mode self-registration introduces,
// where a rule package nothing imports contributes nothing and does not fail to compile.
func TestEveryRuleIsRegistered(t *testing.T) {
	t.Parallel()
	declared := declaredRules(t)
	registered := registeredRuleNames()

	if len(declared) == 0 {
		// A test that finds nothing to check passes for the wrong reason, which is the same shape
		// as the defect it guards against.
		t.Fatal("found no rule declarations to check, so this test proved nothing")
	}

	live := make(map[string]bool, len(registered))
	for _, name := range registered {
		live[name] = true
	}

	for _, declaration := range declared {
		if !live[declaration.ruleName] {
			t.Errorf("rule %s declares the name %q and nothing registered it, so it would never run",
				declaration.identifier, declaration.ruleName)
		}
	}
}

// ruleDeclaration is one `var Name = rule.Rule{...}` found on disk.
//
// Both halves are needed. The identifier is what a reader greps for when the test fails, and the
// declared Name is the only thing the catalog can be checked against, since registration keys on
// the rule's name rather than on the Go identifier that happens to hold it.
type ruleDeclaration struct {
	identifier string
	ruleName   string
}

// declaredRules parses the rules packages for `var Name = rule.Rule{...}` declarations.
//
// Parsing rather than reflecting: a rule that never registers is also absent from anything the
// catalog could reflect over, so reflection would see exactly the rules that are already fine.
func declaredRules(t *testing.T) []ruleDeclaration {
	t.Helper()

	var declarations []ruleDeclaration
	matches, err := filepath.Glob("../rules/*/*.go")
	if err != nil {
		t.Fatalf("globbing rule files: %v", err)
	}

	for _, path := range matches {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}

		fileSet := token.NewFileSet()
		parsed, parseErr := parser.ParseFile(fileSet, path, nil, 0)
		if parseErr != nil {
			t.Fatalf("parsing %s: %v", path, parseErr)
		}

		for _, declaration := range parsed.Decls {
			general, isGeneral := declaration.(*ast.GenDecl)
			if !isGeneral || general.Tok != token.VAR {
				continue
			}
			for _, spec := range general.Specs {
				value, isValue := spec.(*ast.ValueSpec)
				if !isValue || len(value.Names) != 1 || len(value.Values) != 1 {
					continue
				}
				if !isRuleLiteral(value.Values[0]) {
					continue
				}
				name := value.Names[0].Name
				if !ast.IsExported(name) {
					continue
				}
				declaredName, named := ruleNameIn(value.Values[0])
				if !named {
					// A rule literal with no Name field cannot be addressed by the config at all,
					// which is a different defect and one the compiler will not catch either.
					t.Errorf("rule %s.%s declares no Name, so no configuration could address it",
						parsed.Name.Name, name)
					continue
				}
				declarations = append(declarations, ruleDeclaration{
					identifier: parsed.Name.Name + "." + name,
					ruleName:   declaredName,
				})
			}
		}
	}
	return declarations
}

// ruleNameIn reads the Name field out of a `rule.Rule{...}` literal.
func ruleNameIn(expression ast.Expr) (string, bool) {
	composite, isComposite := expression.(*ast.CompositeLit)
	if !isComposite {
		return "", false
	}
	for _, element := range composite.Elts {
		keyed, isKeyed := element.(*ast.KeyValueExpr)
		if !isKeyed {
			continue
		}
		key, isIdentifier := keyed.Key.(*ast.Ident)
		if !isIdentifier || key.Name != "Name" {
			continue
		}
		literal, isLiteral := keyed.Value.(*ast.BasicLit)
		if !isLiteral || literal.Kind != token.STRING {
			return "", false
		}
		return strings.Trim(literal.Value, `"`), true
	}
	return "", false
}

// isRuleLiteral reports whether an expression is a `rule.Rule{...}` composite literal.
func isRuleLiteral(expression ast.Expr) bool {
	composite, isComposite := expression.(*ast.CompositeLit)
	if !isComposite {
		return false
	}
	selector, isSelector := composite.Type.(*ast.SelectorExpr)
	if !isSelector {
		return false
	}
	packageIdentifier, isIdentifier := selector.X.(*ast.Ident)
	return isIdentifier && packageIdentifier.Name == "rule" && selector.Sel.Name == "Rule"
}

// The registry must not contain duplicates: a rule listed twice reports every finding twice.
func TestNoDuplicateRules(t *testing.T) {
	t.Parallel()
	seen := make(map[string]bool)
	for _, subject := range All() {
		if seen[subject.Name] {
			t.Errorf("rule %q appears in the registry more than once", subject.Name)
		}
		seen[subject.Name] = true
	}
}

// Count must reflect what All actually returns, since the coverage line reports it and a coverage
// line that overstates what ran is the failure this tool exists to prevent.
func TestCountMatchesAll(t *testing.T) {
	t.Parallel()
	if Count() != len(All()) {
		t.Errorf("Count() = %d but All() returned %d rules", Count(), len(All()))
	}
}
