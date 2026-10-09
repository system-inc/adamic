package program_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/types/program"
)

// TestAliasAccessorsWalkTheChainToADeprecation guards the two alias accessors this shim exposes.
//
// `Checker_getImmediateAliasedSymbol` and `Checker_resolveAlias` were added so a rule can answer one
// question: is the thing this identifier names deprecated, when the deprecation may sit on an alias
// rather than on the declaration the name resolves to. Importing a symbol gives you an alias of the
// exported one, which is itself an alias of the original, and a `@deprecated` tag can be written on
// any link in that chain.
//
// Before these accessors existed the chain was unreachable and the answer was silently WRONG rather
// than absent. Measured on the seeded case below: reading the resolved symbol's own declarations
// answers "not deprecated" for an export whose alias carries the tag. That is a false negative,
// which costs findings and announces nothing.
//
// # What each assertion has to survive
//
// The precedent in this package sets the standard and it applies here: assert what the value IS,
// never that it is present. A nil check on `resolveAlias` passes for a chain that resolves to the
// wrong symbol entirely, and "the walk found a deprecation" passes for a walk that reports one on
// every input. So each case below names the symbol the chain must reach, and the negative case is
// carried alongside the two positives rather than assumed.
//
// The three cases are chosen to fail differently:
//
//	deprecation on the exported ALIAS      only the walk can find it; the resolved symbol says no
//	deprecation on the ORIGINAL            both the walk and the resolved symbol find it
//	deprecation NOWHERE                    both must say no, or the two above prove nothing
//
// Without the third, a walk that answered "deprecated" unconditionally would pass the first two.
//
// # BOTH accessors panic on a non-alias, and the caller-side gate is mandatory
//
// Each one opens by testing `SymbolFlagsAlias` and panicking when it is unset. That reads like a
// guard and is an assertion: `if symbol.Flags&ast.SymbolFlagsAlias == 0 { panic(...) }` at
// checker.go:2162 and 16361. A rule's walk recovers per FILE rather than per rule, so one unguarded
// call costs every rule that file, which is how this project lost 167 files once already.
//
// This is written down because the first version of this test asserted the opposite. It expected
// `resolveAlias` to return a non-alias unchanged, on a reading of that same line as a guard, and the
// test panicked rather than failing, which is the whole argument for asserting instead of printing.
// A probe that printed its result would have shown the panic as a crash with no verdict attached,
// and the wrong belief would have reached the rule.
func TestAliasAccessorsWalkTheChainToADeprecation(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		// otherSource is the imported module, where the deprecation is written.
		otherSource string
		// wantChainFindsDeprecation is what the alias walk must conclude.
		wantChainFindsDeprecation bool
		// wantResolvedFindsDeprecation is what reading the FINAL symbol alone concludes, which is
		// what a rule would answer without these accessors.
		wantResolvedFindsDeprecation bool
	}{
		{
			name: "deprecation on the exported alias",
			otherSource: "function exported(): void {}\n" +
				"export { /** @deprecated moved */ exported };\n",
			wantChainFindsDeprecation:    true,
			wantResolvedFindsDeprecation: false,
		},
		{
			name:                         "deprecation on the original declaration",
			otherSource:                  "/** @deprecated gone */\nexport function exported(): void {}\n",
			wantChainFindsDeprecation:    true,
			wantResolvedFindsDeprecation: true,
		},
		{
			name:                         "deprecation nowhere",
			otherSource:                  "export function exported(): void {}\n",
			wantChainFindsDeprecation:    false,
			wantResolvedFindsDeprecation: false,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			fileChecker, sourceFile, release := buildAliasSubject(t, testCase.otherSource)
			defer release()

			importedUse := aliasSubjectImportedUse(t, fileChecker, sourceFile)

			// The symbol an import produces must BE an alias, or the two accessors below are never
			// reached and every assertion after this is vacuous.
			if importedUse.Flags&ast.SymbolFlagsAlias == 0 {
				t.Fatalf("the imported symbol does not carry SymbolFlagsAlias, so the alias "+
					"accessors are unreachable and this case measures nothing (flags %v)",
					importedUse.Flags)
			}

			resolved := checker.Checker_resolveAlias(fileChecker, importedUse)
			if resolved == nil {
				t.Fatal("Checker_resolveAlias returned nil for a symbol carrying the alias flag")
			}
			// It must reach the ORIGINAL, not merely something. A chain that stopped at the export
			// would satisfy a nil check and answer the wrong question.
			// The import is renamed, so the alias is `renamed` and the original is `exported`.
			// Naming both sides makes a chain that never moved fail loudly rather than reporting
			// a name that happens to match.
			const aliasName, originalName = "renamed", "exported"
			if resolved.Name != originalName {
				t.Errorf("Checker_resolveAlias reached %q; it must reach the original %q rather "+
					"than stopping at the alias %q", resolved.Name, originalName, aliasName)
			}
			if got := symbolIsDeprecated(resolved); got != testCase.wantResolvedFindsDeprecation {
				t.Errorf("reading the resolved symbol alone says deprecated=%v, wanted %v; that is "+
					"the answer a rule gives WITHOUT the alias walk", got,
					testCase.wantResolvedFindsDeprecation)
			}

			if got := aliasChainFindsDeprecation(t, fileChecker, importedUse); got !=
				testCase.wantChainFindsDeprecation {
				t.Errorf("walking the alias chain says deprecated=%v, wanted %v",
					got, testCase.wantChainFindsDeprecation)
			}
		})
	}
}

