package graphql

// src/language-graphql/visitor-keys.evaluate.js and get-visitor-keys.js.
//
// Upstream starts from graphql-js's QueryDocumentKeys (node_modules/graphql/language/ast.js, 17.0.2),
// removes the five schema coordinate kinds, and runs generateReferenceSharedVisitorKeys, which only
// makes identical key arrays share one object. A plain map by kind is the same lookup in Go.
// createGetVisitorKeys(visitorKeys, "kind") reads the node's kind, which is an estree.Node's Type.

import (
	"fmt"

	"github.com/system-inc/cohere/internal/format/estree"
)

// visitorKeys is graphql-js's QueryDocumentKeys minus ArgumentCoordinate, DirectiveArgumentCoordinate,
// DirectiveCoordinate, MemberCoordinate and TypeCoordinate.
//
// Unable to produce https://github.com/prettier/prettier/issues/18212#issuecomment-3506234429
var visitorKeys = map[string][]string{
	"Name":                      {},
	"Document":                  {"definitions"},
	"OperationDefinition":       {"description", "name", "variableDefinitions", "directives", "selectionSet"},
	"VariableDefinition":        {"description", "variable", "type", "defaultValue", "directives"},
	"Variable":                  {"name"},
	"SelectionSet":              {"selections"},
	"Field":                     {"alias", "name", "arguments", "directives", "selectionSet"},
	"Argument":                  {"name", "value"},
	"FragmentArgument":          {"name", "value"},
	"FragmentSpread":            {"name", "arguments", "directives"},
	"InlineFragment":            {"typeCondition", "directives", "selectionSet"},
	"FragmentDefinition":        {"description", "name", "variableDefinitions", "typeCondition", "directives", "selectionSet"},
	"IntValue":                  {},
	"FloatValue":                {},
	"StringValue":               {},
	"BooleanValue":              {},
	"NullValue":                 {},
	"EnumValue":                 {},
	"ListValue":                 {"values"},
	"ObjectValue":               {"fields"},
	"ObjectField":               {"name", "value"},
	"Directive":                 {"name", "arguments"},
	"NamedType":                 {"name"},
	"ListType":                  {"type"},
	"NonNullType":               {"type"},
	"SchemaDefinition":          {"description", "directives", "operationTypes"},
	"OperationTypeDefinition":   {"type"},
	"ScalarTypeDefinition":      {"description", "name", "directives"},
	"ObjectTypeDefinition":      {"description", "name", "interfaces", "directives", "fields"},
	"FieldDefinition":           {"description", "name", "arguments", "type", "directives"},
	"InputValueDefinition":      {"description", "name", "type", "defaultValue", "directives"},
	"InterfaceTypeDefinition":   {"description", "name", "interfaces", "directives", "fields"},
	"UnionTypeDefinition":       {"description", "name", "directives", "types"},
	"EnumTypeDefinition":        {"description", "name", "directives", "values"},
	"EnumValueDefinition":       {"description", "name", "directives"},
	"InputObjectTypeDefinition": {"description", "name", "directives", "fields"},
	"DirectiveDefinition":       {"description", "name", "arguments", "directives", "locations"},
	"SchemaExtension":           {"directives", "operationTypes"},
	"DirectiveExtension":        {"name", "directives"},
	"ScalarTypeExtension":       {"name", "directives"},
	"ObjectTypeExtension":       {"name", "interfaces", "directives", "fields"},
	"InterfaceTypeExtension":    {"name", "interfaces", "directives", "fields"},
	"UnionTypeExtension":        {"name", "directives", "types"},
	"EnumTypeExtension":         {"name", "directives", "values"},
	"InputObjectTypeExtension":  {"name", "directives", "fields"},
}

// getVisitorKeys is upstream's getVisitorKeys(node): the properties holding child nodes, in order. A
// kind it does not know is an error, as createGetVisitorKeys throws "Missing visitor keys"; the core
// only asks about nodes in the tree, never about a comment.
func getVisitorKeys(node *estree.Node) []string {
	keys, known := visitorKeys[node.Type()]
	if !known {
		panic(fmt.Sprintf("Missing visitor keys for '%s'.", node.Type()))
	}
	return keys
}
