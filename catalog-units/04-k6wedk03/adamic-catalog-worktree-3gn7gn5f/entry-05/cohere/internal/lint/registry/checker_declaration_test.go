package registry

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
)

// A rule that reads the checker must say so, and this is what checks that it did.
//
// `f4793b6` made the exclusive checker lock conditional on `Rule.NeedsTypeChecker`, which recovered
// about 45% of the lint phase: the lock is held across a file's whole walk rather than around one
// query, so acquiring it serialized every file against every other. Nothing paid for that at the
// time, because zero of the 93 rules then registered read the checker.
//
// **That is no longer the shape of the catalog. 2026-08-25: 44 of 216 registered rules declare it**,
// so the lock is acquired on any file one of those 44 applies to. The saving is still real and it is
// conditional now rather than total. Dated because a rule count in a comment decays silently: this
// sentence said "zero" for as long as it took someone to notice, and nothing failed in between.
//
// **The saving is bought with a claim each rule makes about itself.** A rule that declares false and
// then reads `ctx.TypeChecker` receives nil, and nil does not announce itself: the rule either
// panics into a lost file with a coverage note, or takes a nil-guarded path and reports a wrong
// answer quietly. The second is worse and is the likelier of the two, because most checker helpers
// guard nil internally.
//
// So the declaration is only as good as something checking it, which is the same argument as the
// fixture-pair guard. This reads each rule's source for a `ctx.TypeChecker` reference and requires
// the declaration to match.
func TestRulesDeclareTheirCheckerUse(t *testing.T) {
	t.Parallel()
	declarations := checkerDeclarationsOnDisk(t)
	if len(declarations) == 0 {
		// A sweep with nothing to check passes for the wrong reason, which is the defect it guards.
		t.Fatal("found no rule declarations, so this test proved nothing")
	}

	registered := make(map[string]bool, len(All()))
	for _, registration := range ruleRegistrationsByName() {
		registered[registration.name] = registration.needsChecker
	}

	// Which packages read the checker anywhere, so a rule reaching it through a sibling file's
	// helper is not reported as over-declaring.
	packageReadsChecker := map[string]bool{}
	for _, declaration := range declarations {
		if declaration.readsChecker {
			packageReadsChecker[filepath.Dir(declaration.file)] = true
		}
	}

	for _, declaration := range declarations {
		needs, known := registered[declaration.ruleName]
		if !known {
			// Registration is a different guard's job, and failing here too would report one defect
			// twice under two names.
			continue
		}

		switch {
		case declaration.readsChecker && !needs:
			t.Errorf("rule %q reads ctx.TypeChecker in %s and declares NeedsTypeChecker false, so it "+
				"will be handed nil; a nil checker does not announce itself, it either panics into a "+
				"lost file or takes a nil-guarded path and answers wrongly in silence",
				declaration.ruleName, declaration.file)
		case !declaration.readsChecker && needs && !packageReadsChecker[filepath.Dir(declaration.file)]:
			// The safe direction, and still worth reporting. A rule that declares a need it does not
			// have makes every file it applies to acquire an exclusive lock for nothing, which is
			// exactly the cost `f4793b6` removed.
			//
			// The package-level escape exists because this scan is per file and a rule may reach the
			// checker through a helper its own file does not spell. `no-eval` declares the need and
			// asks `resolvesToAGlobal`, which lives beside it in `no_new_native_nonconstructor.go`,
			// so the per-file answer was a false positive that two porters reported independently
			// before it was fixed. Reporting a correct declaration as an over-declaration is worse
			// than missing one: it teaches the next porter that the guard is noise.
			t.Errorf("rule %q declares NeedsTypeChecker and no file in %s reads ctx.TypeChecker, so "+
				"every file it applies to takes the exclusive lock for nothing",
				declaration.ruleName, filepath.Dir(declaration.file))
		}
	}
}

// ruleRegistration is one registered rule's name and what it declared.
type ruleRegistration struct {
	name         string
	needsChecker bool
}

// ruleRegistrationsByName reads the live catalog rather than any file's text.
func ruleRegistrationsByName() []ruleRegistration {
	registrations := make([]ruleRegistration, 0)
	for _, registration := range All() {
		registrations = append(registrations, ruleRegistration{
			name:         registration.Name,
			needsChecker: registration.NeedsTypeChecker,
		})
	}
	return registrations
}

// checkerDeclaration is one rule file and whether its source touches the checker.
type checkerDeclaration struct {
	file         string
	ruleName     string
	readsChecker bool
}

// checkerDeclarationsOnDisk parses every rule file for its declared name and its checker use.
//
// Parsing rather than grepping, because `ctx.TypeChecker` appears in comments discussing the checker
// in several files here, and a text match would report those as reads.
func checkerDeclarationsOnDisk(t *testing.T) []checkerDeclaration {
	t.Helper()

	paths, err := filepath.Glob("../rules/*/*.go")
	if err != nil {
		t.Fatalf("globbing rule files: %v", err)
	}

	var found []checkerDeclaration
	for _, path := range paths {
		base := filepath.Base(path)
		if strings.HasSuffix(base, "_test.go") || base == "register.go" {
			continue
		}

		parsed, parseErr := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if parseErr != nil {
			t.Fatalf("parsing %s: %v", path, parseErr)
		}

		name, named := ruleNameDeclaredIn(parsed)
		if !named {
			continue
		}

		found = append(found, checkerDeclaration{
			file:         filepath.Join(filepath.Base(filepath.Dir(path)), base),
			ruleName:     name,
			readsChecker: fileReadsTypeChecker(parsed),
		})
	}
	return found
}

