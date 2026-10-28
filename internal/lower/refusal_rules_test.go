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
		"100 unionSlotWriteRefusal",
		"200 knownLoweringGap",
		"300 functionCallRefusal",
		"400 phantomRefusal",
		"500 typePredicateRefusal",
		"600 enumerationIndexedWriteRefusal",
		"700 enumerationDescriptorRefusal",
		"800 nodeLibraryRefusal",
		"900 detachedOwnRefusal",
		"1000 recordDivergenceRefusal",
		"1100 recordLiteralRead",
		"1200 indexSignatureStorageGap",
		"1300 deleteRefusal",
		"1400 syntaxRefusal",
		"1500 jsonSchemaCallRefusal",
		"1600 typedArrayUnsupported",
		"1700 predicateArgumentsRefusal",
		"1800 enumRefusal",
		"1900 castRefusal",
		"2000 recordStorageView",
		"2100 namespaceRefusal",
		"2200 ambientRefusal",
		"2300 operatorRefusal",
		"2400 asyncGeneratorRefusal",
		"2500 generatorRefusal",
		"2600 argumentsObjectRefusal",
		"2700 nodeBufferAccessRefusal",
		"2800 shorthandLibraryMethodGap",
		"2900 methodValueRefusal",
		"3000 classOverrideRefusal",
		"3100 stringWideningRefusal",
		"3200 refuseVoidWidening",
		"3250 refuseVoidRead",
		"3300 refuseOptionalWidening",
		"3400 classViewRefusal",
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
