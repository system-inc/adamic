package program

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/suppression"
)

// A directive naming only rules this run did not run is not dead scaffolding.
//
// During the migration most unused directives are this shape: the rule they name is not ported, so
// nothing looked, and the directive withheld nothing for a reason that says nothing about the tree.
// Reporting it beside a genuinely dead one would tell a reader to delete a suppression the gate
// still needs.
func TestADirectiveNamingOnlyUnrunRulesIsNotDead(t *testing.T) {
	t.Parallel()
	ranRule := map[string]bool{"no-invalid-regexp": true}

	unrun := &suppression.Directive{Rules: []string{"structure/exhaustive-deps"}}
	if !namesOnlyUnrunRules(unrun, ranRule) {
		t.Error("a directive naming an unported rule was counted as dead scaffolding")
	}

	// The plugin prefix must not defeat the comparison. Directives are written against the gate's
	// names, which carry the owning plugin; the registry holds the bare name.
	prefixed := &suppression.Directive{Rules: []string{"core/no-invalid-regexp"}}
	if namesOnlyUnrunRules(prefixed, ranRule) {
		t.Error("a prefixed name for a rule that ran was read as naming an unrun rule")
	}
}

// The other direction: anything a running rule could have silenced is dead scaffolding.
//
// Without this, a helper that answered true for everything would pass the test above and move the
// entire count into the not-dead bucket, which is the same silence as never having measured.
func TestADirectiveWhoseRuleRanIsDead(t *testing.T) {
	t.Parallel()
	ranRule := map[string]bool{"no-invalid-regexp": true, "no-self-assign": true}

	ran := &suppression.Directive{Rules: []string{"no-self-assign"}}
	if namesOnlyUnrunRules(ran, ranRule) {
		t.Error("a directive whose rule ran was excused as naming an unrun rule")
	}

	// A mix is not excused either: at least one rule looked and declined to fire, which is the
	// shape worth deleting.
	mixed := &suppression.Directive{Rules: []string{"structure/exhaustive-deps", "no-self-assign"}}
	if namesOnlyUnrunRules(mixed, ranRule) {
		t.Error("a directive naming one run rule and one unrun rule was excused")
	}

	// A blanket names nothing and silences everything in its scope, so something ran and it still
	// withheld nothing. That is dead by the same reasoning.
	blanket := &suppression.Directive{}
	if namesOnlyUnrunRules(blanket, ranRule) {
		t.Error("a blanket directive was excused as naming an unrun rule")
	}
}
