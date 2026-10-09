package graphql

import (
	"github.com/system-inc/cohere/internal/format/estree"
	"github.com/system-inc/cohere/internal/format/printing"
)

// graphql-js 17.0.2, language/parser.js: parse and the Parser class it drives, with parseComments
// from Prettier's src/language-graphql/parser-graphql.js.
//
// # What is not here, and why
//
//   - parseValue, parseConstValue, parseType and parseSchemaCoordinate are graphql-js's other entries.
//     Prettier calls only parse, and parse cannot reach parseSchemaCoordinate (only the
//     parseSchemaCoordinate entry calls it, with its own SchemaCoordinateLexer), so the coordinate
//     nodes (TypeCoordinate, MemberCoordinate, ArgumentCoordinate, DirectiveCoordinate,
//     DirectiveArgumentCoordinate) never appear in a document and are not ported.
//   - Options are Prettier's: experimentalFragmentArguments is true, and the rest keep graphql-js's
//     defaults. noLocation is false, so every node has its location (here, the node's Range). maxTokens
//     is unset, so the token counter it guards would never fire and is not kept; neither is the
//     document's tokenCount, which nothing reads. The diagnostics channel tracing parse is Node's.
//   - A node's kind is the estree.Node's type, and loc is its Range, in bytes (graphql-js's offsets are
//     UTF-16). Its other fields are properties in graphql-js's creation order. A field graphql-js
//     writes as undefined is a property holding nil, and a field it never writes (a FragmentSpread's
//     arguments, when no arguments follow) is absent, so Has answers as `in` would.
//
// graphql-js throws a GraphQLError where parsing fails. Here the parser panics with the *SyntaxError
// and Parse recovers it into the error it returns; no panic leaves Parse.

// Parse is Prettier's GraphQL parser: graphql-js's parse(text, { experimentalFragmentArguments: true }),
// then parseComments, which collects the Comment tokens from the token stream, since graphql-js's AST
// has no comments. Each comment is a "Comment" node whose value is the text after the "#".
func Parse(text string) (document *estree.Node, comments []*estree.Node, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			parseError, isSyntaxError := recovered.(*SyntaxError)
			if !isSyntaxError {
				panic(recovered)
			}
			// Marked as a parse failure, so the format phase skips the file rather than reporting a broken
			// formatter.
			document, comments, err = nil, nil, printing.Syntax(parseError)
		}
	}()

	parser := newParser(text, parseOptions{experimentalFragmentArguments: true})
	startToken := parser.lexer.token
	document = parser.parseDocument()
	endToken := parser.lexer.lastToken
	return document, parseComments(startToken, endToken), nil
}

// parseComments is parser-graphql.js's: the Comment tokens between the document's startToken and
// endToken (the <SOF> and the <EOF>), which lookahead linked into the token list as it skipped them.
func parseComments(startToken *token, endToken *token) []*estree.Node {
	comments := []*estree.Node{}
	for token := startToken; token != endToken; token = token.next {
		if token.kind == tokenKindComment {
			comments = append(comments, estree.New("Comment", token.start, token.end, "value", token.value))
		}
	}

	return comments
}

// parseOptions is the part of graphql-js's ParseOptions that Prettier sets.
type parseOptions struct {
	// EXPERIMENTAL:
	//
	// If enabled, the parser will understand and parse fragment variable definitions and arguments on
	// fragment spreads. Fragment variable definitions will be represented in the `variableDefinitions`
	// field of the FragmentDefinitionNode. Fragment spread arguments will be represented in the
	// `arguments` field of FragmentSpreadNode.
	//
	//	{
	//	  t { ...A(var: true) }
	//	}
	//	fragment A($var: Boolean = false) on T {
	//	  ...B(x: $var)
	//	}
	experimentalFragmentArguments bool
}

// parser is graphql-js's Parser class.
type parser struct {
	lexer   *lexer
	options parseOptions
}

func newParser(source string, options parseOptions) *parser {
	return &parser{lexer: newLexer(source), options: options}
}

// parseName converts a name lex token into a name parse node.
func (parser *parser) parseName() *estree.Node {
	token := parser.expectToken(tokenKindName)
	return parser.node(token, "Name",
		"value", token.value,
	)
}

// Implements the parsing rules in the Document section.

// parseDocument parses Document : Definition+
func (parser *parser) parseDocument() *estree.Node {
	start := parser.lexer.token
	return parser.node(start, "Document",
		"definitions", parser.many(tokenKindSOF, parser.parseDefinition, tokenKindEOF),
	)
}

// parseDefinition parses
//
//	Definition :
//	  - ExecutableDefinition
//	  - TypeSystemDefinition
//	  - TypeSystemExtension
//
//	ExecutableDefinition :
//	  - OperationDefinition
//	  - FragmentDefinition
//
//	TypeSystemDefinition :
//	  - SchemaDefinition
//	  - TypeDefinition
//	  - DirectiveDefinition
//
//	TypeDefinition :
//	  - ScalarTypeDefinition
//	  - ObjectTypeDefinition
//	  - InterfaceTypeDefinition
//	  - UnionTypeDefinition
//	  - EnumTypeDefinition
//	  - InputObjectTypeDefinition
func (parser *parser) parseDefinition() *estree.Node {
	if parser.peek(tokenKindBraceL) {
		return parser.parseOperationDefinition()
	}

	// Many definitions begin with a description and require a lookahead.
	hasDescription := parser.peekDescription()
	keywordToken := parser.lexer.token
	if hasDescription {
		keywordToken = parser.lexer.lookahead()
	}

	if hasDescription && keywordToken.kind == tokenKindBraceL {
		panic(syntaxError(parser.lexer.body, parser.lexer.token.start,
			"Unexpected description, descriptions are not supported on shorthand queries."))
	}

	if keywordToken.kind == tokenKindName {
		switch keywordToken.value {
		case "schema":
			return parser.parseSchemaDefinition()
		case "scalar":
			return parser.parseScalarTypeDefinition()
		case "type":
			return parser.parseObjectTypeDefinition()
		case "interface":
			return parser.parseInterfaceTypeDefinition()
		case "union":
			return parser.parseUnionTypeDefinition()
		case "enum":
			return parser.parseEnumTypeDefinition()
		case "input":
			return parser.parseInputObjectTypeDefinition()
		case "directive":
			return parser.parseDirectiveDefinition()
		}

		switch keywordToken.value {
		case "query", "mutation", "subscription":
			return parser.parseOperationDefinition()
		case "fragment":
			return parser.parseFragmentDefinition()
		}

		if hasDescription {
			panic(syntaxError(parser.lexer.body, parser.lexer.token.start,
				"Unexpected description, only GraphQL definitions support descriptions."))
		}

		switch keywordToken.value {
		case "extend":
			return parser.parseTypeSystemExtension()
		}
	}

	panic(parser.unexpected(keywordToken))
}

