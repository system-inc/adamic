// A port of cohere's internal/format/graphql/parser.go to Adamic 0.1: graphql-js 17.0.2,
// language/parser.js, parse and the Parser class it drives, with parseComments from Prettier's
// src/language-graphql/parser-graphql.js.
//
// As in the Go: Prettier's options (experimentalFragmentArguments true, noLocation false, maxTokens
// unset), and parse alone of graphql-js's entries, so no schema coordinates.
//
// Where the port's shape differs from the Go's, and why:
//
//   - Nodes. The Go's are estree.Nodes, bags of properties in graphql-js's creation order. Here a node
//     is a GraphNode: its kind, its location, and its fields as a list in that order, so that a field
//     graphql-js writes as undefined is in the list holding Undefined, and one it never writes (a
//     FragmentSpread's arguments, when none are written) is not in it. A node's children are its
//     indexes in the document's table of nodes, not the nodes themselves: a node holding a list of
//     nodes it was built from is what 0.1's cycle rule refuses (mediaquery's GAPS.md, gaps 4 and 5),
//     and a document's lists can be long, which rules out copying each list to grow it, as the
//     mediaquery port does (the values port keeps a table too).
//   - The list helpers. graphql-js's any, many, optionalMany and delimitedMany take the item's parse
//     function as a value. A function value that can throw doesn't lower yet (GAPS.md, gap 1), and
//     neither does a call to a method declared later (gitignore's GAPS.md, gap 3), which a dispatching
//     method in the helpers' place would be. So each list is its helper's
//     loop, written where the helper is called, and marked with the helper's name.
//   - The two cycles. parseValueLiteral, parseList, parseObject and parseObjectField call each other,
//     and so do parseSelectionSet, parseSelection, parseField and parseFragment; with gap 3 no order
//     of methods lowers either, so each cycle is one method, parseValueLiteral and parseSelectionSet,
//     with the others written inside it where they're called, and marked.
//   - Methods are in callee-first order (gap 3), not the Go's.
//   - Optional lists. graphql-js returns a list or undefined; here such a method returns the field's
//     Value, a List or Undefined, since returning undefined from a function that returns an array or
//     undefined lowers to C that clang refuses (suppression's GAPS.md, gap 2).

import { panic } from 'adamic';
import { isPunctuatorTokenKind, Lexer, syntaxError } from './lexer.ts';
import type { Token, TokenKind } from './token.ts';

// Value is what a node's field holds. Text is in the output form (lexer.ts).
export type Value =
	| { readonly kind: 'Undefined' }
	| { readonly kind: 'Text'; readonly text: string }
	| { readonly kind: 'Flag'; readonly flag: boolean }
	| { readonly kind: 'Node'; readonly node: number }
	| { readonly kind: 'List'; readonly nodes: readonly number[] };

// Field is one of a node's properties.
export interface Field {
	readonly key: string;
	readonly value: Value;
}

// GraphNode is one node of graphql-js's AST: its kind, its loc's start and end, and its other fields in
// the order graphql-js writes them.
export class GraphNode {
	readonly kind: string;
	readonly start: number;
	readonly end: number;
	readonly fields: readonly Field[];

	constructor(kind: string, start: number, end: number, fields: readonly Field[]) {
		this.kind = kind;
		this.start = start;
		this.end = end;
		this.fields = fields;
	}
}

// Comment is one of parseComments's comments: a Comment token's location and the text after its "#".
export interface Comment {
	readonly start: number;
	readonly end: number;
	readonly value: string;
}

// Document is what parse gives: the table of every node, the Document node's index in it, and the
// comments.
export interface Document {
	readonly nodes: readonly GraphNode[];
	readonly root: number;
	readonly comments: readonly Comment[];
}

const undefinedValue: Value = { kind: 'Undefined' };

function text(key: string, value: string): Field {
	return { key, value: { kind: 'Text', text: value } };
}

function flag(key: string, value: boolean): Field {
	return { key, value: { kind: 'Flag', flag: value } };
}

// child is a field holding a node, or undefined.
function child(key: string, node: number | undefined): Field {
	return { key, value: node === undefined ? undefinedValue : { kind: 'Node', node } };
}

function list(key: string, nodes: readonly number[]): Field {
	return { key, value: { kind: 'List', nodes } };
}

function field(key: string, value: Value): Field {
	return { key, value };
}

// isDirectiveLocation is whether a name is one of language/directiveLocation.js's DirectiveLocation,
// whose keys are its values.
function isDirectiveLocation(name: string): boolean {
	switch (name) {
		case 'QUERY':
		case 'MUTATION':
		case 'SUBSCRIPTION':
		case 'FIELD':
		case 'FRAGMENT_DEFINITION':
		case 'FRAGMENT_SPREAD':
		case 'INLINE_FRAGMENT':
		case 'VARIABLE_DEFINITION':
		case 'FRAGMENT_VARIABLE_DEFINITION':
		case 'SCHEMA':
		case 'SCALAR':
		case 'OBJECT':
		case 'FIELD_DEFINITION':
		case 'ARGUMENT_DEFINITION':
		case 'INTERFACE':
		case 'UNION':
		case 'ENUM':
		case 'ENUM_VALUE':
		case 'INPUT_OBJECT':
		case 'INPUT_FIELD_DEFINITION':
		case 'DIRECTIVE_DEFINITION':
			return true;
	}
	return false;
}

// getTokenKindDesc is a helper function to describe a token kind as a string for debugging.
function getTokenKindDesc(kind: TokenKind): string {
	return isPunctuatorTokenKind(kind) ? `"${kind}"` : kind;
}

