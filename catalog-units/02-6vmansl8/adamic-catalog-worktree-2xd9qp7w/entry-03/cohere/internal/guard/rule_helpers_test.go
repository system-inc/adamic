package guard

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// A rule package holding its own helpers is where the catalog drifts, and this is the census that
// makes it visible.
//
// Measured when this was written: 16 files across 6 rule packages contain no rule declaration at
// all. They are helpers, shared inside one namespace and invisible to every other. `next/attributes.go`
// decides how a JSX attribute is matched; when `react` is ported it will decide again, differently,
// and neither decision will ever be reconciled because both packages pass their own tests.
//
// The house rule is already explicit: two helpers that do almost the same thing is worse than one
// that is slightly the wrong shape, because the drift is permanent.
//
// This test does not fail on a helper existing. A namespace-local helper is often correct, and a
// guard that refused them outright would be worked around within a week. What it refuses is a
// helper whose name matches something the shared utility layer already exports, because that is a
// second implementation of a decision that already has a home.
func TestRulePackagesDoNotShadowSharedUtilities(t *testing.T) {
	t.Parallel()

	shared := exportedUtilityNames(t)
	if len(shared) == 0 {
		// A census with nothing to compare against passes for the wrong reason, which is the same
		// shape as the defect it guards against. The utility layer is never legitimately empty.
		t.Fatal("found no exported utilities, so this test compared nothing")
	}

	local := ruleLocalFunctionNames(t)
	if len(local) == 0 {
		t.Fatal("found no rule-package functions, so this test compared nothing")
	}

	for _, candidate := range local {
		// Compared case-insensitively on the bare name. A rule package writes `hasAttributeNamed`
		// unexported while the shelf exports `HasAttributeNamed`, and those are the same decision
		// spelled for two different scopes, which is exactly the collision worth catching.
		if owner, taken := shared[strings.ToLower(candidate.name)]; taken {
			t.Errorf("%s declares %s, which the shared layer already provides as %s; "+
				"two implementations of one decision drift permanently, and both will keep passing "+
				"their own tests while they do",
				candidate.file, candidate.name, owner)
		}
	}
}

// TestRulePackageHelperCensus reports which rule packages carry helper files.
//
// It never fails. The number is the point: it is the size of the retrofit, it is not visible in any
// rule count, and it grows silently every time a porter needs something the shelf does not have
// yet. Printed with `-v` so the figure in a task body can be re-derived rather than remembered.
func TestRulePackageHelperCensus(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob("../lint/rules/*/*.go")
	if err != nil {
		t.Fatalf("globbing rule files: %v", err)
	}

	byPackage := map[string][]string{}
	for _, path := range files {
		base := filepath.Base(path)
		if strings.HasSuffix(base, "_test.go") || base == "register.go" {
			continue
		}
		if declaresARule(t, path) {
			continue
		}
		packageName := filepath.Base(filepath.Dir(path))
		byPackage[packageName] = append(byPackage[packageName], base)
	}

	names := make([]string, 0, len(byPackage))
	total := 0
	for packageName, helpers := range byPackage {
		names = append(names, packageName)
		total += len(helpers)
	}
	sort.Strings(names)

	t.Logf("helper files inside rule packages: %d across %d packages", total, len(names))
	for _, packageName := range names {
		sort.Strings(byPackage[packageName])
		t.Logf("  %-12s %s", packageName, strings.Join(byPackage[packageName], " "))
	}
}

// ruleLocalFunction is one function declared in a rule package, with where it lives.
type ruleLocalFunction struct {
	file string
	name string
}