// Implements the parsing rules in the Operations section.

// parseOperationDefinition parses
//
//	OperationDefinition :
//	 - SelectionSet
//	 - OperationType Name? VariableDefinitions? Directives? SelectionSet
func (parser *parser) parseOperationDefinition() *estree.Node {
	start := parser.lexer.token
	if parser.peek(tokenKindBraceL) {
		return parser.node(start, "OperationDefinition",
			"operation", "query",
			"description", nil,
			"name", nil,
			"variableDefinitions", nil,
			"directives", nil,
			"selectionSet", parser.parseSelectionSet(),
		)
	}
	description := parser.parseDescription()
	operation := parser.parseOperationType()
	var name *estree.Node
	if parser.peek(tokenKindName) {
		name = parser.parseName()
	}
	return parser.node(start, "OperationDefinition",
		"operation", operation,
		"description", description,
		"name", name,
		"variableDefinitions", parser.parseVariableDefinitions(),
		"directives", parser.parseDirectives(false),
		"selectionSet", parser.parseSelectionSet(),
	)
}

// parseOperationType parses OperationType : one of query mutation subscription
func (parser *parser) parseOperationType() string {
	operationToken := parser.expectToken(tokenKindName)
	switch operationToken.value {
	case "query":
		return "query"
	case "mutation":
		return "mutation"
	case "subscription":
		return "subscription"
	}

	panic(parser.unexpected(operationToken))
}

// parseVariableDefinitions parses VariableDefinitions : ( VariableDefinition+ )
func (parser *parser) parseVariableDefinitions() []*estree.Node {
	return parser.optionalMany(tokenKindParenL, parser.parseVariableDefinition, tokenKindParenR)
}

// parseVariableDefinition parses VariableDefinition : Variable : Type DefaultValue? Directives[Const]?
func (parser *parser) parseVariableDefinition() *estree.Node {
	start := parser.lexer.token
	description := parser.parseDescription()
	variable := parser.parseVariable()
	parser.expectToken(tokenKindColon)
	typeReference := parser.parseTypeReference()
	var defaultValue *estree.Node
	if parser.expectOptionalToken(tokenKindEquals) {
		defaultValue = parser.parseConstValueLiteral()
	}
	return parser.node(start, "VariableDefinition",
		"description", description,
		"variable", variable,
		"type", typeReference,
		"defaultValue", defaultValue,
		"directives", parser.parseConstDirectives(),
	)
}

// parseVariable parses Variable : $ Name
func (parser *parser) parseVariable() *estree.Node {
	start := parser.lexer.token
	parser.expectToken(tokenKindDollar)
	return parser.node(start, "Variable",
		"name", parser.parseName(),
	)
}

// parseSelectionSet parses
//
//	SelectionSet : { Selection+ }
func (parser *parser) parseSelectionSet() *estree.Node {
	start := parser.lexer.token
	return parser.node(start, "SelectionSet",
		"selections", parser.many(tokenKindBraceL, parser.parseSelection, tokenKindBraceR),
	)
}

// parseSelection parses
//
//	Selection :
//	  - Field
//	  - FragmentSpread
//	  - InlineFragment
func (parser *parser) parseSelection() *estree.Node {
	if parser.peek(tokenKindSpread) {
		return parser.parseFragment()
	}
	return parser.parseField()
}

// parseField parses
//
//	Field : Alias? Name Arguments? Directives? SelectionSet?
//
//	Alias : Name :
func (parser *parser) parseField() *estree.Node {
	start := parser.lexer.token

	nameOrAlias := parser.parseName()
	var alias, name *estree.Node
	if parser.expectOptionalToken(tokenKindColon) {
		alias = nameOrAlias
		name = parser.parseName()
	} else {
		name = nameOrAlias
	}

	arguments := parser.parseArguments(false)
	directives := parser.parseDirectives(false)
	var selectionSet *estree.Node
	if parser.peek(tokenKindBraceL) {
		selectionSet = parser.parseSelectionSet()
	}
	return parser.node(start, "Field",
		"alias", alias,
		"name", name,
		"arguments", arguments,
		"directives", directives,
		"selectionSet", selectionSet,
	)
}

// parseArguments parses Arguments[Const] : ( Argument[?Const]+ )
func (parser *parser) parseArguments(isConst bool) []*estree.Node {
	item := parser.parseArgument
	if isConst {
		item = parser.parseConstArgument
	}
	return parser.optionalMany(tokenKindParenL, item, tokenKindParenR)
}

func (parser *parser) parseFragmentArguments() []*estree.Node {
	item := parser.parseFragmentArgument
	return parser.optionalMany(tokenKindParenL, item, tokenKindParenR)
}