// getTokenDesc is a helper function to describe a token as a string for debugging.
function getTokenDesc(token: Token): string {
	const value = token.value;
	return getTokenKindDesc(token.kind) + (value !== undefined ? ` "${value}"` : '');
}

// Parser is graphql-js's Parser class.
class Parser {
	readonly lexer: Lexer;
	// Every node made, in the order made; a node's children are indexes into it.
	readonly nodes: GraphNode[] = [];

	constructor(source: string) {
		this.lexer = new Lexer(source);
	}

	// Core parsing utility functions

	// node returns a node that, if configured to do so, sets a "loc" field as a location object, used to
	// identify the place in the source that created a given parsed object. graphql-js reads lastToken
	// after the object literal is built, so the end covers every field; here the fields are an argument,
	// made before the call, to the same effect.
	node(startToken: Token, kind: string, fields: readonly Field[]): number {
		this.nodes.push(new GraphNode(kind, startToken.start, this.lexer.lastToken.end, fields));
		return this.nodes.length - 1;
	}

	// advanceLexer advances the lexer. graphql-js counts the tokens here against maxTokens, which
	// Prettier leaves unset.
	advanceLexer(): void {
		this.lexer.advance();
	}

	// peek determines if the next token is of a given kind.
	peek(kind: TokenKind): boolean {
		return this.lexer.token.kind === kind;
	}

	// expectToken: if the next token is of the given kind, return that token after advancing the lexer.
	// Otherwise, do not change the parser state and throw an error.
	expectToken(kind: TokenKind): Token {
		const token = this.lexer.token;
		if (token.kind === kind) {
			this.advanceLexer();
			return token;
		}

		throw new Error(syntaxError(this.lexer.body, token.start, `Expected ${getTokenKindDesc(kind)}, found ${getTokenDesc(token)}.`));
	}

	// expectOptionalToken: if the next token is of the given kind, return "true" after advancing the
	// lexer. Otherwise, do not change the parser state and return "false".
	expectOptionalToken(kind: TokenKind): boolean {
		const token = this.lexer.token;
		if (token.kind === kind) {
			this.advanceLexer();
			return true;
		}
		return false;
	}

	// expectKeyword: if the next token is a given keyword, advance the lexer. Otherwise, do not change the
	// parser state and throw an error.
	expectKeyword(value: string): void {
		const token = this.lexer.token;
		if (token.kind === 'Name' && token.value === value) {
			this.advanceLexer();
		} else {
			throw new Error(syntaxError(this.lexer.body, token.start, `Expected "${value}", found ${getTokenDesc(token)}.`));
		}
	}

	// expectOptionalKeyword: if the next token is a given keyword, return "true" after advancing the
	// lexer. Otherwise, do not change the parser state and return "false".
	expectOptionalKeyword(value: string): boolean {
		const token = this.lexer.token;
		if (token.kind === 'Name' && token.value === value) {
			this.advanceLexer();
			return true;
		}
		return false;
	}

	// unexpected is a helper function for creating an error when an unexpected lexed token is
	// encountered: its message, which the caller throws (GAPS.md, gap 2).
	unexpected(atToken: Token | undefined): string {
		const token = atToken ?? this.lexer.token;
		return syntaxError(this.lexer.body, token.start, `Unexpected ${getTokenDesc(token)}.`);
	}

	// parseName converts a name lex token into a name parse node.
	parseName(): number {
		const token = this.expectToken('Name');
		return this.node(token, 'Name', [text('value', token.value ?? '')]);
	}

	// Implements the parsing rules in the Types section.

	// parseNamedType parses NamedType : Name
	parseNamedType(): number {
		const start = this.lexer.token;
		return this.node(start, 'NamedType', [child('name', this.parseName())]);
	}

	// parseTypeReference parses
	//
	//	Type :
	//	  - NamedType
	//	  - ListType
	//	  - NonNullType
	parseTypeReference(): number {
		const start = this.lexer.token;
		let typeReference: number;
		if (this.expectOptionalToken('[')) {
			const innerType = this.parseTypeReference();
			this.expectToken(']');
			typeReference = this.node(start, 'ListType', [child('type', innerType)]);
		} else {
			typeReference = this.parseNamedType();
		}

		if (this.expectOptionalToken('!')) {
			return this.node(start, 'NonNullType', [child('type', typeReference)]);
		}

		return typeReference;
	}

	// Implements the parsing rules in the Values section.

	parseStringLiteral(): number {
		const token = this.lexer.token;
		this.advanceLexer();
		return this.node(token, 'StringValue', [text('value', token.value ?? ''), flag('block', token.kind === 'BlockString')]);
	}