// ruleLocalFunctionNames returns every function declared in a rule package's helper files.
//
// Only files that declare no rule are read. A helper living beside the rule that uses it, inside a
// file that is that rule, is not the drift this looks for.
func ruleLocalFunctionNames(t *testing.T) []ruleLocalFunction {
	t.Helper()

	files, err := filepath.Glob("../lint/rules/*/*.go")
	if err != nil {
		t.Fatalf("globbing rule files: %v", err)
	}
	if len(files) == 0 {
		// A glob that matched nothing finds no helper to refuse, which passes for the wrong reason.
		t.Fatal("found no rule files, so this test read nothing")
	}

	var found []ruleLocalFunction
	for _, path := range files {
		base := filepath.Base(path)
		if strings.HasSuffix(base, "_test.go") || base == "register.go" {
			continue
		}
		if declaresARule(t, path) {
			continue
		}
		for _, name := range functionNamesIn(t, path) {
			found = append(found, ruleLocalFunction{
				file: filepath.Join(filepath.Base(filepath.Dir(path)), base),
				name: name,
			})
		}
	}
	return found
}

// exportedUtilityNames returns every exported function the shared utility layer provides, keyed by
// lowercased name so a comparison can ignore the exported/unexported spelling.
func exportedUtilityNames(t *testing.T) map[string]string {
	t.Helper()

	files, err := filepath.Glob("../lint/ecmascript/*/*/*.go")
	if err != nil {
		t.Fatalf("globbing utility files: %v", err)
	}
	nested, err := filepath.Glob("../lint/ecmascript/*/*.go")
	if err != nil {
		t.Fatalf("globbing utility files: %v", err)
	}
	files = append(files, nested...)
	if len(files) == 0 {
		// With no utility names, no helper can shadow one, which passes for the wrong reason.
		t.Fatal("found no shared utility files, so this test read nothing")
	}

	names := map[string]string{}
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		for _, name := range functionNamesIn(t, path) {
			if !ast.IsExported(name) {
				continue
			}
			names[strings.ToLower(name)] = filepath.Base(filepath.Dir(path)) + "." + name
		}
	}
	return names
}

// declaresARule reports whether a file contains a `var Name = rule.Rule{...}` declaration.
func declaresARule(t *testing.T, path string) bool {
	t.Helper()

	parsed := parseGoFile(t, path)
	found := false
	ast.Inspect(parsed, func(node ast.Node) bool {
		composite, isComposite := node.(*ast.CompositeLit)
		if !isComposite {
			return true
		}
		selector, isSelector := composite.Type.(*ast.SelectorExpr)
		if !isSelector {
			return true
		}
		identifier, isIdentifier := selector.X.(*ast.Ident)
		if isIdentifier && identifier.Name == "rule" && selector.Sel.Name == "Rule" {
			found = true
			return false
		}
		return true
	})
	return found
}

// functionNamesIn returns the names of every top-level function in a file, methods excluded.
//
// Methods are excluded because a method's name is scoped by its receiver, so `(*Walker).Next` and a
// shared `Next` are not the same decision and flagging them would be noise.
func functionNamesIn(t *testing.T, path string) []string {
	t.Helper()

	parsed := parseGoFile(t, path)
	var names []string
	for _, declaration := range parsed.Decls {
		function, isFunction := declaration.(*ast.FuncDecl)
		if !isFunction || function.Recv != nil {
			continue
		}
		names = append(names, function.Name.Name)
	}
	return names
}

// parseGoFile parses one file or fails the test naming it.
func parseGoFile(t *testing.T, path string) *ast.File {
	t.Helper()

	parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatalf("parsing %s: %v", path, err)
	}
	return parsed
}

// wrappedAccessor is one raw AST accessor a shared utility exists to answer, and the utility that
// answers it.
type wrappedAccessor struct {
	// call is the source text a rule package writes when it reaches past the shelf.
	call string
	// utility names what to call instead, so the failure tells a porter where to go.
	utility string
	// decision is what the shelf function decides that the raw accessor does not, so a reader can
	// judge whether their case is the exception rather than obeying the guard blindly.
	decision string
}