// parseArgument parses Argument[Const] : Name : Value[?Const]. Its isConst defaults to false, which a
// method value passed as a parse function cannot spell, so the parse function is parseArgument and the
// body is parseArgumentConst.
func (parser *parser) parseArgument() *estree.Node {
	return parser.parseArgumentConst(false)
}

func (parser *parser) parseArgumentConst(isConst bool) *estree.Node {
	start := parser.lexer.token
	name := parser.parseName()

	parser.expectToken(tokenKindColon)
	return parser.node(start, "Argument",
		"name", name,
		"value", parser.parseValueLiteral(isConst),
	)
}

func (parser *parser) parseConstArgument() *estree.Node {
	return parser.parseArgumentConst(true)
}

func (parser *parser) parseFragmentArgument() *estree.Node {
	start := parser.lexer.token
	name := parser.parseName()

	parser.expectToken(tokenKindColon)
	return parser.node(start, "FragmentArgument",
		"name", name,
		"value", parser.parseValueLiteral(false),
	)
}

// Implements the parsing rules in the Fragments section.

// parseFragment corresponds to both FragmentSpread and InlineFragment in the spec.
//
//	FragmentSpread : ... FragmentName Arguments? Directives?
//
//	InlineFragment : ... TypeCondition? Directives? SelectionSet
func (parser *parser) parseFragment() *estree.Node {
	start := parser.lexer.token
	parser.expectToken(tokenKindSpread)

	hasTypeCondition := parser.expectOptionalKeyword("on")
	if !hasTypeCondition && parser.peek(tokenKindName) {
		name := parser.parseFragmentName()
		if parser.peek(tokenKindParenL) && parser.options.experimentalFragmentArguments {
			arguments := parser.parseFragmentArguments()
			return parser.node(start, "FragmentSpread",
				"name", name,
				"arguments", arguments,
				"directives", parser.parseDirectives(false),
			)
		}
		return parser.node(start, "FragmentSpread",
			"name", name,
			"directives", parser.parseDirectives(false),
		)
	}
	var typeCondition *estree.Node
	if hasTypeCondition {
		typeCondition = parser.parseNamedType()
	}
	directives := parser.parseDirectives(false)
	return parser.node(start, "InlineFragment",
		"typeCondition", typeCondition,
		"directives", directives,
		"selectionSet", parser.parseSelectionSet(),
	)
}

// parseFragmentDefinition parses
//
//	FragmentDefinition :
//	  - fragment FragmentName VariableDefinitions? on TypeCondition Directives? SelectionSet
//
//	TypeCondition : NamedType
func (parser *parser) parseFragmentDefinition() *estree.Node {
	start := parser.lexer.token
	description := parser.parseDescription()
	parser.expectKeyword("fragment")
	if parser.options.experimentalFragmentArguments {
		name := parser.parseFragmentName()
		variableDefinitions := parser.parseVariableDefinitions()
		parser.expectKeyword("on")
		typeCondition := parser.parseNamedType()
		directives := parser.parseDirectives(false)
		return parser.node(start, "FragmentDefinition",
			"description", description,
			"name", name,
			"variableDefinitions", variableDefinitions,
			"typeCondition", typeCondition,
			"directives", directives,
			"selectionSet", parser.parseSelectionSet(),
		)
	}
	name := parser.parseFragmentName()
	parser.expectKeyword("on")
	typeCondition := parser.parseNamedType()
	directives := parser.parseDirectives(false)
	return parser.node(start, "FragmentDefinition",
		"description", description,
		"name", name,
		"typeCondition", typeCondition,
		"directives", directives,
		"selectionSet", parser.parseSelectionSet(),
	)
}

// parseFragmentName parses FragmentName : Name but not `on`
func (parser *parser) parseFragmentName() *estree.Node {
	if parser.lexer.token.value == "on" {
		panic(parser.unexpected(nil))
	}
	return parser.parseName()
}

// Implements the parsing rules in the Values section.

// parseValueLiteral parses
//
//	Value[Const] :
//	  - [~Const] Variable
//	  - IntValue
//	  - FloatValue
//	  - StringValue
//	  - BooleanValue
//	  - NullValue
//	  - EnumValue
//	  - ListValue[?Const]
//	  - ObjectValue[?Const]
//
//	BooleanValue : one of `true` `false`
//
//	NullValue : `null`
//
//	EnumValue : Name but not `true`, `false` or `null`
func (parser *parser) parseValueLiteral(isConst bool) *estree.Node {
	token := parser.lexer.token
	switch token.kind {
	case tokenKindBracketL:
		return parser.parseList(isConst)
	case tokenKindBraceL:
		return parser.parseObject(isConst)
	case tokenKindInt:
		parser.advanceLexer()
		return parser.node(token, "IntValue",
			"value", token.value,
		)
	case tokenKindFloat:
		parser.advanceLexer()
		return parser.node(token, "FloatValue",
			"value", token.value,
		)
	case tokenKindString, tokenKindBlockString:
		return parser.parseStringLiteral()
	case tokenKindName:
		parser.advanceLexer()
		switch token.value {
		case "true":
			return parser.node(token, "BooleanValue",
				"value", true,
			)
		case "false":
			return parser.node(token, "BooleanValue",
				"value", false,
			)
		case "null":
			return parser.node(token, "NullValue")
		default:
			return parser.node(token, "EnumValue",
				"value", token.value,
			)
		}
	case tokenKindDollar:
		if isConst {
			parser.expectToken(tokenKindDollar)
			if parser.lexer.token.kind == tokenKindName {
				variableName := parser.lexer.token.value
				panic(syntaxError(parser.lexer.body, token.start,
					`Unexpected variable "$`+variableName+`" in constant value.`))
			}
			panic(parser.unexpected(token))
		}
		return parser.parseVariable()
	default:
		panic(parser.unexpected(nil))
	}
}