	// parseVariable parses Variable : $ Name
	parseVariable(): number {
		const start = this.lexer.token;
		this.expectToken('$');
		return this.node(start, 'Variable', [child('name', this.parseName())]);
	}

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
	//
	// with parseList, parseObject and parseObjectField written in it, where it calls them (gap 3).
	parseValueLiteral(isConst: boolean): number {
		const token = this.lexer.token;
		switch (token.kind) {
			case '[': {
				// parseList parses
				//
				//	ListValue[Const] :
				//	  - [ ]
				//	  - [ Value[?Const]+ ]
				const start = this.lexer.token;
				// any('[', parseValueLiteral(isConst), ']')
				this.expectToken('[');
				const values: number[] = [];
				while (!this.expectOptionalToken(']')) {
					values.push(this.parseValueLiteral(isConst));
				}
				return this.node(start, 'ListValue', [list('values', values)]);
			}
			case '{': {
				// parseObject parses
				//
				//	ObjectValue[Const] :
				//	  - { }
				//	  - { ObjectField[?Const]+ }
				const start = this.lexer.token;
				// any('{', parseObjectField(isConst), '}')
				this.expectToken('{');
				const fields: number[] = [];
				while (!this.expectOptionalToken('}')) {
					// parseObjectField parses ObjectField[Const] : Name : Value[?Const]
					const fieldStart = this.lexer.token;
					const name = this.parseName();
					this.expectToken(':');
					fields.push(this.node(fieldStart, 'ObjectField', [child('name', name), child('value', this.parseValueLiteral(isConst))]));
				}
				return this.node(start, 'ObjectValue', [list('fields', fields)]);
			}
			case 'Int':
				this.advanceLexer();
				return this.node(token, 'IntValue', [text('value', token.value ?? '')]);
			case 'Float':
				this.advanceLexer();
				return this.node(token, 'FloatValue', [text('value', token.value ?? '')]);
			case 'String':
			case 'BlockString':
				return this.parseStringLiteral();
			case 'Name':
				this.advanceLexer();
				switch (token.value) {
					case 'true':
						return this.node(token, 'BooleanValue', [flag('value', true)]);
					case 'false':
						return this.node(token, 'BooleanValue', [flag('value', false)]);
					case 'null':
						return this.node(token, 'NullValue', []);
					default:
						return this.node(token, 'EnumValue', [text('value', token.value ?? '')]);
				}
			case '$':
				if (isConst) {
					this.expectToken('$');
					if (this.lexer.token.kind === 'Name') {
						const variableName = this.lexer.token.value ?? '';
						throw new Error(syntaxError(this.lexer.body, token.start, `Unexpected variable "$${variableName}" in constant value.`));
					}
					throw new Error(this.unexpected(token));
				}
				return this.parseVariable();
			default:
				throw new Error(this.unexpected(undefined));
		}
	}

	// Implements the parsing rules in the Directives section.

	// parseArguments parses Arguments[Const] : ( Argument[?Const]+ ), with parseArgument and
	// parseConstArgument, which parse Argument[Const] : Name : Value[?Const], written in it.
	parseArguments(isConst: boolean): Value {
		// optionalMany('(', isConst ? parseConstArgument : parseArgument, ')')
		if (!this.expectOptionalToken('(')) {
			return undefinedValue;
		}
		const nodes: number[] = [];
		do {
			const start = this.lexer.token;
			const name = this.parseName();
			this.expectToken(':');
			nodes.push(this.node(start, 'Argument', [child('name', name), child('value', this.parseValueLiteral(isConst))]));
		} while (!this.expectOptionalToken(')'));
		return { kind: 'List', nodes };
	}

	// parseDirective parses
	//
	//	Directive[Const] : @ Name Arguments[?Const]?
	parseDirective(isConst: boolean): number {
		const start = this.lexer.token;
		this.expectToken('@');
		const name = this.parseName();
		return this.node(start, 'Directive', [child('name', name), field('arguments', this.parseArguments(isConst))]);
	}

	// parseDirectives parses Directives[Const] : Directive[?Const]+
	parseDirectives(isConst: boolean): Value {
		const directives: number[] = [];
		while (this.peek('@')) {
			directives.push(this.parseDirective(isConst));
		}
		if (directives.length > 0) {
			return { kind: 'List', nodes: directives };
		}
		return undefinedValue;
	}

	parseConstDirectives(): Value {
		return this.parseDirectives(true);
	}

	// Implements the parsing rules in the Operations section.

	// parseFragmentName parses FragmentName : Name but not `on`
	parseFragmentName(): number {
		if (this.lexer.token.value === 'on') {
			throw new Error(this.unexpected(undefined));
		}
		return this.parseName();
	}

	// parseFragmentArguments is optionalMany('(', parseFragmentArgument, ')'), with parseFragmentArgument
	// written in it.
	parseFragmentArguments(): Value {
		if (!this.expectOptionalToken('(')) {
			return undefinedValue;
		}
		const nodes: number[] = [];
		do {
			const start = this.lexer.token;
			const name = this.parseName();
			this.expectToken(':');
			nodes.push(this.node(start, 'FragmentArgument', [child('name', name), child('value', this.parseValueLiteral(false))]));
		} while (!this.expectOptionalToken(')'));
		return { kind: 'List', nodes };
	}