// TestAliasAccessorsRefuseANonAlias pins the precondition both accessors share.
//
// Each panics when handed a symbol without `SymbolFlagsAlias`, so every caller needs the gate. The
// panic is provoked here deliberately and recovered, because the alternative is a comment asserting
// a precondition nothing checks, and a precondition nobody has run is how the first version of this
// file came to claim the opposite.
//
// Recovering in a package 49 rules import is a real cost, so it buys something specific: it names
// which of the two behaves this way. Both do, and a future upstream that softened either one would
// fail here rather than silently making a caller's gate redundant.
func TestAliasAccessorsRefuseANonAlias(t *testing.T) {
	t.Parallel()
	fileChecker, sourceFile, release := buildAliasSubject(t, "export function exported(): void {}\n")
	defer release()

	// The locally declared function is not an alias of anything.
	local := aliasSubjectLocalDeclaration(t, fileChecker, sourceFile)
	if local.Flags&ast.SymbolFlagsAlias != 0 {
		t.Fatalf("the local declaration carries SymbolFlagsAlias, so it is the wrong subject "+
			"for this test (flags %v)", local.Flags)
	}

	cases := []struct {
		accessor string
		call     func()
	}{
		{
			accessor: "Checker_resolveAlias",
			call:     func() { checker.Checker_resolveAlias(fileChecker, local) },
		},
		{
			accessor: "Checker_getImmediateAliasedSymbol",
			call: func() {
				checker.Checker_getImmediateAliasedSymbol(fileChecker, local)
			},
		},
	}
	for _, testCase := range cases {
		// Not parallel: every case calls into the parent's one checker, which the parent's deferred release()
		// returns when its body ends, before parallel subtests would start.
		t.Run(testCase.accessor, func(t *testing.T) {
			panicked := func() (panicked bool) {
				defer func() {
					if recover() != nil {
						panicked = true
					}
				}()
				testCase.call()
				return false
			}()
			if !panicked {
				t.Errorf("%s accepted a symbol without SymbolFlagsAlias; every caller gates on that "+
					"flag because this panics, and a caller that stopped gating would now be "+
					"relying on behavior this test says has changed", testCase.accessor)
			}
		})
	}
}

// aliasChainFindsDeprecation walks each link of the alias chain, checking every one.
//
// This is the shape the rule will use: a deprecation may sit on any link, so stopping at the first
// or jumping straight to the last both miss cases the corpus records.
func aliasChainFindsDeprecation(
	t *testing.T,
	fileChecker *checker.Checker,
	symbol *ast.Symbol,
) bool {
	t.Helper()
	current := symbol
	// The bound is a guard against a cyclic chain rather than a real limit; a chain this deep in a
	// two-file fixture would itself be the finding.
	for hop := 0; hop < 8; hop++ {
		if symbolIsDeprecated(current) {
			return true
		}
		if current.Flags&ast.SymbolFlagsAlias == 0 {
			return false
		}
		next := checker.Checker_getImmediateAliasedSymbol(fileChecker, current)
		if next == nil || next == current {
			return false
		}
		current = next
	}
	t.Fatal("the alias chain did not terminate within eight hops, which a two-file fixture cannot " +
		"legitimately produce")
	return false
}