func (parser *parser) parseConstValueLiteral() *estree.Node {
	return parser.parseValueLiteral(true)
}

func (parser *parser) parseStringLiteral() *estree.Node {
	token := parser.lexer.token
	parser.advanceLexer()
	return parser.node(token, "StringValue",
		"value", token.value,
		"block", token.kind == tokenKindBlockString,
	)
}

// parseList parses
//
//	ListValue[Const] :
//	  - [ ]
//	  - [ Value[?Const]+ ]
func (parser *parser) parseList(isConst bool) *estree.Node {
	item := func() *estree.Node { return parser.parseValueLiteral(isConst) }
	start := parser.lexer.token
	return parser.node(start, "ListValue",
		"values", parser.any(tokenKindBracketL, item, tokenKindBracketR),
	)
}

// parseObject parses
//
//	ObjectValue[Const] :
//	  - { }
//	  - { ObjectField[?Const]+ }
func (parser *parser) parseObject(isConst bool) *estree.Node {
	item := func() *estree.Node { return parser.parseObjectField(isConst) }
	start := parser.lexer.token
	return parser.node(start, "ObjectValue",
		"fields", parser.any(tokenKindBraceL, item, tokenKindBraceR),
	)
}

// parseObjectField parses ObjectField[Const] : Name : Value[?Const]
func (parser *parser) parseObjectField(isConst bool) *estree.Node {
	start := parser.lexer.token
	name := parser.parseName()
	parser.expectToken(tokenKindColon)

	return parser.node(start, "ObjectField",
		"name", name,
		"value", parser.parseValueLiteral(isConst),
	)
}

// Implements the parsing rules in the Directives section.

// parseDirectives parses Directives[Const] : Directive[?Const]+
func (parser *parser) parseDirectives(isConst bool) []*estree.Node {
	var directives []*estree.Node
	for parser.peek(tokenKindAt) {
		directives = append(directives, parser.parseDirective(isConst))
	}
	if len(directives) > 0 {
		return directives
	}
	return nil
}

func (parser *parser) parseConstDirectives() []*estree.Node {
	return parser.parseDirectives(true)
}

// parseDirective parses
//
//	Directive[Const] : @ Name Arguments[?Const]?
func (parser *parser) parseDirective(isConst bool) *estree.Node {
	start := parser.lexer.token
	parser.expectToken(tokenKindAt)
	name := parser.parseName()
	return parser.node(start, "Directive",
		"name", name,
		"arguments", parser.parseArguments(isConst),
	)
}

// Implements the parsing rules in the Types section.

// parseTypeReference parses
//
//	Type :
//	  - NamedType
//	  - ListType
//	  - NonNullType
func (parser *parser) parseTypeReference() *estree.Node {
	start := parser.lexer.token
	var typeReference *estree.Node
	if parser.expectOptionalToken(tokenKindBracketL) {
		innerType := parser.parseTypeReference()
		parser.expectToken(tokenKindBracketR)
		typeReference = parser.node(start, "ListType",
			"type", innerType,
		)
	} else {
		typeReference = parser.parseNamedType()
	}

	if parser.expectOptionalToken(tokenKindBang) {
		return parser.node(start, "NonNullType",
			"type", typeReference,
		)
	}

	return typeReference
}

// parseNamedType parses NamedType : Name
func (parser *parser) parseNamedType() *estree.Node {
	start := parser.lexer.token
	return parser.node(start, "NamedType",
		"name", parser.parseName(),
	)
}

// Implements the parsing rules in the Type Definition section.

func (parser *parser) peekDescription() bool {
	return parser.peek(tokenKindString) || parser.peek(tokenKindBlockString)
}

// parseDescription parses Description : StringValue
func (parser *parser) parseDescription() *estree.Node {
	if parser.peekDescription() {
		return parser.parseStringLiteral()
	}
	return nil
}

// parseSchemaDefinition parses
//
//	SchemaDefinition : Description? schema Directives[Const]? { OperationTypeDefinition+ }
func (parser *parser) parseSchemaDefinition() *estree.Node {
	start := parser.lexer.token
	description := parser.parseDescription()
	parser.expectKeyword("schema")
	directives := parser.parseConstDirectives()
	operationTypes := parser.many(tokenKindBraceL, parser.parseOperationTypeDefinition, tokenKindBraceR)
	return parser.node(start, "SchemaDefinition",
		"description", description,
		"directives", directives,
		"operationTypes", operationTypes,
	)
}

// parseOperationTypeDefinition parses OperationTypeDefinition : OperationType : NamedType
func (parser *parser) parseOperationTypeDefinition() *estree.Node {
	start := parser.lexer.token
	operation := parser.parseOperationType()
	parser.expectToken(tokenKindColon)
	typeReference := parser.parseNamedType()
	return parser.node(start, "OperationTypeDefinition",
		"operation", operation,
		"type", typeReference,
	)
}

// parseScalarTypeDefinition parses ScalarTypeDefinition : Description? scalar Name Directives[Const]?
func (parser *parser) parseScalarTypeDefinition() *estree.Node {
	start := parser.lexer.token
	description := parser.parseDescription()
	parser.expectKeyword("scalar")
	name := parser.parseName()
	directives := parser.parseConstDirectives()
	return parser.node(start, "ScalarTypeDefinition",
		"description", description,
		"name", name,
		"directives", directives,
	)
}

// parseObjectTypeDefinition parses
//
//	ObjectTypeDefinition :
//	  Description?
//	  type Name ImplementsInterfaces? Directives[Const]? FieldsDefinition?
func (parser *parser) parseObjectTypeDefinition() *estree.Node {
	start := parser.lexer.token
	description := parser.parseDescription()
	parser.expectKeyword("type")
	name := parser.parseName()
	interfaces := parser.parseImplementsInterfaces()
	directives := parser.parseConstDirectives()
	fields := parser.parseFieldsDefinition()
	return parser.node(start, "ObjectTypeDefinition",
		"description", description,
		"name", name,
		"interfaces", interfaces,
		"directives", directives,
		"fields", fields,
	)
}