	// parseSelectionSet parses
	//
	//	SelectionSet : { Selection+ }
	//
	// with parseSelection, parseField and parseFragment written in it, where it calls them (gap 3).
	parseSelectionSet(): number {
		const start = this.lexer.token;
		// many('{', parseSelection, '}')
		this.expectToken('{');
		const selections: number[] = [];
		do {
			// parseSelection parses
			//
			//	Selection :
			//	  - Field
			//	  - FragmentSpread
			//	  - InlineFragment
			if (this.peek('...')) {
				// parseFragment corresponds to both FragmentSpread and InlineFragment in the spec.
				//
				//	FragmentSpread : ... FragmentName Arguments? Directives?
				//
				//	InlineFragment : ... TypeCondition? Directives? SelectionSet
				const fragmentStart = this.lexer.token;
				this.expectToken('...');

				const hasTypeCondition = this.expectOptionalKeyword('on');
				if (!hasTypeCondition && this.peek('Name')) {
					const name = this.parseFragmentName();
					// Prettier's experimentalFragmentArguments is true.
					if (this.peek('(')) {
						const fragmentArguments = this.parseFragmentArguments();
						selections.push(
							this.node(fragmentStart, 'FragmentSpread', [child('name', name), field('arguments', fragmentArguments), field('directives', this.parseDirectives(false))]),
						);
					} else {
						selections.push(this.node(fragmentStart, 'FragmentSpread', [child('name', name), field('directives', this.parseDirectives(false))]));
					}
				} else {
					const typeCondition = hasTypeCondition ? this.parseNamedType() : undefined;
					const directives = this.parseDirectives(false);
					selections.push(
						this.node(fragmentStart, 'InlineFragment', [child('typeCondition', typeCondition), field('directives', directives), child('selectionSet', this.parseSelectionSet())]),
					);
				}
			} else {
				// parseField parses
				//
				//	Field : Alias? Name Arguments? Directives? SelectionSet?
				//
				//	Alias : Name :
				const fieldStart = this.lexer.token;

				const nameOrAlias = this.parseName();
				let alias: number | undefined = undefined;
				let name: number;
				if (this.expectOptionalToken(':')) {
					alias = nameOrAlias;
					name = this.parseName();
				} else {
					name = nameOrAlias;
				}

				const fieldArguments = this.parseArguments(false);
				const directives = this.parseDirectives(false);
				const selectionSet = this.peek('{') ? this.parseSelectionSet() : undefined;
				selections.push(
					this.node(fieldStart, 'Field', [
						child('alias', alias),
						child('name', name),
						field('arguments', fieldArguments),
						field('directives', directives),
						child('selectionSet', selectionSet),
					]),
				);
			}
		} while (!this.expectOptionalToken('}'));
		return this.node(start, 'SelectionSet', [list('selections', selections)]);
	}

	// Implements the parsing rules in the Type Definition section.

	peekDescription(): boolean {
		return this.peek('String') || this.peek('BlockString');
	}

	// parseDescription parses Description : StringValue
	parseDescription(): number | undefined {
		if (this.peekDescription()) {
			return this.parseStringLiteral();
		}
		return undefined;
	}

	// parseVariableDefinition parses VariableDefinition : Variable : Type DefaultValue? Directives[Const]?
	parseVariableDefinition(): number {
		const start = this.lexer.token;
		const description = this.parseDescription();
		const variable = this.parseVariable();
		this.expectToken(':');
		const typeReference = this.parseTypeReference();
		const defaultValue = this.expectOptionalToken('=') ? this.parseValueLiteral(true) : undefined;
		return this.node(start, 'VariableDefinition', [
			child('description', description),
			child('variable', variable),
			child('type', typeReference),
			child('defaultValue', defaultValue),
			field('directives', this.parseConstDirectives()),
		]);
	}

	// parseVariableDefinitions parses VariableDefinitions : ( VariableDefinition+ )
	parseVariableDefinitions(): Value {
		// optionalMany('(', parseVariableDefinition, ')')
		if (!this.expectOptionalToken('(')) {
			return undefinedValue;
		}
		const nodes: number[] = [];
		do {
			nodes.push(this.parseVariableDefinition());
		} while (!this.expectOptionalToken(')'));
		return { kind: 'List', nodes };
	}

	// parseOperationType parses OperationType : one of query mutation subscription
	parseOperationType(): string {
		const operationToken = this.expectToken('Name');
		switch (operationToken.value) {
			case 'query':
				return 'query';
			case 'mutation':
				return 'mutation';
			case 'subscription':
				return 'subscription';
		}

		throw new Error(this.unexpected(operationToken));
	}

	// parseOperationDefinition parses
	//
	//	OperationDefinition :
	//	 - SelectionSet
	//	 - OperationType Name? VariableDefinitions? Directives? SelectionSet
	parseOperationDefinition(): number {
		const start = this.lexer.token;
		if (this.peek('{')) {
			return this.node(start, 'OperationDefinition', [
				text('operation', 'query'),
				field('description', undefinedValue),
				field('name', undefinedValue),
				field('variableDefinitions', undefinedValue),
				field('directives', undefinedValue),
				child('selectionSet', this.parseSelectionSet()),
			]);
		}
		const description = this.parseDescription();
		const operation = this.parseOperationType();
		const name = this.peek('Name') ? this.parseName() : undefined;
		return this.node(start, 'OperationDefinition', [
			text('operation', operation),
			child('description', description),
			child('name', name),
			field('variableDefinitions', this.parseVariableDefinitions()),
			field('directives', this.parseDirectives(false)),
			child('selectionSet', this.parseSelectionSet()),
		]);
	}

	// parseFragmentDefinition parses
	//
	//	FragmentDefinition :
	//	  - fragment FragmentName VariableDefinitions? on TypeCondition Directives? SelectionSet
	//
	//	TypeCondition : NamedType
	//
	// in the shape Prettier's experimentalFragmentArguments gives it.
	parseFragmentDefinition(): number {
		const start = this.lexer.token;
		const description = this.parseDescription();
		this.expectKeyword('fragment');
		const name = this.parseFragmentName();
		const variableDefinitions = this.parseVariableDefinitions();
		this.expectKeyword('on');
		const typeCondition = this.parseNamedType();
		const directives = this.parseDirectives(false);
		return this.node(start, 'FragmentDefinition', [
			child('description', description),
			child('name', name),
			field('variableDefinitions', variableDefinitions),
			child('typeCondition', typeCondition),
			field('directives', directives),
			child('selectionSet', this.parseSelectionSet()),
		]);
	}

