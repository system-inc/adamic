package nexus

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// The suffixes each declaration form may end in.
//
// Reading mid-file, the suffix tells you the role at a glance: OrderType is a type alias,
// OrderProperties is a React props interface, OperatorKind is an enum-shaped const. The reader
// learns what a name is without scrolling to where it was declared, which is the whole return on
// spending four characters.
var (
	typeAliasSuffixes = []string{"Type", "Properties", "Interface", "Options"}
	interfaceSuffixes = []string{"Interface", "Properties", "Options"}
)

// The rule's messages, one handle per id, whose wording lives in
// `policy/messages/consistency-require-type-suffix.json`.
var (
	consistencyRequireTypeSuffixNoTypeAliasSuffixText = policy.MessageOf("nexus/consistency-require-type-suffix", "noTypeAliasSuffix")
	consistencyRequireTypeSuffixNoInterfaceSuffixText = policy.MessageOf("nexus/consistency-require-type-suffix", "noInterfaceSuffix")
	consistencyRequireTypeSuffixNoConstEnumSuffixText = policy.MessageOf("nexus/consistency-require-type-suffix", "noConstEnumSuffix")
)

func messageNoTypeAliasSuffix(name string) rule.Message {
	return rule.Message{
		Id:          "noTypeAliasSuffix",
		Description: consistencyRequireTypeSuffixNoTypeAliasSuffixText.Render(map[string]string{"name": name}),
	}
}

func messageNoInterfaceSuffix(name string) rule.Message {
	return rule.Message{
		Id:          "noInterfaceSuffix",
		Description: consistencyRequireTypeSuffixNoInterfaceSuffixText.Render(map[string]string{"name": name}),
	}
}

func messageNoConstEnumSuffix(name string) rule.Message {
	return rule.Message{
		Id:          "noConstEnumSuffix",
		Description: consistencyRequireTypeSuffixNoConstEnumSuffixText.Render(map[string]string{"name": name}),
	}
}

// ConsistencyRequireTypeSuffix enforces the role-suffix convention on declarations.
//
//	valid:   type OrderType = { id: string }
//	valid:   interface OrderProperties { id: string }
//	valid:   const OperatorKind = { Add: 'Add' } as const
//	invalid: type Order = { id: string }
//	invalid: interface Order { id: string }
//	invalid: const Operator = { Add: 'Add' } as const
//
// No fix. The new name is a judgment about what the value represents, and the rule offers four
// suffixes precisely because it cannot tell which one is right.
var ConsistencyRequireTypeSuffix = rule.Rule{
	Name: "nexus/consistency-require-type-suffix",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindTypeAliasDeclaration: func(node *ast.Node) {
				name := node.AsTypeAliasDeclaration().Name()
				if name == nil || hasAnySuffix(name.Text(), typeAliasSuffixes) {
					return
				}
				ctx.ReportNode(name, messageNoTypeAliasSuffix(name.Text()))
			},

			ast.KindInterfaceDeclaration: func(node *ast.Node) {
				name := node.AsInterfaceDeclaration().Name()
				if name == nil || hasAnySuffix(name.Text(), interfaceSuffixes) {
					return
				}
				ctx.ReportNode(name, messageNoInterfaceSuffix(name.Text()))
			},

			ast.KindVariableDeclaration: func(node *ast.Node) {
				declaration := node.AsVariableDeclaration()
				if declaration == nil || declaration.Name() == nil {
					return
				}
				name := declaration.Name()
				if name.Kind != ast.KindIdentifier {
					return
				}
				if !isConstEnumShape(declaration.Initializer) {
					return
				}
				if strings.HasSuffix(name.Text(), "Kind") {
					return
				}
				ctx.ReportNode(name, messageNoConstEnumSuffix(name.Text()))
			},
		}
	},
}

// hasAnySuffix reports whether a name ends in any of the allowed suffixes.
func hasAnySuffix(name string, suffixes []string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}
	return false
}

// isConstEnumShape detects an object literal wrapped in `as const` whose every property is
// `KeyName: 'KeyName'`, with key and string value matching exactly.
//
// That signature is a discriminated value set where the runtime form is the type form, which is
// what the Kind convention applies to. The key-equals-value check is what makes the detector
// discriminate rather than merely detect: a lookup table (a CSS class map, an id-by-name
// dictionary, a wire-name translation) is structurally `{ key: 'unrelated string' } as const` too,
// but its keys are labels and its values are unrelated data, so it fails the check cleanly.
// A computed key like `[OtherKind.X]: 'something'` fails it too, which is correct: that is a typed
// map over an existing kind rather than a new kind.
func isConstEnumShape(initializer *ast.Node) bool {
	if initializer == nil || initializer.Kind != ast.KindAsExpression {
		return false
	}

	asExpression := initializer.AsAsExpression()
	if asExpression == nil || asExpression.Type == nil {
		return false
	}
	if asExpression.Type.Kind != ast.KindTypeReference {
		return false
	}
	typeName := asExpression.Type.AsTypeReferenceNode().TypeName
	if typeName == nil || typeName.Kind != ast.KindIdentifier || typeName.Text() != "const" {
		return false
	}

	if asExpression.Expression == nil || asExpression.Expression.Kind != ast.KindObjectLiteralExpression {
		return false
	}
	objectLiteral := asExpression.Expression.AsObjectLiteralExpression()
	if objectLiteral == nil || objectLiteral.Properties == nil || len(objectLiteral.Properties.Nodes) == 0 {
		return false
	}

	// Every property has to match. One lookup-table entry is enough to say this is not an enum,
	// which is the conservative direction: a missed enum costs a naming nudge, while a flagged
	// lookup table asks for a rename that would be wrong.
	for _, property := range objectLiteral.Properties.Nodes {
		if property.Kind != ast.KindPropertyAssignment {
			return false
		}
		assignment := property.AsPropertyAssignment()
		if assignment == nil {
			return false
		}

		key := assignment.Name()
		if key == nil || key.Kind != ast.KindIdentifier {
			return false
		}
		value := assignment.Initializer
		if value == nil || value.Kind != ast.KindStringLiteral {
			return false
		}
		if key.Text() != value.Text() {
			return false
		}
	}
	return true
}