// parseImplementsInterfaces parses
//
//	ImplementsInterfaces :
//	  - implements `&`? NamedType
//	  - ImplementsInterfaces & NamedType
func (parser *parser) parseImplementsInterfaces() []*estree.Node {
	if parser.expectOptionalKeyword("implements") {
		return parser.delimitedMany(tokenKindAmp, parser.parseNamedType)
	}
	return nil
}

// parseFieldsDefinition parses
//
//	FieldsDefinition : { FieldDefinition+ }
func (parser *parser) parseFieldsDefinition() []*estree.Node {
	return parser.optionalMany(tokenKindBraceL, parser.parseFieldDefinition, tokenKindBraceR)
}

// parseFieldDefinition parses
//
//	FieldDefinition :
//	  - Description? Name ArgumentsDefinition? : Type Directives[Const]?
func (parser *parser) parseFieldDefinition() *estree.Node {
	start := parser.lexer.token
	description := parser.parseDescription()
	name := parser.parseName()
	arguments := parser.parseArgumentDefs()
	parser.expectToken(tokenKindColon)
	typeReference := parser.parseTypeReference()
	directives := parser.parseConstDirectives()
	return parser.node(start, "FieldDefinition",
		"description", description,
		"name", name,
		"arguments", arguments,
		"type", typeReference,
		"directives", directives,
	)
}

// parseArgumentDefs parses ArgumentsDefinition : ( InputValueDefinition+ )
func (parser *parser) parseArgumentDefs() []*estree.Node {
	return parser.optionalMany(tokenKindParenL, parser.parseInputValueDef, tokenKindParenR)
}

// parseInputValueDef parses
//
//	InputValueDefinition :
//	  - Description? Name : Type DefaultValue? Directives[Const]?
func (parser *parser) parseInputValueDef() *estree.Node {
	start := parser.lexer.token
	description := parser.parseDescription()
	name := parser.parseName()
	parser.expectToken(tokenKindColon)
	typeReference := parser.parseTypeReference()
	var defaultValue *estree.Node
	if parser.expectOptionalToken(tokenKindEquals) {
		defaultValue = parser.parseConstValueLiteral()
	}
	directives := parser.parseConstDirectives()
	return parser.node(start, "InputValueDefinition",
		"description", description,
		"name", name,
		"type", typeReference,
		"defaultValue", defaultValue,
		"directives", directives,
	)
}

// parseInterfaceTypeDefinition parses
//
//	InterfaceTypeDefinition :
//	  - Description? interface Name Directives[Const]? FieldsDefinition?
func (parser *parser) parseInterfaceTypeDefinition() *estree.Node {
	start := parser.lexer.token
	description := parser.parseDescription()
	parser.expectKeyword("interface")
	name := parser.parseName()
	interfaces := parser.parseImplementsInterfaces()
	directives := parser.parseConstDirectives()
	fields := parser.parseFieldsDefinition()
	return parser.node(start, "InterfaceTypeDefinition",
		"description", description,
		"name", name,
		"interfaces", interfaces,
		"directives", directives,
		"fields", fields,
	)
}

// parseUnionTypeDefinition parses
//
//	UnionTypeDefinition :
//	  - Description? union Name Directives[Const]? UnionMemberTypes?
func (parser *parser) parseUnionTypeDefinition() *estree.Node {
	start := parser.lexer.token
	description := parser.parseDescription()
	parser.expectKeyword("union")
	name := parser.parseName()
	directives := parser.parseConstDirectives()
	types := parser.parseUnionMemberTypes()
	return parser.node(start, "UnionTypeDefinition",
		"description", description,
		"name", name,
		"directives", directives,
		"types", types,
	)
}

// parseUnionMemberTypes parses
//
//	UnionMemberTypes :
//	  - = `|`? NamedType
//	  - UnionMemberTypes | NamedType
func (parser *parser) parseUnionMemberTypes() []*estree.Node {
	if parser.expectOptionalToken(tokenKindEquals) {
		return parser.delimitedMany(tokenKindPipe, parser.parseNamedType)
	}
	return nil
}

// parseEnumTypeDefinition parses
//
//	EnumTypeDefinition :
//	  - Description? enum Name Directives[Const]? EnumValuesDefinition?
func (parser *parser) parseEnumTypeDefinition() *estree.Node {
	start := parser.lexer.token
	description := parser.parseDescription()
	parser.expectKeyword("enum")
	name := parser.parseName()
	directives := parser.parseConstDirectives()
	values := parser.parseEnumValuesDefinition()
	return parser.node(start, "EnumTypeDefinition",
		"description", description,
		"name", name,
		"directives", directives,
		"values", values,
	)
}

// parseEnumValuesDefinition parses
//
//	EnumValuesDefinition : { EnumValueDefinition+ }
func (parser *parser) parseEnumValuesDefinition() []*estree.Node {
	return parser.optionalMany(tokenKindBraceL, parser.parseEnumValueDefinition, tokenKindBraceR)
}

// parseEnumValueDefinition parses EnumValueDefinition : Description? EnumValue Directives[Const]?
func (parser *parser) parseEnumValueDefinition() *estree.Node {
	start := parser.lexer.token
	description := parser.parseDescription()
	name := parser.parseEnumValueName()
	directives := parser.parseConstDirectives()
	return parser.node(start, "EnumValueDefinition",
		"description", description,
		"name", name,
		"directives", directives,
	)
}