	// parseOperationTypeDefinition parses OperationTypeDefinition : OperationType : NamedType
	parseOperationTypeDefinition(): number {
		const start = this.lexer.token;
		const operation = this.parseOperationType();
		this.expectToken(':');
		const typeReference = this.parseNamedType();
		return this.node(start, 'OperationTypeDefinition', [text('operation', operation), child('type', typeReference)]);
	}

	// parseOperationTypeDefinitions is many('{', parseOperationTypeDefinition, '}') when required, and
	// optionalMany when not.
	parseOperationTypeDefinitions(required: boolean): Value {
		if (required) {
			this.expectToken('{');
		} else if (!this.expectOptionalToken('{')) {
			return undefinedValue;
		}
		const nodes: number[] = [];
		do {
			nodes.push(this.parseOperationTypeDefinition());
		} while (!this.expectOptionalToken('}'));
		return { kind: 'List', nodes };
	}

	// parseSchemaDefinition parses
	//
	//	SchemaDefinition : Description? schema Directives[Const]? { OperationTypeDefinition+ }
	parseSchemaDefinition(): number {
		const start = this.lexer.token;
		const description = this.parseDescription();
		this.expectKeyword('schema');
		const directives = this.parseConstDirectives();
		const operationTypes = this.parseOperationTypeDefinitions(true);
		return this.node(start, 'SchemaDefinition', [child('description', description), field('directives', directives), field('operationTypes', operationTypes)]);
	}

	// parseScalarTypeDefinition parses ScalarTypeDefinition : Description? scalar Name Directives[Const]?
	parseScalarTypeDefinition(): number {
		const start = this.lexer.token;
		const description = this.parseDescription();
		this.expectKeyword('scalar');
		const name = this.parseName();
		const directives = this.parseConstDirectives();
		return this.node(start, 'ScalarTypeDefinition', [child('description', description), child('name', name), field('directives', directives)]);
	}

	// delimitedNamedTypes is delimitedMany(delimiterKind, parseNamedType): a non-empty list of parse
	// nodes, which may begin with a lex token of delimiterKind followed by items separated by lex tokens
	// of delimiterKind.
	delimitedNamedTypes(delimiterKind: TokenKind): Value {
		this.expectOptionalToken(delimiterKind);

		const nodes: number[] = [];
		do {
			nodes.push(this.parseNamedType());
		} while (this.expectOptionalToken(delimiterKind));
		return { kind: 'List', nodes };
	}

	// parseImplementsInterfaces parses
	//
	//	ImplementsInterfaces :
	//	  - implements `&`? NamedType
	//	  - ImplementsInterfaces & NamedType
	parseImplementsInterfaces(): Value {
		if (this.expectOptionalKeyword('implements')) {
			return this.delimitedNamedTypes('&');
		}
		return undefinedValue;
	}

	// parseInputValueDef parses
	//
	//	InputValueDefinition :
	//	  - Description? Name : Type DefaultValue? Directives[Const]?
	parseInputValueDef(): number {
		const start = this.lexer.token;
		const description = this.parseDescription();
		const name = this.parseName();
		this.expectToken(':');
		const typeReference = this.parseTypeReference();
		const defaultValue = this.expectOptionalToken('=') ? this.parseValueLiteral(true) : undefined;
		const directives = this.parseConstDirectives();
		return this.node(start, 'InputValueDefinition', [
			child('description', description),
			child('name', name),
			child('type', typeReference),
			child('defaultValue', defaultValue),
			field('directives', directives),
		]);
	}

	// parseInputValueDefs is optionalMany(openKind, parseInputValueDef, closeKind): parseArgumentDefs,
	// ArgumentsDefinition : ( InputValueDefinition+ ), and parseInputFieldsDefinition,
	// InputFieldsDefinition : { InputValueDefinition+ }.
	parseInputValueDefs(openKind: TokenKind, closeKind: TokenKind): Value {
		if (!this.expectOptionalToken(openKind)) {
			return undefinedValue;
		}
		const nodes: number[] = [];
		do {
			nodes.push(this.parseInputValueDef());
		} while (!this.expectOptionalToken(closeKind));
		return { kind: 'List', nodes };
	}

	// parseFieldDefinition parses
	//
	//	FieldDefinition :
	//	  - Description? Name ArgumentsDefinition? : Type Directives[Const]?
	parseFieldDefinition(): number {
		const start = this.lexer.token;
		const description = this.parseDescription();
		const name = this.parseName();
		const fieldArguments = this.parseInputValueDefs('(', ')');
		this.expectToken(':');
		const typeReference = this.parseTypeReference();
		const directives = this.parseConstDirectives();
		return this.node(start, 'FieldDefinition', [
			child('description', description),
			child('name', name),
			field('arguments', fieldArguments),
			child('type', typeReference),
			field('directives', directives),
		]);
	}

	// parseFieldsDefinition parses
	//
	//	FieldsDefinition : { FieldDefinition+ }
	parseFieldsDefinition(): Value {
		// optionalMany('{', parseFieldDefinition, '}')
		if (!this.expectOptionalToken('{')) {
			return undefinedValue;
		}
		const nodes: number[] = [];
		do {
			nodes.push(this.parseFieldDefinition());
		} while (!this.expectOptionalToken('}'));
		return { kind: 'List', nodes };
	}