// wrappedAccessors are the raw reaches that have already been retrofitted away, and must not return.
//
// Every entry here was found in the tree rather than imagined, and each one had been written two or
// more times independently. That is the bar for adding one: this list is not a style preference
// about which spelling is prettier, it is a record of decisions that were duplicated in practice.
var wrappedAccessors = []wrappedAccessor{
	{
		call:     "AsNamedImports()",
		utility:  "imports.BindingsOf",
		decision: "which of the three shapes one import statement carries, kept separate",
	},
	{
		call:     "AsJsxSelfClosingElement().TagName",
		utility:  "jsx.ElementParts",
		decision: "that a self-closing element never produces a JsxOpeningElement",
	},
	{
		call:     "AsJsxOpeningElement().TagName",
		utility:  "jsx.ElementParts",
		decision: "that a self-closing element never produces a JsxOpeningElement",
	},
}

// TestRulePackagesDoNotReachPastWrappedAccessors refuses a rule package re-deriving a decision the
// shared layer already made, when the re-derivation is invisible to a name comparison.
//
// This is the half `TestRulePackagesDoNotShadowSharedUtilities` cannot see, and the gap was measured
// rather than predicted. Six retrofits landed today; not one of them would have tripped the name
// guard, because the local helpers were called `reportForwardRefImport` and `importsGraphqlSpecifier`
// while the shelf exports `BindingsOf` and `ImportedNameOf`. Nothing collided. The name guard
// catches a second implementation that admits what it is by its name, and the expensive case is
// exactly the one that does not.
//
// The signal it uses instead is the reach itself. A rule calling `AsNamedImports()` is walking past
// a function whose whole purpose is that walk, and that is visible in the source without knowing
// what anybody named their helper.
//
// **It deliberately does not try to detect duplicated judgment in general.** A fourth retrofit today
// fit only halfway: `boundary-no-project-theme-value` had the shelf's two-kind element split spelled
// locally, wrapped around an upward walk from an attribute that is genuinely its own and lives
// nowhere else. A guard claiming to find "a local helper that duplicates a shared one" would have to
// rule on that file, and would be wrong whichever way it ruled. This one asks a narrower question it
// can answer exactly.
//
// An exemption is a comment on the reaching line or the line above it, saying why. That keeps the
// escape hatch in the file that needs it rather than in a list here that nobody reads, and it accepts
// the comment where an author naturally writes one: the first exemption this guard demanded was
// written above the line, and requiring it trailing would have made the guard shape the prose.
func TestRulePackagesDoNotReachPastWrappedAccessors(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob("../lint/rules/*/*.go")
	if err != nil {
		t.Fatalf("globbing rule files: %v", err)
	}
	if len(files) == 0 {
		// A sweep with nothing to read passes for the wrong reason, which is the same shape as the
		// defect it guards against.
		t.Fatal("found no rule files, so this test read nothing")
	}

	checked := 0
	for _, path := range files {
		if strings.HasSuffix(filepath.Base(path), "_test.go") {
			continue
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("reading %s: %v", path, err)
		}
		checked++

		lines := strings.Split(string(contents), "\n")
		for lineNumber, line := range lines {
			// A reason on the line itself, or on the line directly above it, is an exemption. Both
			// are places a reader sees the reason where the reach is rather than three files away.
			if strings.Contains(line, "//") {
				continue
			}
			if lineNumber > 0 && strings.Contains(strings.TrimSpace(lines[lineNumber-1]), "//") {
				continue
			}
			for _, accessor := range wrappedAccessors {
				if !strings.Contains(line, accessor.call) {
					continue
				}
				t.Errorf("%s:%d reaches for %s, which %s already answers by deciding %s; "+
					"call the utility, or write the reason on this line if this case is genuinely "+
					"different",
					filepath.Base(path), lineNumber+1, accessor.call, accessor.utility, accessor.decision)
			}
		}
	}

	if checked == 0 {
		t.Fatal("every rule file was skipped, so this test compared nothing")
	}
}
