package high_level_intermediate_representation

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
	"github.com/system-inc/cohere/static_single_assignment"
)

// TestIdentifiersShareNoStorageAcrossFunctions holds what makes carving identifiers from chunks safe.
//
// Two functions are lowered in one file, each with enough values to need several chunks. Within each,
// every identifier is still the one its id names after later chunks were started, so no pointer moved.
// Across them, no identifier is shared, and writing through every identifier of the first leaves the
// second exactly as it was.
func TestIdentifiersShareNoStorageAcrossFunctions(t *testing.T) {
	t.Parallel()
	var body strings.Builder
	for index := 0; index < 60; index++ {
		body.WriteString("  const value" + strings.Repeat("x", index%5) + string(rune('a'+index%26)) + " = input + 1;\n")
	}
	source := "function first(input: number) {\n" + body.String() + "  return input;\n}\n" +
		"function second(input: number) {\n" + body.String() + "  return input;\n}\n"

	var functions []*Function
	probe := rule.Rule{Name: "identifier-slab", NeedsTypeChecker: true, Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{ast.KindSourceFile: func(node *ast.Node) {
			forEachFunctionLike(node, func(node *ast.Node) {
				if function := Lower(node, ctx.TypeChecker); function != nil {
					Construct(function)
					functions = append(functions, function)
				}
			})
		}}
	}}
	rule_testing.RunTypedFiles(t, probe, map[string]string{"/fixture.ts": source}, "/fixture.ts")
	if len(functions) != 2 {
		t.Fatalf("lowered %d functions, want two", len(functions))
	}

	for _, function := range functions {
		if len(function.Identifiers) <= 16 {
			t.Fatalf("%s has %d identifiers, too few to span a second chunk", function.Name, len(function.Identifiers))
		}
		for index, identifier := range function.Identifiers {
			if identifier.Id != static_single_assignment.IdentifierId(index) {
				t.Fatalf("%s: identifier %d reads as %d, so a pointer into an earlier chunk moved",
					function.Name, index, identifier.Id)
			}
		}
	}

	first, second := functions[0], functions[1]
	owned := map[*Identifier]bool{}
	for _, identifier := range first.Identifiers {
		owned[identifier] = true
	}
	before := make([]Identifier, len(second.Identifiers))
	for index, identifier := range second.Identifiers {
		if owned[identifier] {
			t.Fatalf("identifier %d of %s is also one of %s's", index, second.Name, first.Name)
		}
		before[index] = *identifier
	}
	for _, identifier := range first.Identifiers {
		identifier.Name = "written"
		identifier.Declaration = 0
	}
	for index, identifier := range second.Identifiers {
		if *identifier != before[index] {
			t.Fatalf("writing %s's identifiers changed %s's identifier %d", first.Name, second.Name, index)
		}
	}
}