	// parseObjectTypeDefinition parses
	//
	//	ObjectTypeDefinition :
	//	  Description?
	//	  type Name ImplementsInterfaces? Directives[Const]? FieldsDefinition?
	parseObjectTypeDefinition(): number {
		const start = this.lexer.token;
		const description = this.parseDescription();
		this.expectKeyword('type');
		const name = this.parseName();
		const interfaces = this.parseImplementsInterfaces();
		const directives = this.parseConstDirectives();
		const fields = this.parseFieldsDefinition();
		return this.node(start, 'ObjectTypeDefinition', [
			child('description', description),
			child('name', name),
			field('interfaces', interfaces),
			field('directives', directives),
			field('fields', fields),
		]);
	}

	// parseInterfaceTypeDefinition parses
	//
	//	InterfaceTypeDefinition :
	//	  - Description? interface Name Directives[Const]? FieldsDefinition?
	parseInterfaceTypeDefinition(): number {
		const start = this.lexer.token;
		const description = this.parseDescription();
		this.expectKeyword('interface');
		const name = this.parseName();
		const interfaces = this.parseImplementsInterfaces();
		const directives = this.parseConstDirectives();
		const fields = this.parseFieldsDefinition();
		return this.node(start, 'InterfaceTypeDefinition', [
			child('description', description),
			child('name', name),
			field('interfaces', interfaces),
			field('directives', directives),
			field('fields', fields),
		]);
	}

	// parseUnionMemberTypes parses
	//
	//	UnionMemberTypes :
	//	  - = `|`? NamedType
	//	  - UnionMemberTypes | NamedType
	parseUnionMemberTypes(): Value {
		if (this.expectOptionalToken('=')) {
			return this.delimitedNamedTypes('|');
		}
		return undefinedValue;
	}

	// parseUnionTypeDefinition parses
	//
	//	UnionTypeDefinition :
	//	  - Description? union Name Directives[Const]? UnionMemberTypes?
	parseUnionTypeDefinition(): number {
		const start = this.lexer.token;
		const description = this.parseDescription();
		this.expectKeyword('union');
		const name = this.parseName();
		const directives = this.parseConstDirectives();
		const types = this.parseUnionMemberTypes();
		return this.node(start, 'UnionTypeDefinition', [child('description', description), child('name', name), field('directives', directives), field('types', types)]);
	}

	// parseEnumValueName parses EnumValue : Name but not `true`, `false` or `null`
	parseEnumValueName(): number {
		switch (this.lexer.token.value) {
			case 'true':
			case 'false':
			case 'null':
				throw new Error(syntaxError(this.lexer.body, this.lexer.token.start, `${getTokenDesc(this.lexer.token)} is reserved and cannot be used for an enum value.`));
		}
		return this.parseName();
	}

	// parseEnumValueDefinition parses EnumValueDefinition : Description? EnumValue Directives[Const]?
	parseEnumValueDefinition(): number {
		const start = this.lexer.token;
		const description = this.parseDescription();
		const name = this.parseEnumValueName();
		const directives = this.parseConstDirectives();
		return this.node(start, 'EnumValueDefinition', [child('description', description), child('name', name), field('directives', directives)]);
	}

	// parseEnumValuesDefinition parses
	//
	//	EnumValuesDefinition : { EnumValueDefinition+ }
	parseEnumValuesDefinition(): Value {
		// optionalMany('{', parseEnumValueDefinition, '}')
		if (!this.expectOptionalToken('{')) {
			return undefinedValue;
		}
		const nodes: number[] = [];
		do {
			nodes.push(this.parseEnumValueDefinition());
		} while (!this.expectOptionalToken('}'));
		return { kind: 'List', nodes };
	}

	// parseEnumTypeDefinition parses
	//
	//	EnumTypeDefinition :
	//	  - Description? enum Name Directives[Const]? EnumValuesDefinition?
	parseEnumTypeDefinition(): number {
		const start = this.lexer.token;
		const description = this.parseDescription();
		this.expectKeyword('enum');
		const name = this.parseName();
		const directives = this.parseConstDirectives();
		const values = this.parseEnumValuesDefinition();
		return this.node(start, 'EnumTypeDefinition', [child('description', description), child('name', name), field('directives', directives), field('values', values)]);
	}

	// parseInputObjectTypeDefinition parses
	//
	//	InputObjectTypeDefinition :
	//	  - Description? input Name Directives[Const]? InputFieldsDefinition?
	parseInputObjectTypeDefinition(): number {
		const start = this.lexer.token;
		const description = this.parseDescription();
		this.expectKeyword('input');
		const name = this.parseName();
		const directives = this.parseConstDirectives();
		const fields = this.parseInputValueDefs('{', '}');
		return this.node(start, 'InputObjectTypeDefinition', [child('description', description), child('name', name), field('directives', directives), field('fields', fields)]);
	}

	// parseDirectiveLocation parses
	//
	//	DirectiveLocation :
	//	  - ExecutableDirectiveLocation
	//	  - TypeSystemDirectiveLocation
	//
	// one of language/directiveLocation.js's names (isDirectiveLocation). The Go reads the Name node's
	// value; the token it was made from is start.
	parseDirectiveLocation(): number {
		const start = this.lexer.token;
		const name = this.parseName();
		if (isDirectiveLocation(start.value ?? '')) {
			return name;
		}
		throw new Error(this.unexpected(start));
	}