// parseEnumValueName parses EnumValue : Name but not `true`, `false` or `null`
func (parser *parser) parseEnumValueName() *estree.Node {
	switch parser.lexer.token.value {
	case "true", "false", "null":
		panic(syntaxError(parser.lexer.body, parser.lexer.token.start,
			getTokenDesc(parser.lexer.token)+" is reserved and cannot be used for an enum value."))
	}
	return parser.parseName()
}

// parseInputObjectTypeDefinition parses
//
//	InputObjectTypeDefinition :
//	  - Description? input Name Directives[Const]? InputFieldsDefinition?
func (parser *parser) parseInputObjectTypeDefinition() *estree.Node {
	start := parser.lexer.token
	description := parser.parseDescription()
	parser.expectKeyword("input")
	name := parser.parseName()
	directives := parser.parseConstDirectives()
	fields := parser.parseInputFieldsDefinition()
	return parser.node(start, "InputObjectTypeDefinition",
		"description", description,
		"name", name,
		"directives", directives,
		"fields", fields,
	)
}

// parseInputFieldsDefinition parses
//
//	InputFieldsDefinition : { InputValueDefinition+ }
func (parser *parser) parseInputFieldsDefinition() []*estree.Node {
	return parser.optionalMany(tokenKindBraceL, parser.parseInputValueDef, tokenKindBraceR)
}

// parseTypeSystemExtension parses
//
//	TypeSystemExtension :
//	  - SchemaExtension
//	  - TypeExtension
//	  - DirectiveExtension
//
//	TypeExtension :
//	  - ScalarTypeExtension
//	  - ObjectTypeExtension
//	  - InterfaceTypeExtension
//	  - UnionTypeExtension
//	  - EnumTypeExtension
//	  - InputObjectTypeDefinition
func (parser *parser) parseTypeSystemExtension() *estree.Node {
	keywordToken := parser.lexer.lookahead()

	if keywordToken.kind == tokenKindName {
		switch keywordToken.value {
		case "schema":
			return parser.parseSchemaExtension()
		case "scalar":
			return parser.parseScalarTypeExtension()
		case "type":
			return parser.parseObjectTypeExtension()
		case "interface":
			return parser.parseInterfaceTypeExtension()
		case "union":
			return parser.parseUnionTypeExtension()
		case "enum":
			return parser.parseEnumTypeExtension()
		case "input":
			return parser.parseInputObjectTypeExtension()
		case "directive":
			return parser.parseDirectiveExtension()
		}
	}

	panic(parser.unexpected(keywordToken))
}

// parseSchemaExtension parses
//
//	SchemaExtension :
//	  - extend schema Directives[Const]? { OperationTypeDefinition+ }
//	  - extend schema Directives[Const]
func (parser *parser) parseSchemaExtension() *estree.Node {
	start := parser.lexer.token
	parser.expectKeyword("extend")
	parser.expectKeyword("schema")
	directives := parser.parseConstDirectives()
	operationTypes := parser.optionalMany(tokenKindBraceL, parser.parseOperationTypeDefinition, tokenKindBraceR)
	if directives == nil && operationTypes == nil {
		panic(parser.unexpected(nil))
	}
	return parser.node(start, "SchemaExtension",
		"directives", directives,
		"operationTypes", operationTypes,
	)
}

// parseScalarTypeExtension parses
//
//	ScalarTypeExtension :
//	  - extend scalar Name Directives[Const]
func (parser *parser) parseScalarTypeExtension() *estree.Node {
	start := parser.lexer.token
	parser.expectKeyword("extend")
	parser.expectKeyword("scalar")
	name := parser.parseName()
	directives := parser.parseConstDirectives()
	if directives == nil {
		panic(parser.unexpected(nil))
	}
	return parser.node(start, "ScalarTypeExtension",
		"name", name,
		"directives", directives,
	)
}

// parseObjectTypeExtension parses
//
//	ObjectTypeExtension :
//	 - extend type Name ImplementsInterfaces? Directives[Const]? FieldsDefinition
//	 - extend type Name ImplementsInterfaces? Directives[Const]
//	 - extend type Name ImplementsInterfaces
func (parser *parser) parseObjectTypeExtension() *estree.Node {
	start := parser.lexer.token
	parser.expectKeyword("extend")
	parser.expectKeyword("type")
	name := parser.parseName()
	interfaces := parser.parseImplementsInterfaces()
	directives := parser.parseConstDirectives()
	fields := parser.parseFieldsDefinition()
	if interfaces == nil && directives == nil && fields == nil {
		panic(parser.unexpected(nil))
	}
	return parser.node(start, "ObjectTypeExtension",
		"name", name,
		"interfaces", interfaces,
		"directives", directives,
		"fields", fields,
	)
}

// parseInterfaceTypeExtension parses
//
//	InterfaceTypeExtension :
//	 - extend interface Name ImplementsInterfaces? Directives[Const]? FieldsDefinition
//	 - extend interface Name ImplementsInterfaces? Directives[Const]
//	 - extend interface Name ImplementsInterfaces
func (parser *parser) parseInterfaceTypeExtension() *estree.Node {
	start := parser.lexer.token
	parser.expectKeyword("extend")
	parser.expectKeyword("interface")
	name := parser.parseName()
	interfaces := parser.parseImplementsInterfaces()
	directives := parser.parseConstDirectives()
	fields := parser.parseFieldsDefinition()
	if interfaces == nil && directives == nil && fields == nil {
		panic(parser.unexpected(nil))
	}
	return parser.node(start, "InterfaceTypeExtension",
		"name", name,
		"interfaces", interfaces,
		"directives", directives,
		"fields", fields,
	)
}

// parseUnionTypeExtension parses
//
//	UnionTypeExtension :
//	  - extend union Name Directives[Const]? UnionMemberTypes
//	  - extend union Name Directives[Const]
func (parser *parser) parseUnionTypeExtension() *estree.Node {
	start := parser.lexer.token
	parser.expectKeyword("extend")
	parser.expectKeyword("union")
	name := parser.parseName()
	directives := parser.parseConstDirectives()
	types := parser.parseUnionMemberTypes()
	if directives == nil && types == nil {
		panic(parser.unexpected(nil))
	}
	return parser.node(start, "UnionTypeExtension",
		"name", name,
		"directives", directives,
		"types", types,
	)
}

