package lower

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// refuse's per-node rules run in the order they ran inline in its walk before they registered
// (#86z5k6j): the rule a node meets first is the reason a program stops there. A new rule adds its
// line here at its ordinal.
func TestRefusalRulesRunInOrder(t *testing.T) {
	t.Parallel()
	expected := []string{
		// hostBindingGlobalThisRefusal stops a write through globalThis naming a host binding as its own
		// ruling says, before a rule that meets globalThis or delete first would call it something else.
		"50 hostBindingGlobalThisRefusal",
		// globalObjectValueRefusal stops the global object as a value at its consumer's visit, so its words
		// win over the record view's, node:globals.global's and namespaceRefusal's for those uses.
		"60 globalObjectValueRefusal",
		// The absent member stops before generic writes, descriptors, keys or host-slot lowering.
		"70 hostBindingAbsentMemberRefusal",
		"100 unionSlotWriteRefusal",
		// intlCollatorRefusal names each road to a collator head 1 of #zjndrx5 doesn't build at the
		// node that takes it, before a rule that meets the call, the class or the member first would
		// call it something else.
		"150 intlCollatorRefusal",
		// objectAssignCallableSourceRefusal judges an Object.assign call by its sources' types at the call, before
		// knownLoweringGap or functionCallRefusal meet the call and before lowering reaches the intersection value
		// its result would be (#k8aak04).
		"160 objectAssignCallableSourceRefusal",
		"200 knownLoweringGap",
		"300 functionCallRefusal",
		"400 phantomRefusal",
		// phantomArrayPresenceRefusal judges an observation on the brand before lowering reads the value
		// as its base array, so b09's union `in` hears it, not inProperty's union refusal.
		"450 phantomArrayPresenceRefusal",
		"500 typePredicateRefusal",
		"600 enumerationIndexedWriteRefusal",
		"700 enumerationDescriptorRefusal",
		"800 nodeLibraryRefusal",
		// stdoutHandleSetBlockingRefusal names setBlocking read off stdout's handle as a value before a
		// later rule could call that read something else (#5abf5v8).
		"850 stdoutHandleSetBlockingRefusal",
		// stdoutHandleUnheldRefusal stops a computed-key call or a bind naming setBlocking in a program
		// that reads the handle, on every road the origin walk doesn't follow (#0bxc4jh).
		"851 stdoutHandleUnheldRefusal",
		"900 detachedOwnRefusal",
		"1000 recordDivergenceRefusal",
		"1100 recordLiteralRead",
		"1200 indexSignatureStorageGap",
		"1300 deleteRefusal",
		// recordDeclaredDeleteRefusal judges only a record entry's delete, which deleteRefusal let by.
		"1350 recordDeclaredDeleteRefusal",
		"1400 syntaxRefusal",
		"1500 jsonSchemaCallRefusal",
		"1600 typedArrayUnsupported",
		"1700 predicateArgumentsRefusal",
		"1800 enumRefusal",
		"1900 castRefusal",
		"1950 keptCheckDowncastRefusal",
		// recordDeclaredKeyRefusal names, in .a, a view recordStorageView would call a storage gap.
		"1975 recordDeclaredKeyRefusal",
		"2000 recordStorageView",
		"2050 classDeadZoneGap",
		"2100 namespaceRefusal",
		"2200 ambientRefusal",
		"2250 hostBindingRefusal",
		"2300 operatorRefusal",
		"2400 asyncGeneratorRefusal",
		"2600 argumentsObjectRefusal",
		"2700 nodeBufferAccessRefusal",
		"2800 shorthandLibraryMethodGap",
		"2890 boundObservation",
		"2900 methodValueRefusal",
		"3000 classOverrideRefusal",
		"3100 stringWideningRefusal",
		"3200 refuseVoidWidening",
		"3250 refuseVoidRead",
		"3260 refuseVoidContainerRead",
		// refuseAnyRelation records its NotYet and lets the walk go on, so a NotYet here can lose to
		// a later rule anywhere in the module (refusalPass.deferred).
		"3270 refuseAnyRelation",
		// A member the program declares any is refused where it's declared (#tmdhz1d, step 2).
		"3280 refuseAnyMember",
		// Primitive object views are representation gaps, before optional field hiding.
		"3290 primitiveObjectViewGap",
		"3300 refuseOptionalWidening",
		"3330 narrowedNullishReadGap",
		// constraintRelationRefusal judges the type arguments a site binds once every rule has judged
		// the site itself, and before the walk reaches its arguments.
		"3350 constraintRelationRefusal",
		"3400 classViewRefusal",
		// inNarrowedWriteGap stops a write the in test's Record member alone takes (#yqbg9vz), last, so
		// a rule that meets the same node first keeps its words.
		"3450 inNarrowedWriteGap",
	}
	var registered []string
	for _, rule := range refusalRules.rules {
		registered = append(registered, fmt.Sprintf("%d %s", rule.ordinal, rule.name))
	}
	if !slices.Equal(registered, expected) {
		t.Fatalf("refusal rules run in this order:\n%s\nwant:\n%s", strings.Join(registered, "\n"), strings.Join(expected, "\n"))
	}
}

// Two rules can't share an ordinal, whichever registers first.
func TestRefusalRuleOrdinalTakenOnce(t *testing.T) {
	t.Parallel()
	var registry refusalRegistry
	none := func(*lowering, *ast.Node) error { return nil }
	registry.register(200, "second", none)
	registry.register(100, "first", none)
	defer func() {
		if recover() == nil {
			t.Fatal("a second rule at ordinal 100 registered")
		}
		if registry.rules[0].name != "first" || registry.rules[1].name != "second" {
			t.Fatalf("rules aren't in ordinal order: %s, %s", registry.rules[0].name, registry.rules[1].name)
		}
	}()
	registry.register(100, "again", none)
}