	// parseDirectiveDefinition parses
	//
	//	DirectiveDefinition :
	//	  - Description? directive @ Name ArgumentsDefinition? Directives[Const]? `repeatable`? on DirectiveLocations
	//
	// with parseDirectiveLocations, delimitedMany('|', parseDirectiveLocation), written in it.
	parseDirectiveDefinition(): number {
		const start = this.lexer.token;
		const description = this.parseDescription();
		this.expectKeyword('directive');
		this.expectToken('@');
		const name = this.parseName();
		const directiveArguments = this.parseInputValueDefs('(', ')');
		const directives = this.parseConstDirectives();
		const repeatable = this.expectOptionalKeyword('repeatable');
		this.expectKeyword('on');
		this.expectOptionalToken('|');
		const locations: number[] = [];
		do {
			locations.push(this.parseDirectiveLocation());
		} while (this.expectOptionalToken('|'));
		return this.node(start, 'DirectiveDefinition', [
			child('description', description),
			child('name', name),
			field('arguments', directiveArguments),
			field('directives', directives),
			flag('repeatable', repeatable),
			list('locations', locations),
		]);
	}

	// parseSchemaExtension parses
	//
	//	SchemaExtension :
	//	  - extend schema Directives[Const]? { OperationTypeDefinition+ }
	//	  - extend schema Directives[Const]
	parseSchemaExtension(): number {
		const start = this.lexer.token;
		this.expectKeyword('extend');
		this.expectKeyword('schema');
		const directives = this.parseConstDirectives();
		const operationTypes = this.parseOperationTypeDefinitions(false);
		if (directives.kind === 'Undefined' && operationTypes.kind === 'Undefined') {
			throw new Error(this.unexpected(undefined));
		}
		return this.node(start, 'SchemaExtension', [field('directives', directives), field('operationTypes', operationTypes)]);
	}

	// parseScalarTypeExtension parses
	//
	//	ScalarTypeExtension :
	//	  - extend scalar Name Directives[Const]
	parseScalarTypeExtension(): number {
		const start = this.lexer.token;
		this.expectKeyword('extend');
		this.expectKeyword('scalar');
		const name = this.parseName();
		const directives = this.parseConstDirectives();
		if (directives.kind === 'Undefined') {
			throw new Error(this.unexpected(undefined));
		}
		return this.node(start, 'ScalarTypeExtension', [child('name', name), field('directives', directives)]);
	}

	// parseObjectTypeExtension parses
	//
	//	ObjectTypeExtension :
	//	 - extend type Name ImplementsInterfaces? Directives[Const]? FieldsDefinition
	//	 - extend type Name ImplementsInterfaces? Directives[Const]
	//	 - extend type Name ImplementsInterfaces
	parseObjectTypeExtension(): number {
		const start = this.lexer.token;
		this.expectKeyword('extend');
		this.expectKeyword('type');
		const name = this.parseName();
		const interfaces = this.parseImplementsInterfaces();
		const directives = this.parseConstDirectives();
		const fields = this.parseFieldsDefinition();
		if (interfaces.kind === 'Undefined' && directives.kind === 'Undefined' && fields.kind === 'Undefined') {
			throw new Error(this.unexpected(undefined));
		}
		return this.node(start, 'ObjectTypeExtension', [child('name', name), field('interfaces', interfaces), field('directives', directives), field('fields', fields)]);
	}

	// parseInterfaceTypeExtension parses
	//
	//	InterfaceTypeExtension :
	//	 - extend interface Name ImplementsInterfaces? Directives[Const]? FieldsDefinition
	//	 - extend interface Name ImplementsInterfaces? Directives[Const]
	//	 - extend interface Name ImplementsInterfaces
	parseInterfaceTypeExtension(): number {
		const start = this.lexer.token;
		this.expectKeyword('extend');
		this.expectKeyword('interface');
		const name = this.parseName();
		const interfaces = this.parseImplementsInterfaces();
		const directives = this.parseConstDirectives();
		const fields = this.parseFieldsDefinition();
		if (interfaces.kind === 'Undefined' && directives.kind === 'Undefined' && fields.kind === 'Undefined') {
			throw new Error(this.unexpected(undefined));
		}
		return this.node(start, 'InterfaceTypeExtension', [child('name', name), field('interfaces', interfaces), field('directives', directives), field('fields', fields)]);
	}

	// parseUnionTypeExtension parses
	//
	//	UnionTypeExtension :
	//	  - extend union Name Directives[Const]? UnionMemberTypes
	//	  - extend union Name Directives[Const]
	parseUnionTypeExtension(): number {
		const start = this.lexer.token;
		this.expectKeyword('extend');
		this.expectKeyword('union');
		const name = this.parseName();
		const directives = this.parseConstDirectives();
		const types = this.parseUnionMemberTypes();
		if (directives.kind === 'Undefined' && types.kind === 'Undefined') {
			throw new Error(this.unexpected(undefined));
		}
		return this.node(start, 'UnionTypeExtension', [child('name', name), field('directives', directives), field('types', types)]);
	}

	// parseEnumTypeExtension parses
	//
	//	EnumTypeExtension :
	//	  - extend enum Name Directives[Const]? EnumValuesDefinition
	//	  - extend enum Name Directives[Const]
	parseEnumTypeExtension(): number {
		const start = this.lexer.token;
		this.expectKeyword('extend');
		this.expectKeyword('enum');
		const name = this.parseName();
		const directives = this.parseConstDirectives();
		const values = this.parseEnumValuesDefinition();
		if (directives.kind === 'Undefined' && values.kind === 'Undefined') {
			throw new Error(this.unexpected(undefined));
		}
		return this.node(start, 'EnumTypeExtension', [child('name', name), field('directives', directives), field('values', values)]);
	}