// ruleNameDeclaredIn returns the Name of the first `rule.Rule{...}` literal in a file.
func ruleNameDeclaredIn(parsed *ast.File) (string, bool) {
	name := ""
	ast.Inspect(parsed, func(node ast.Node) bool {
		if name != "" {
			return false
		}
		composite, isComposite := node.(*ast.CompositeLit)
		if !isComposite {
			return true
		}
		selector, isSelector := composite.Type.(*ast.SelectorExpr)
		if !isSelector {
			return true
		}
		identifier, isIdentifier := selector.X.(*ast.Ident)
		if !isIdentifier || identifier.Name != "rule" || selector.Sel.Name != "Rule" {
			return true
		}
		for _, element := range composite.Elts {
			keyed, isKeyed := element.(*ast.KeyValueExpr)
			if !isKeyed {
				continue
			}
			key, isKey := keyed.Key.(*ast.Ident)
			if !isKey || key.Name != "Name" {
				continue
			}
			if literal, isLiteral := keyed.Value.(*ast.BasicLit); isLiteral && literal.Kind == token.STRING {
				name = strings.Trim(literal.Value, `"`)
			}
		}
		return true
	})
	return name, name != ""
}

// fileReadsTypeChecker reports whether a file selects `.TypeChecker` off anything.
//
// The receiver is not checked, because a rule reaches the checker through its context by whatever
// name the closure parameter has, and several files here spell it `ctx` while others spell it
// `context`. Selecting `.TypeChecker` off anything in a rule file is the read this cares about.
func fileReadsTypeChecker(parsed *ast.File) bool {
	found := false
	ast.Inspect(parsed, func(node ast.Node) bool {
		if found {
			return false
		}
		if selector, isSelector := node.(*ast.SelectorExpr); isSelector && selector.Sel.Name == "TypeChecker" {
			found = true
			return false
		}
		return true
	})
	return found
}

// TestCheckerDeclarationShareStillSupportsTheConclusion makes the claim in the comments falsifiable.
//
// Two doc comments quote how many rules declare `NeedsTypeChecker`, because the number decides
// whether the conditional checker lock is a blanket saving or a conditional one. Both were wrong for
// an unknown stretch, in the direction that tells a reader the lock costs nothing:
// `rule.go` said one of 112 and `checker_declaration_test.go` said zero of 93, while the real figure
// had reached 44 of 216.
//
// Nothing failed in between, which is the whole problem. A dated comment is honest about being a
// measurement and still says nothing when it expires, so this asserts the claim instead.
//
// # Why this is a band and not the bare equality it used to be
//
// It was written as `wantRegistered == 216 && wantNeedsChecker == 44`, deliberately, so that adding
// any rule forced a reader to revisit the conclusion. That was right when a rule landed every few
// days. It stops being right during a porting wave: the total moves on every single commit, so the
// guard fires constantly on the one number it does not actually care about, and the fix each time is
// to retype a literal. A guard that cries on every commit is one a porter learns to satisfy without
// reading, which is the opposite of the review it was built to force.
//
// **The claim being defended is the share, not the total.** The comment in `rule.go` concludes that
// the checker acquisition is "a real saving and nothing like the blanket one this comment used to
// describe". That conclusion survives 44 of 216 and 44 of 300; it would not survive 200 of 300,
// where the lock is skipped so rarely it is no longer worth the branch, nor 2 of 300, where the
// blanket claim it replaced would be true again.
//
// So the band is on the ratio, wide enough that ordinary porting cannot trip it and narrow enough
// that a real shift does. When it fails, the fix is not to widen the band: it is to read both
// comments and decide whether what they conclude still follows.
func TestCheckerDeclarationShareStillSupportsTheConclusion(t *testing.T) {
	t.Parallel()
	// The share at which the conclusion was measured, and the range over which it still holds.
	// Below the floor, the acquisition is skipped on nearly every file and the blanket claim these
	// comments replaced would be honest again. Above the ceiling, it is skipped so seldom that the
	// conditional is not buying what the comments say it buys.
	const (
		measuredShare = 44.0 / 216.0
		floorShare    = 0.05
		ceilingShare  = 0.50
	)

	rules := All()
	needsChecker := 0
	for _, rule := range rules {
		if rule.NeedsTypeChecker {
			needsChecker++
		}
	}
	if len(rules) == 0 {
		t.Fatal("no rules registered, so this measures nothing")
	}

	share := float64(needsChecker) / float64(len(rules))
	if share < floorShare || share > ceilingShare {
		t.Errorf("%d of %d registered rules declare NeedsTypeChecker (%.0f%%), outside the %.0f%%..%.0f%% "+
			"band the comments here and in `internal/rule/rule.go` draw their conclusion from; that "+
			"conclusion was measured at %.0f%%. Read both comments and decide whether what they say "+
			"about the checker lock still follows, then move the band deliberately rather than to "+
			"make this pass",
			needsChecker, len(rules), 100*share, 100*floorShare, 100*ceilingShare, 100*measuredShare)
	}
}