// symbolIsDeprecated reads the `@deprecated` tag off a symbol's own declarations.
func symbolIsDeprecated(symbol *ast.Symbol) bool {
	for _, declaration := range symbol.Declarations {
		if ast.IsDeprecatedDeclaration(declaration) {
			return true
		}
	}
	return false
}

// buildAliasSubject builds a two-file program: a module that exports, and one that imports it.
//
// Two files are required rather than convenient. A same-file `export { original }` produces no alias
// at all, because the use site resolves straight to the declaration, so a single-file fixture reports
// `SymbolFlagsAlias` unset and every assertion here would pass vacuously. That is not hypothetical:
// it was the first shape tried while developing this, and it read as a broken shim.
func buildAliasSubject(t *testing.T, otherSource string) (*checker.Checker, *ast.SourceFile, func()) {
	t.Helper()
	directory := t.TempDir()

	// The import is RENAMED so that "the chain reached the original" is falsifiable. With both
	// names spelled the same, an assertion that the resolved symbol is named `original` passes for
	// a chain that never moved at all, and a mutation removing the check survives. Measured: it
	// did survive, until the rename made the two names differ.
	const subjectSource = "import { exported as renamed } from './Other';\n" +
		"export function localOnly(): void {}\n" +
		"renamed();\n"

	write := func(name string, contents string) {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(contents), 0o644); err != nil {
			t.Fatalf("writing %s: %v", name, err)
		}
	}
	write("Subject.ts", subjectSource)
	write("Other.ts", otherSource)
	write("tsconfig.json",
		`{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"],"types":[]},`+
			`"include":["*.ts"]}`)

	graph, err := program.Build(program.Options{
		ConfigFileName:   filepath.Join(directory, "tsconfig.json"),
		CurrentDirectory: directory,
		SingleThreaded:   true,
	})
	if err != nil {
		t.Fatalf("building the type graph: %v", err)
	}

	var subject *ast.SourceFile
	for _, file := range graph.ProjectFiles() {
		if filepath.Base(file.FileName().AsString()) == "Subject.ts" {
			subject = file
		}
	}
	if subject == nil {
		t.Fatal("the program does not contain Subject.ts, so nothing below is measured")
	}
	fileChecker, release := graph.CheckerForFile(context.Background(), subject)
	return fileChecker, subject, release
}

// aliasSubjectImportedUse returns the symbol at the `original()` call site.
//
// The USE site rather than the import specifier, because that is what a rule visiting identifiers
// is handed and the two need not resolve identically.
func aliasSubjectImportedUse(
	t *testing.T,
	fileChecker *checker.Checker,
	sourceFile *ast.SourceFile,
) *ast.Symbol {
	t.Helper()
	var found *ast.Symbol
	var walk func(node *ast.Node) bool
	walk = func(node *ast.Node) bool {
		if node.Kind == ast.KindCallExpression {
			callee := node.AsCallExpression().Expression
			if callee != nil && callee.Kind == ast.KindIdentifier &&
				callee.Text() == "renamed" {
				found = fileChecker.GetSymbolAtLocation(callee)
			}
		}
		node.ForEachChild(walk)
		return false
	}
	sourceFile.AsNode().ForEachChild(walk)

	if found == nil {
		t.Fatal("no symbol at the `renamed()` call site, so every assertion would pass vacuously")
	}
	return found
}

// aliasSubjectLocalDeclaration returns the symbol of the locally declared function.
func aliasSubjectLocalDeclaration(
	t *testing.T,
	fileChecker *checker.Checker,
	sourceFile *ast.SourceFile,
) *ast.Symbol {
	t.Helper()
	var found *ast.Symbol
	var walk func(node *ast.Node) bool
	walk = func(node *ast.Node) bool {
		if node.Kind == ast.KindFunctionDeclaration {
			if name := node.Name(); name != nil && name.Text() == "localOnly" {
				found = fileChecker.GetSymbolAtLocation(name)
			}
		}
		node.ForEachChild(walk)
		return false
	}
	sourceFile.AsNode().ForEachChild(walk)

	if found == nil {
		t.Fatal("no symbol for the local declaration, so this test measures nothing")
	}
	return found
}