// parseEnumTypeExtension parses
//
//	EnumTypeExtension :
//	  - extend enum Name Directives[Const]? EnumValuesDefinition
//	  - extend enum Name Directives[Const]
func (parser *parser) parseEnumTypeExtension() *estree.Node {
	start := parser.lexer.token
	parser.expectKeyword("extend")
	parser.expectKeyword("enum")
	name := parser.parseName()
	directives := parser.parseConstDirectives()
	values := parser.parseEnumValuesDefinition()
	if directives == nil && values == nil {
		panic(parser.unexpected(nil))
	}
	return parser.node(start, "EnumTypeExtension",
		"name", name,
		"directives", directives,
		"values", values,
	)
}

// parseInputObjectTypeExtension parses
//
//	InputObjectTypeExtension :
//	  - extend input Name Directives[Const]? InputFieldsDefinition
//	  - extend input Name Directives[Const]
func (parser *parser) parseInputObjectTypeExtension() *estree.Node {
	start := parser.lexer.token
	parser.expectKeyword("extend")
	parser.expectKeyword("input")
	name := parser.parseName()
	directives := parser.parseConstDirectives()
	fields := parser.parseInputFieldsDefinition()
	if directives == nil && fields == nil {
		panic(parser.unexpected(nil))
	}
	return parser.node(start, "InputObjectTypeExtension",
		"name", name,
		"directives", directives,
		"fields", fields,
	)
}

func (parser *parser) parseDirectiveExtension() *estree.Node {
	start := parser.lexer.token
	parser.expectKeyword("extend")
	parser.expectKeyword("directive")
	parser.expectToken(tokenKindAt)
	name := parser.parseName()
	directives := parser.parseConstDirectives()
	if directives == nil {
		panic(parser.unexpected(nil))
	}
	return parser.node(start, "DirectiveExtension",
		"name", name,
		"directives", directives,
	)
}

// parseDirectiveDefinition parses
//
//	DirectiveDefinition :
//	  - Description? directive @ Name ArgumentsDefinition? Directives[Const]? `repeatable`? on DirectiveLocations
func (parser *parser) parseDirectiveDefinition() *estree.Node {
	start := parser.lexer.token
	description := parser.parseDescription()
	parser.expectKeyword("directive")
	parser.expectToken(tokenKindAt)
	name := parser.parseName()
	arguments := parser.parseArgumentDefs()
	directives := parser.parseConstDirectives()
	repeatable := parser.expectOptionalKeyword("repeatable")
	parser.expectKeyword("on")
	locations := parser.parseDirectiveLocations()
	return parser.node(start, "DirectiveDefinition",
		"description", description,
		"name", name,
		"arguments", arguments,
		"directives", directives,
		"repeatable", repeatable,
		"locations", locations,
	)
}

// parseDirectiveLocations parses
//
//	DirectiveLocations :
//	  - `|`? DirectiveLocation
//	  - DirectiveLocations | DirectiveLocation
func (parser *parser) parseDirectiveLocations() []*estree.Node {
	return parser.delimitedMany(tokenKindPipe, parser.parseDirectiveLocation)
}

// directiveLocation is language/directiveLocation.js's DirectiveLocation, whose keys are its values.
var directiveLocation = map[string]bool{
	"QUERY":                        true,
	"MUTATION":                     true,
	"SUBSCRIPTION":                 true,
	"FIELD":                        true,
	"FRAGMENT_DEFINITION":          true,
	"FRAGMENT_SPREAD":              true,
	"INLINE_FRAGMENT":              true,
	"VARIABLE_DEFINITION":          true,
	"FRAGMENT_VARIABLE_DEFINITION": true,
	"SCHEMA":                       true,
	"SCALAR":                       true,
	"OBJECT":                       true,
	"FIELD_DEFINITION":             true,
	"ARGUMENT_DEFINITION":          true,
	"INTERFACE":                    true,
	"UNION":                        true,
	"ENUM":                         true,
	"ENUM_VALUE":                   true,
	"INPUT_OBJECT":                 true,
	"INPUT_FIELD_DEFINITION":       true,
	"DIRECTIVE_DEFINITION":         true,
}

// parseDirectiveLocation parses
//
//	DirectiveLocation :
//	  - ExecutableDirectiveLocation
//	  - TypeSystemDirectiveLocation
//
//	ExecutableDirectiveLocation : one of
//	  `QUERY`
//	  `MUTATION`
//	  `SUBSCRIPTION`
//	  `FIELD`
//	  `FRAGMENT_DEFINITION`
//	  `FRAGMENT_SPREAD`
//	  `INLINE_FRAGMENT`
//	  `VARIABLE_DEFINITION`
//	  `FRAGMENT_VARIABLE_DEFINITION`
//
//	TypeSystemDirectiveLocation : one of
//	  `SCHEMA`
//	  `SCALAR`
//	  `OBJECT`
//	  `FIELD_DEFINITION`
//	  `ARGUMENT_DEFINITION`
//	  `INTERFACE`
//	  `UNION`
//	  `ENUM`
//	  `ENUM_VALUE`
//	  `INPUT_OBJECT`
//	  `INPUT_FIELD_DEFINITION`
//	  `DIRECTIVE_DEFINITION`
func (parser *parser) parseDirectiveLocation() *estree.Node {
	start := parser.lexer.token
	name := parser.parseName()
	if directiveLocation[name.String("value")] {
		return name
	}
	panic(parser.unexpected(start))
}

// Core parsing utility functions