	// parseInputObjectTypeExtension parses
	//
	//	InputObjectTypeExtension :
	//	  - extend input Name Directives[Const]? InputFieldsDefinition
	//	  - extend input Name Directives[Const]
	parseInputObjectTypeExtension(): number {
		const start = this.lexer.token;
		this.expectKeyword('extend');
		this.expectKeyword('input');
		const name = this.parseName();
		const directives = this.parseConstDirectives();
		const fields = this.parseInputValueDefs('{', '}');
		if (directives.kind === 'Undefined' && fields.kind === 'Undefined') {
			throw new Error(this.unexpected(undefined));
		}
		return this.node(start, 'InputObjectTypeExtension', [child('name', name), field('directives', directives), field('fields', fields)]);
	}

	parseDirectiveExtension(): number {
		const start = this.lexer.token;
		this.expectKeyword('extend');
		this.expectKeyword('directive');
		this.expectToken('@');
		const name = this.parseName();
		const directives = this.parseConstDirectives();
		if (directives.kind === 'Undefined') {
			throw new Error(this.unexpected(undefined));
		}
		return this.node(start, 'DirectiveExtension', [child('name', name), field('directives', directives)]);
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
	parseTypeSystemExtension(): number {
		const keywordToken = this.lexer.lookahead();

		if (keywordToken.kind === 'Name') {
			switch (keywordToken.value) {
				case 'schema':
					return this.parseSchemaExtension();
				case 'scalar':
					return this.parseScalarTypeExtension();
				case 'type':
					return this.parseObjectTypeExtension();
				case 'interface':
					return this.parseInterfaceTypeExtension();
				case 'union':
					return this.parseUnionTypeExtension();
				case 'enum':
					return this.parseEnumTypeExtension();
				case 'input':
					return this.parseInputObjectTypeExtension();
				case 'directive':
					return this.parseDirectiveExtension();
			}
		}

		throw new Error(this.unexpected(keywordToken));
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
	parseDefinition(): number {
		if (this.peek('{')) {
			return this.parseOperationDefinition();
		}

		// Many definitions begin with a description and require a lookahead.
		const hasDescription = this.peekDescription();
		const keywordToken = hasDescription ? this.lexer.lookahead() : this.lexer.token;

		if (hasDescription && keywordToken.kind === '{') {
			throw new Error(syntaxError(this.lexer.body, this.lexer.token.start, 'Unexpected description, descriptions are not supported on shorthand queries.'));
		}

		if (keywordToken.kind === 'Name') {
			switch (keywordToken.value) {
				case 'schema':
					return this.parseSchemaDefinition();
				case 'scalar':
					return this.parseScalarTypeDefinition();
				case 'type':
					return this.parseObjectTypeDefinition();
				case 'interface':
					return this.parseInterfaceTypeDefinition();
				case 'union':
					return this.parseUnionTypeDefinition();
				case 'enum':
					return this.parseEnumTypeDefinition();
				case 'input':
					return this.parseInputObjectTypeDefinition();
				case 'directive':
					return this.parseDirectiveDefinition();
			}

			switch (keywordToken.value) {
				case 'query':
				case 'mutation':
				case 'subscription':
					return this.parseOperationDefinition();
				case 'fragment':
					return this.parseFragmentDefinition();
			}

			if (hasDescription) {
				throw new Error(syntaxError(this.lexer.body, this.lexer.token.start, 'Unexpected description, only GraphQL definitions support descriptions.'));
			}

			if (keywordToken.value === 'extend') {
				return this.parseTypeSystemExtension();
			}
		}

		throw new Error(this.unexpected(keywordToken));
	}

	// Implements the parsing rules in the Document section.

	// parseDocument parses Document : Definition+
	parseDocument(): number {
		const start = this.lexer.token;
		// many('<SOF>', parseDefinition, '<EOF>')
		this.expectToken('<SOF>');
		const definitions: number[] = [];
		do {
			definitions.push(this.parseDefinition());
		} while (!this.expectOptionalToken('<EOF>'));
		return this.node(start, 'Document', [list('definitions', definitions)]);
	}
}

// parseComments is parser-graphql.js's: the Comment tokens between the document's startToken and
// endToken (the <SOF> and the <EOF>), which lookahead read into the token list as it skipped them.
function parseComments(lexer: Lexer, startIndex: number, endIndex: number): Comment[] {
	const comments: Comment[] = [];
	for (let index = startIndex; index < endIndex; index++) {
		const token = lexer.tokens[index] ?? panic(`no token ${index}`);
		if (token.kind === 'Comment') {
			comments.push({ start: token.start, end: token.end, value: token.value ?? '' });
		}
	}
	return comments;
}

// ParseResult is a document, or the message of the syntax error that refused the text.
export type ParseResult = { readonly kind: 'Parsed'; readonly document: Document } | { readonly kind: 'Refused'; readonly message: string };

// parse is Prettier's GraphQL parser: graphql-js's parse(text, { experimentalFragmentArguments: true }),
// then parseComments, which collects the Comment tokens from the token stream, since graphql-js's AST
// has no comments. graphql-js throws a GraphQLError where parsing fails, and the Go panics; this catches
// what the lexer and parser throw, as the Go's Parse recovers it.
export function parse(source: string): ParseResult {
	try {
		const parser = new Parser(source);
		const startIndex = parser.lexer.tokenIndex;
		const root = parser.parseDocument();
		const endIndex = parser.lexer.lastTokenIndex;
		return { kind: 'Parsed', document: { nodes: parser.nodes, root, comments: parseComments(parser.lexer, startIndex, endIndex) } };
	} catch (error) {
		return { kind: 'Refused', message: error instanceof Error ? error.message : 'not an Error' };
	}
}
