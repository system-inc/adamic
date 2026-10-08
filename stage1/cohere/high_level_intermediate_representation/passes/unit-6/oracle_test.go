//go:build lintoracle

// Overlay beside Go HIR with the owner's printer and checkpoint framing.
// This fixture proves the shared Scope-terminal contract needed by unit 6.
package high_level_intermediate_representation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/cohere/mutation_aliasing"
	"github.com/system-inc/cohere/static_single_assignment"
)

func unit6ScopeRows(scopes *ReactiveScopes, identity ScopeIdentity) ([]OracleExtraIdentity, []OracleSidecarRow) {
	var identities []OracleExtraIdentity
	var rows []OracleSidecarRow
	for _, scope := range scopes.Ids() {
		id := int(scope)
		identities = append(identities, OracleExtraIdentity{FunctionPath: "$", Kind: "scope", Id: id})
		bounds := identity.RangeOf(identity.GroupOf(scope))
		rows = append(rows, OracleSidecarRow{Namespace: "unit6.scopes", FunctionPath: "$", AnchorKind: "scope", AnchorId: &id, Key: "range", Payload: fmt.Sprintf("%d\t%d", bounds.Start, bounds.End)})
		group := identity.GroupOf(scope)
		rows = append(rows, OracleSidecarRow{Namespace: "unit6.scopes", FunctionPath: "$", AnchorKind: "scope", AnchorId: &id, Key: "group", Payload: fmt.Sprint(group)})
		for _, member := range scopes.MembersOf(scope) {
			identifier := int(member)
			rows = append(rows, OracleSidecarRow{Namespace: "unit6.scopes", FunctionPath: "$", AnchorKind: "identifier", AnchorId: &identifier, Key: "scope", Payload: fmt.Sprint(scope)})
		}
	}
	return identities, rows
}

func unit6TerminalCheckpoint(function *Function, scopes *ReactiveScopes, identity ScopeIdentity, outcome ScopeTerminals) string {
	identities, rows := unit6ScopeRows(scopes, identity)
	rows = append(rows, OracleSidecarRow{Namespace: "unit6.terminals", FunctionPath: "$", AnchorKind: "function", Key: "outcome", Payload: fmt.Sprintf("%d\t%d\t%d", outcome.Built, outcome.BlocksSplit, outcome.PhiOperandsRekeyed)})
	return OracleWriteCheckpoint(OracleCheckpoint{Key: "unit6-single-scope", Pass: "unit6.scope-terminals", Graph: oracleDump(function), Identities: identities, Sidecars: append(OracleInputFacts(function, nil), rows...)})
}

func TestUnit6ScopeTerminalFixture(t *testing.T) {
	t.Parallel()
	// The shortest constructed graph with a scoped allocation and a return.
	// The inputs are Go-produced ranges/sets, not an Adamic analysis answer.
	function := NewFunction(nil, "f", FunctionKindOther)
	block := function.NewBlock(BlockKindBlock)
	function.Entry = block.Id
	place := Place{Identifier: function.NewIdentifier("", nil, 0).Id}
	function.Returns = place
	function.AddInstruction(block, &Instruction{LValue: place, Value: &ArrayExpression{}, Order: 1})
	block.Terminal = &Return{Value: place, Order: 2}
	ranges := &mutation_aliasing.MutableRanges{}
	ranges.Set(place.Identifier, mutation_aliasing.MutableRange{Start: 1, End: 2})
	set := &DisjointSet{}
	set.Union([]static_single_assignment.IdentifierId{place.Identifier})
	scopes := AssignReactiveScopesWithSets(function, ranges, set)
	scopes = AlignMethodCallScopes(function, scopes)
	aligned, merged := AlignThenMergeReactiveScopes(function, scopes)
	identity := MergedScopeIdentity{Aligned: aligned, Merged: merged}
	if invalid := ValidateScopes(function, scopes); len(invalid) != 0 {
		t.Fatalf("unexpected invalid scopes: %v", invalid)
	}
	if precondition := ScopeTerminalsPrecondition(function, scopes, identity); precondition != 0 {
		t.Fatalf("unexpected nesting failure: %d", precondition)
	}
	before := unit6TerminalCheckpoint(function, scopes, identity, ScopeTerminals{})
	outcome := BuildReactiveScopeTerminals(function, scopes, identity)
	after := unit6TerminalCheckpoint(function, scopes, identity, outcome)
	if outcome.Built != 1 || outcome.BlocksSplit != 1 || outcome.PhiOperandsRekeyed != 0 {
		t.Fatalf("scope rewrite did not fire: %+v", outcome)
	}
	if !strings.Contains(after, " Scope ") || strings.Contains(before, " Scope ") {
		t.Fatal("fixture does not distinguish the terminal rewrite")
	}
	if after != unit6TerminalCheckpoint(function, scopes, identity, outcome) {
		t.Fatal("scope checkpoint is nondeterministic")
	}
	destination := os.Getenv("HIR_UNIT6_FIXTURES")
	if destination == "" {
		t.Fatal("HIR_UNIT6_FIXTURES is required")
	}
	for name, data := range map[string]string{"single-scope.before.checkpoint": before, "single-scope.after.checkpoint": after} {
		if err := os.WriteFile(filepath.Join(destination, name), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Log("Go scope-terminal fixture: built=1 blocks-split=1 phi-operands-rekeyed=0; deterministic before/after framing")
}