// node returns a node that, if configured to do so, sets a "loc" field as a location object, used to
// identify the place in the source that created a given parsed object.
//
// graphql-js reads lastToken after the object literal is built, so the end covers every field; here
// the fields are the arguments, which Go evaluates before the call, to the same effect. A field that
// is a nil list is stored as nil itself: a typed nil slice in an interface would read as present.
func (parser *parser) node(startToken *token, kind string, keysAndValues ...any) *estree.Node {
	for index := 1; index < len(keysAndValues); index += 2 {
		if list, isList := keysAndValues[index].([]*estree.Node); isList && list == nil {
			keysAndValues[index] = nil
		}
	}
	return estree.New(kind, startToken.start, parser.lexer.lastToken.end, keysAndValues...)
}

// peek determines if the next token is of a given kind.
func (parser *parser) peek(kind tokenKind) bool {
	return parser.lexer.token.kind == kind
}

// expectToken: if the next token is of the given kind, return that token after advancing the lexer.
// Otherwise, do not change the parser state and throw an error.
func (parser *parser) expectToken(kind tokenKind) *token {
	token := parser.lexer.token
	if token.kind == kind {
		parser.advanceLexer()
		return token
	}

	panic(syntaxError(parser.lexer.body, token.start,
		"Expected "+getTokenKindDesc(kind)+", found "+getTokenDesc(token)+"."))
}

// expectOptionalToken: if the next token is of the given kind, return "true" after advancing the
// lexer. Otherwise, do not change the parser state and return "false".
func (parser *parser) expectOptionalToken(kind tokenKind) bool {
	token := parser.lexer.token
	if token.kind == kind {
		parser.advanceLexer()
		return true
	}
	return false
}

// expectKeyword: if the next token is a given keyword, advance the lexer. Otherwise, do not change the
// parser state and throw an error.
func (parser *parser) expectKeyword(value string) {
	token := parser.lexer.token
	if token.kind == tokenKindName && token.value == value {
		parser.advanceLexer()
	} else {
		panic(syntaxError(parser.lexer.body, token.start,
			`Expected "`+value+`", found `+getTokenDesc(token)+"."))
	}
}

// expectOptionalKeyword: if the next token is a given keyword, return "true" after advancing the
// lexer. Otherwise, do not change the parser state and return "false".
func (parser *parser) expectOptionalKeyword(value string) bool {
	token := parser.lexer.token
	if token.kind == tokenKindName && token.value == value {
		parser.advanceLexer()
		return true
	}
	return false
}

// unexpected is a helper function for creating an error when an unexpected lexed token is encountered.
func (parser *parser) unexpected(atToken *token) *SyntaxError {
	token := atToken
	if token == nil {
		token = parser.lexer.token
	}
	return syntaxError(parser.lexer.body, token.start, "Unexpected "+getTokenDesc(token)+".")
}

// any returns a possibly empty list of parse nodes, determined by the parseFn. This list begins with a
// lex token of openKind and ends with a lex token of closeKind. Advances the parser to the next lex
// token after the closing token.
func (parser *parser) any(openKind tokenKind, parseFn func() *estree.Node, closeKind tokenKind) []*estree.Node {
	parser.expectToken(openKind)
	nodes := []*estree.Node{}
	for !parser.expectOptionalToken(closeKind) {
		nodes = append(nodes, parseFn())
	}
	return nodes
}

// optionalMany returns a list of parse nodes, determined by the parseFn. It can be empty only if open
// token is missing otherwise it will always return non-empty list that begins with a lex token of
// openKind and ends with a lex token of closeKind. Advances the parser to the next lex token after the
// closing token.
func (parser *parser) optionalMany(openKind tokenKind, parseFn func() *estree.Node, closeKind tokenKind) []*estree.Node {
	if parser.expectOptionalToken(openKind) {
		nodes := []*estree.Node{}
		for {
			nodes = append(nodes, parseFn())
			if parser.expectOptionalToken(closeKind) {
				break
			}
		}
		return nodes
	}
	return nil
}

// many returns a non-empty list of parse nodes, determined by the parseFn. This list begins with a lex
// token of openKind and ends with a lex token of closeKind. Advances the parser to the next lex token
// after the closing token.
func (parser *parser) many(openKind tokenKind, parseFn func() *estree.Node, closeKind tokenKind) []*estree.Node {
	parser.expectToken(openKind)
	nodes := []*estree.Node{}
	for {
		nodes = append(nodes, parseFn())
		if parser.expectOptionalToken(closeKind) {
			break
		}
	}
	return nodes
}

// delimitedMany returns a non-empty list of parse nodes, determined by the parseFn. This list may
// begin with a lex token of delimiterKind followed by items separated by lex tokens of tokenKind.
// Advances the parser to the next lex token after last item in the list.
func (parser *parser) delimitedMany(delimiterKind tokenKind, parseFn func() *estree.Node) []*estree.Node {
	parser.expectOptionalToken(delimiterKind)

	nodes := []*estree.Node{}
	for {
		nodes = append(nodes, parseFn())
		if !parser.expectOptionalToken(delimiterKind) {
			break
		}
	}
	return nodes
}

// advanceLexer advances the lexer. graphql-js counts the tokens here against maxTokens, which Prettier
// leaves unset.
func (parser *parser) advanceLexer() {
	parser.lexer.advance()
}

// getTokenDesc is a helper function to describe a token as a string for debugging.
func getTokenDesc(token *token) string {
	description := getTokenKindDesc(token.kind)
	if token.hasValue {
		description += ` "` + token.value + `"`
	}
	return description
}

// getTokenKindDesc is a helper function to describe a token kind as a string for debugging.
func getTokenKindDesc(kind tokenKind) string {
	if isPunctuatorTokenKind(kind) {
		return `"` + string(kind) + `"`
	}
	return string(kind)
}
