// Expression descent follows cohere/TypeScript/tsc/internal/parser/parser.go.
import { panic } from 'adamic';
import { Scanner } from '../scanner/scanner.ts';
import { Statements } from './statements.ts';
import { Jsx } from './jsx.ts';
import { ParseNode } from './nodes.ts';
import { precedence, reservedKinds } from './grammar.ts';
import { arrowAhead, typeArgumentsAhead, jsxArrowAhead } from './lookahead.ts';
import type { ParserStateInterface } from './lookahead.ts';

export class Parser {
    readonly scanner: Scanner;
    readonly path: string;
    readonly jsx: boolean;
    readonly javascript: boolean;
    readonly nodes: ParseNode[] = [];
    readonly roots: number[] = [];
    recoveredJsx = false;
    constructor(text: string, path = 'source') {
        this.path = path;
        this.javascript =
            path.endsWith('.js') || path.endsWith('.jsx') || path.endsWith('.mjs') || path.endsWith('.cjs');
        this.jsx = this.javascript || path.endsWith('.tsx');
        this.scanner = new Scanner(text);
        this.next();
    }
    kind(): string {
        return this.scanner.kind;
    }
    next(): void {
        this.scanner.scan();
    }
    node(index: number): ParseNode {
        return this.nodes[index] ?? panic('missing node');
    }
    make(kind: string, pos: number, children: number[] = []): number {
        this.nodes.push(new ParseNode(kind, pos, this.scanner.fullStart, children));
        return this.nodes.length - 1;
    }
    token(): number {
        const id = this.make(this.kind(), this.scanner.fullStart);
        this.next();
        this.node(id).end = this.scanner.fullStart;
        return id;
    }
    expect(kind: string): void {
        if(this.kind() !== kind) {
            panic(`parser slice expected ${kind}, got ${this.kind()} at ${this.scanner.start} in ${this.path}`);
        }
        this.next();
    }
    literal(): number {
        const id = this.make(this.kind(), this.scanner.fullStart);
        const node = this.node(id);
        node.text = this.scanner.value;
        if(this.kind() === 'NoSubstitutionTemplateLiteral') {
            node.raw = this.scanner.text.slice(this.scanner.start + 1, this.scanner.pos - 1);
        }
        node.literalFlags =
            this.scanner.flags &
            (this.kind() === 'StringLiteral'
                ? 72716
                : this.kind() === 'RegularExpressionLiteral'
                  ? 4
                  : this.kind() === 'NoSubstitutionTemplateLiteral'
                    ? 7180
                    : 25584);
        this.next();
        node.end = this.scanner.fullStart;
        return id;
    }
    identifier(): number {
        const pos = this.scanner.fullStart;
        const text = this.scanner.value;
        this.next();
        const id = this.make('Identifier', pos);
        this.node(id).text = text;
        return id;
    }
    primary(): number {
        const pos = this.scanner.fullStart;
        switch(this.kind()) {
            case 'EndOfFile':
                if(this.recoveredJsx) {
                    // Go inserts a zero-width operand after a recovered JSX delimiter.
                    return this.make('Identifier', pos);
                }
                return panic(`parser slice missing expression at ${this.scanner.start} in ${this.path}`);
            case 'NumericLiteral':
            case 'BigIntLiteral':
            case 'StringLiteral':
            case 'NoSubstitutionTemplateLiteral':
                return this.literal();
            case 'SlashToken':
            case 'SlashEqualsToken':
                this.scanner.rescanSlash();
                return this.literal();
            case 'ThisKeyword':
            case 'SuperKeyword':
            case 'NullKeyword':
            case 'TrueKeyword':
            case 'FalseKeyword':
                return this.token();
            case 'PrivateIdentifier': {
                const id = this.identifier();
                const old = this.node(id);
                this.nodes[id] = new ParseNode('PrivateIdentifier', old.pos, old.end, []);
                this.node(id).text = old.text;
                return id;
            }
            case 'OpenBracketToken':
                return this.array();
            case 'OpenBraceToken':
                return this.object();
            case 'TemplateHead':
                return this.template();
            case 'NewKeyword': {
                this.next();
                if(this.kind() === 'DotToken') {
                    this.next();
                    const name = this.identifier();
                    const id = this.make('MetaProperty', pos, [name]);
                    this.node(id).operator = 'NewKeyword';
                    return id;
                }
                const target = this.suffix(this.primary(), false);
                const children =
                    this.node(target).kind === 'ExpressionWithTypeArguments'
                        ? this.node(target).children.slice()
                        : [target];
                let list = -1;
                let trailing = false;
                if(this.kind() === 'OpenParenToken') {
                    const args = this.arguments();
                    children.push(...args);
                    list = args.length;
                    trailing = this.lastTrailing;
                }
                const id = this.make('NewExpression', pos, children);
                this.node(id).list = list;
                this.node(id).trailing = trailing;
                return id;
            }
            case 'ImportKeyword': {
                if(this.peek() === 'DotToken') {
                    this.next();
                    this.next();
                    const name = this.identifier();
                    const id = this.make('MetaProperty', pos, [name]);
                    this.node(id).operator = 'ImportKeyword';
                    return id;
                }
                return this.token();
            }
            case 'ClassKeyword':
                return this.classDeclaration(pos, [], true);
            case 'FunctionKeyword':
                return this.functionExpression();
            case 'AsyncKeyword':
                if(this.peek() === 'FunctionKeyword') {
                    return this.functionExpression();
                }
                return this.identifier();
            case 'OpenParenToken': {
                this.next();
                const expression = this.allowInExpression();
                this.expect('CloseParenToken');
                return this.make('ParenthesizedExpression', pos, [expression]);
            }
            default:
                if(this.kind() === 'Identifier' || this.kind().endsWith('Keyword')) {
                    return this.identifier();
                }
                panic(`parser slice unsupported primary ${this.kind()} at ${this.scanner.start} in ${this.path}`);
        }
    }
    jsxParser(): Jsx {
        return new Jsx({
            scanner: this.scanner,
            path: this.path,
            javascript: this.javascript,
            kind: () => this.kind(),
            next: () => {
                this.next();
            },
            expect: (kind) => {
                this.expect(kind);
            },
            node: (index) => this.node(index),
            make: (kind, pos, children) => this.make(kind, pos, children),
            identifier: () => this.identifier(),
            token: () => this.token(),
            literal: () => this.literal(),
            expression: () => this.allowInExpression(),
            typeArguments: () => this.typeArguments(),
            missingGreater: () => {
                this.recoveredJsx = true;
            },
            typeTrailing: () => this.lastTypeTrailing,
        });
    }
    bindingIdentifier(): boolean {
        return (
            this.kind() === 'Identifier' || (this.kind().endsWith('Keyword') && !reservedKinds.includes(this.kind()))
        );
    }
    nextIdentifierSameLine(): boolean {
        const state = this.mark();
        this.next();
        const result = this.bindingIdentifier() && (this.scanner.flags & 1) === 0;
        this.rewind(state);
        return result;
    }
    // TypeScript recognizes an unmistakable yield operand outside generators,
    // but keeps yield(x) as an ordinary call and never crosses a line break.
    yieldOperandAhead(): boolean {
        const state = this.mark();
        this.next();
        const result =
            (this.scanner.flags & 1) === 0 &&
            (this.kind() === 'Identifier' ||
                this.kind().endsWith('Keyword') ||
                ['NumericLiteral', 'BigIntLiteral', 'StringLiteral'].includes(this.kind()));
        this.rewind(state);
        return result;
    }
    mark(): ParserStateInterface {
        return {
            pos: this.scanner.pos,
            start: this.scanner.start,
            fullStart: this.scanner.fullStart,
            kind: this.kind(),
            value: this.scanner.value,
            flags: this.scanner.flags,
            errors: this.scanner.errors.length,
            nodes: this.nodes.length,
        };
    }
    rewind(state: ParserStateInterface): void {
        this.scanner.pos = state.pos;
        this.scanner.start = state.start;
        this.scanner.fullStart = state.fullStart;
        this.scanner.kind = state.kind;
        this.scanner.value = state.value;
        this.scanner.flags = state.flags;
        this.scanner.errors.splice(state.errors);
        this.nodes.splice(state.nodes);
    }
    docTypes(): number[] {
        const types: number[] = [];
        let index = this.scanner.text.indexOf('@type');
        while(index >= 0) {
            const brace = this.scanner.text.indexOf('{', index);
            if(brace < 0) {
                panic('doc type missing brace');
            }
            this.scanner.pos = brace + 1;
            this.next();
            const pos = this.scanner.fullStart;
            const variadic = this.kind() === 'DotDotDotToken';
            if(variadic) {
                this.next();
            }
            let type = this.returnType();
            if(variadic) {
                type = this.make('JSDocVariadicType', pos, [type]);
            }
            if(this.kind() === 'EqualsToken') {
                this.next();
                type = this.make('JSDocOptionalType', pos, [type]);
            }
            this.expect('CloseBraceToken');
            types.push(type);
            index = this.scanner.text.indexOf('@type', this.scanner.pos);
        }
        return types;
    }
    tupleNameAhead(): boolean {
        if(!this.bindingIdentifier()) {
            return false;
        }
        const saved = this.mark();
        this.next();
        if(this.kind() === 'QuestionToken') {
            this.next();
        }
        const result = this.kind() === 'ColonToken';
        this.rewind(saved);
        return result;
    }
    lastTypeTrailing = false;
    typeArguments(): number[] {
        const result: number[] = [];
        let trailing = false;
        this.expect('LessThanToken');
        while(this.kind() !== 'GreaterThanToken') {
            result.push(this.type());
            trailing = this.kind() === 'CommaToken';
            if(this.kind() !== 'CommaToken') {
                break;
            }
            this.next();
        }
        this.expect('GreaterThanToken');
        this.lastTypeTrailing = trailing;
        return result;
    }
    typeParameters(): number[] {
        const result: number[] = [];
        if(this.kind() !== 'LessThanToken') {
            return result;
        }
        this.next();
        while(this.kind() !== 'GreaterThanToken') {
            const pos = this.scanner.fullStart;
            const children: number[] = [];
            while(this.kind() === 'ConstKeyword' || this.kind() === 'InKeyword' || this.kind() === 'OutKeyword') {
                children.push(this.token());
            }
            children.push(this.identifier());
            if(this.kind() === 'ExtendsKeyword') {
                this.next();
                children.push(this.type(0, false));
            }
            if(this.kind() === 'EqualsToken') {
                this.next();
                children.push(this.type());
            }
            result.push(this.make('TypeParameter', pos, children));
            if(this.kind() !== 'CommaToken') {
                break;
            }
            this.next();
        }
        this.expect('GreaterThanToken');
        return result;
    }
    entityName(): number {
        const pos = this.scanner.fullStart;
        let name = this.identifier();
        while(this.kind() === 'DotToken') {
            this.next();
            const right = this.identifier();
            name = this.make('QualifiedName', pos, [name, right]);
        }
        return name;
    }
    importType(): number {
        const pos = this.scanner.fullStart;
        const query = this.kind() === 'TypeOfKeyword';
        if(query) {
            this.next();
        }
        this.expect('ImportKeyword');
        this.expect('OpenParenToken');
        const children = [this.type()];
        if(this.kind() === 'CommaToken') {
            this.next();
            this.expect('OpenBraceToken');
            const operator = this.kind();
            this.next();
            this.expect('ColonToken');
            children.push(this.statements().attributeBody(this.scanner.fullStart, operator));
            if(this.kind() === 'CommaToken') {
                this.next();
            }
            this.expect('CloseBraceToken');
        }
        this.expect('CloseParenToken');
        if(this.kind() === 'DotToken') {
            this.next();
            children.push(this.entityName());
        }
        if(this.kind() === 'LessThanToken') {
            for(const type of this.typeArguments()) {
                children.push(type);
            }
        }
        const id = this.make('ImportType', pos, children);
        if(query) {
            this.node(id).operator = 'TypeOfKeyword';
        }
        return id;
    }
    type(minimum = 0, conditional = true): number {
        const pos = this.scanner.fullStart;
        let left: number;
        if(this.kind() === 'BarToken' || this.kind() === 'AmpersandToken') {
            const union = this.kind() === 'BarToken';
            this.next();
            const types = [this.type(union ? 1 : 2, false)];
            while(this.kind() === (union ? 'BarToken' : 'AmpersandToken')) {
                this.next();
                types.push(this.type(union ? 1 : 2, false));
            }
            left = this.make(union ? 'UnionType' : 'IntersectionType', pos, types);
        }
        else if(
            this.kind() === 'KeyOfKeyword' ||
            this.kind() === 'ReadonlyKeyword' ||
            this.kind() === 'UniqueKeyword'
        ) {
            const operator = this.kind();
            this.next();
            const operand = this.type(2, false);
            left = this.make('TypeOperator', pos, [operand]);
            this.node(left).operator = operator;
        }
        else if(this.kind() === 'AsteriskToken') {
            this.next();
            left = this.make('JSDocAllType', pos);
        }
        else if(this.kind() === 'QuestionToken' || this.kind() === 'ExclamationToken') {
            const nullable = this.kind() === 'QuestionToken';
            this.next();
            left = this.make(nullable ? 'JSDocNullableType' : 'JSDocNonNullableType', pos, [this.type(2, false)]);
        }
        else if(this.kind() === 'MinusToken') {
            this.next();
            const operand = this.literal();
            const expression = this.make('PrefixUnaryExpression', pos, [operand]);
            this.node(expression).operator = 'MinusToken';
            if(this.expressionDepth === 0) {
                this.roots.push(expression);
            }
            left = this.make('LiteralType', pos, [expression]);
        }
        else if(this.kind() === 'InferKeyword') {
            this.next();
            const start = this.scanner.fullStart;
            const children = [this.identifier()];
            if(this.kind() === 'ExtendsKeyword') {
                const state = this.mark();
                const roots = this.roots.length;
                this.next();
                const constraint = this.type(0, false);
                if(conditional && this.kind() === 'QuestionToken') {
                    this.rewind(state);
                    this.roots.splice(roots);
                }
                else {
                    children.push(constraint);
                }
            }
            const parameter = this.make('TypeParameter', start, children);
            left = this.make('InferType', pos, [parameter]);
        }
        else if(
            this.kind() === 'ImportKeyword' ||
            (this.kind() === 'TypeOfKeyword' && this.peek() === 'ImportKeyword')
        ) {
            left = this.importType();
        }
        else if(this.kind() === 'TypeOfKeyword') {
            this.next();
            const name = this.entityName();
            const children = [name];
            if(this.expressionDepth === 0) {
                this.roots.push(name);
            }
            if(this.kind() === 'LessThanToken') {
                for(const type of this.typeArguments()) {
                    children.push(type);
                }
            }
            left = this.make('TypeQuery', pos, children);
        }
        else if(this.kind() === 'NewKeyword' || (this.kind() === 'AbstractKeyword' && this.peek() === 'NewKeyword')) {
            const children: number[] = [];
            if(this.kind() === 'AbstractKeyword') {
                children.push(this.token());
            }
            this.expect('NewKeyword');
            for(const type of this.typeParameters()) {
                children.push(type);
            }
            for(const parameter of this.parameters()) {
                children.push(parameter);
            }
            this.expect('EqualsGreaterThanToken');
            children.push(this.returnType());
            left = this.make('ConstructorType', pos, children);
        }
        else if(this.kind() === 'OpenParenToken' || this.kind() === 'LessThanToken') {
            if(arrowAhead(this.scanner, true)) {
                const children = this.typeParameters();
                for(const parameter of this.parameters()) {
                    children.push(parameter);
                }
                this.expect('EqualsGreaterThanToken');
                children.push(this.returnType());
                left = this.make('FunctionType', pos, children);
            }
            else {
                this.expect('OpenParenToken');
                const type = this.type();
                this.expect('CloseParenToken');
                left = this.make('ParenthesizedType', pos, [type]);
            }
        }
        else if(this.kind() === 'TemplateHead') {
            left = this.templateType();
        }
        else if(this.kind() === 'OpenBraceToken') {
            left = this.mappedAhead() ? this.mappedType() : this.typeLiteral();
        }
        else if(this.kind() === 'OpenBracketToken') {
            this.next();
            const children: number[] = [];
            while(this.kind() !== 'CloseBracketToken') {
                const start = this.scanner.fullStart;
                const rest = this.kind() === 'DotDotDotToken' ? this.token() : -1;
                if(this.tupleNameAhead()) {
                    const parts: number[] = [];
                    if(rest >= 0) {
                        parts.push(rest);
                    }
                    parts.push(this.identifier());
                    if(this.kind() === 'QuestionToken') {
                        parts.push(this.token());
                    }
                    this.expect('ColonToken');
                    parts.push(this.type());
                    children.push(this.make('NamedTupleMember', start, parts));
                }
                else {
                    const child = this.type();
                    if(rest >= 0) {
                        children.push(this.make('RestType', start, [child]));
                    }
                    else if(this.kind() === 'QuestionToken') {
                        this.next();
                        children.push(this.make('OptionalType', start, [child]));
                    }
                    else {
                        children.push(child);
                    }
                }
                if(this.kind() !== 'CommaToken') {
                    break;
                }
                this.next();
            }
            this.expect('CloseBracketToken');
            left = this.make('TupleType', pos, children);
        }
        else if(
            this.kind() === 'StringLiteral' ||
            this.kind() === 'NumericLiteral' ||
            this.kind() === 'BigIntLiteral' ||
            this.kind() === 'NoSubstitutionTemplateLiteral' ||
            this.kind() === 'TrueKeyword' ||
            this.kind() === 'FalseKeyword' ||
            this.kind() === 'NullKeyword'
        ) {
            const literal = this.kind().endsWith('Keyword') ? this.token() : this.literal();
            if(
                this.expressionDepth === 0 &&
                (this.node(literal).kind === 'TrueKeyword' ||
                    this.node(literal).kind === 'FalseKeyword' ||
                    this.node(literal).kind === 'NullKeyword')
            ) {
                this.roots.push(literal);
            }
            left = this.make('LiteralType', pos, [literal]);
        }
        else if(
            this.kind() === 'AnyKeyword' ||
            this.kind() === 'UnknownKeyword' ||
            this.kind() === 'NumberKeyword' ||
            this.kind() === 'StringKeyword' ||
            this.kind() === 'BooleanKeyword' ||
            this.kind() === 'BigIntKeyword' ||
            this.kind() === 'SymbolKeyword' ||
            this.kind() === 'UndefinedKeyword' ||
            this.kind() === 'NeverKeyword' ||
            this.kind() === 'ObjectKeyword' ||
            this.kind() === 'VoidKeyword'
        ) {
            if(this.kind() !== 'VoidKeyword' && this.peek() === 'DotToken') {
                const children = [this.entityName()];
                if(this.kind() === 'LessThanToken') {
                    for(const type of this.typeArguments()) {
                        children.push(type);
                    }
                }
                left = this.make('TypeReference', pos, children);
            }
            else {
                left = this.token();
            }
        }
        else if(this.kind() === 'AssertsKeyword' && this.nextIdentifierSameLine()) {
            left = this.returnType();
        }
        else if(this.kind() === 'ThisKeyword') {
            this.next();
            left = this.make('ThisType', pos);
            if(this.kind() === 'IsKeyword' && (this.scanner.flags & 1) === 0) {
                this.next();
                left = this.make('TypePredicate', pos, [left, this.type()]);
            }
        }
        else {
            const name = this.entityName();
            const children = [name];
            if(this.kind() === 'LessThanToken' && (this.scanner.flags & 1) === 0) {
                for(const type of this.typeArguments()) {
                    children.push(type);
                }
            }
            left = this.make('TypeReference', pos, children);
        }
        while(
            (this.kind() === 'OpenBracketToken' ||
                this.kind() === 'ExclamationToken' ||
                (this.kind() === 'QuestionToken' &&
                    (this.peek() === 'SemicolonToken' ||
                        this.peek() === 'CloseParenToken' ||
                        this.peek() === 'BarToken' ||
                        this.peek() === 'AmpersandToken' ||
                        this.peek() === 'GreaterThanToken' ||
                        this.peek() === 'CloseBraceToken'))) &&
            (this.scanner.flags & 1) === 0
        ) {
            if(this.kind() !== 'OpenBracketToken') {
                const nullable = this.kind() === 'QuestionToken';
                this.next();
                left = this.make(nullable ? 'JSDocNullableType' : 'JSDocNonNullableType', pos, [left]);
                continue;
            }
            this.next();
            const children = [left];
            const array = this.kind() === 'CloseBracketToken';
            if(!array) {
                children.push(this.type());
            }
            this.expect('CloseBracketToken');
            left = this.make(array ? 'ArrayType' : 'IndexedAccessType', pos, children);
        }
        while(this.kind() === 'BarToken' || this.kind() === 'AmpersandToken') {
            const union = this.kind() === 'BarToken';
            const rank = union ? 1 : 2;
            if(rank <= minimum) {
                break;
            }
            const types = [left];
            while(this.kind() === (union ? 'BarToken' : 'AmpersandToken')) {
                this.next();
                types.push(this.type(rank, false));
            }
            left = this.make(union ? 'UnionType' : 'IntersectionType', pos, types);
        }
        if(conditional && minimum === 0 && this.kind() === 'ExtendsKeyword' && (this.scanner.flags & 1) === 0) {
            this.next();
            const constraint = this.type(0, false);
            this.expect('QuestionToken');
            const yes = this.type();
            this.expect('ColonToken');
            const no = this.type();
            left = this.make('ConditionalType', pos, [left, constraint, yes, no]);
        }
        return left;
    }
    mappedAhead(): boolean {
        const state = this.mark();
        this.next();
        if(this.kind() === 'PlusToken' || this.kind() === 'MinusToken') {
            this.next();
        }
        if(this.kind() === 'ReadonlyKeyword') {
            this.next();
        }
        let result = false;
        if(this.kind() === 'OpenBracketToken') {
            this.next();
            this.next();
            result = this.kind() === 'InKeyword';
        }
        this.rewind(state);
        return result;
    }
    mappedType(): number {
        const pos = this.scanner.fullStart;
        this.expect('OpenBraceToken');
        const children: number[] = [];
        if(this.kind() === 'PlusToken' || this.kind() === 'MinusToken') {
            children.push(this.token());
            this.expect('ReadonlyKeyword');
        }
        else if(this.kind() === 'ReadonlyKeyword') {
            children.push(this.token());
        }
        this.expect('OpenBracketToken');
        const start = this.scanner.fullStart;
        const name = this.identifier();
        this.expect('InKeyword');
        const constraint = this.type();
        children.push(this.make('TypeParameter', start, [name, constraint]));
        if(this.kind() === 'AsKeyword') {
            this.next();
            children.push(this.type());
        }
        this.expect('CloseBracketToken');
        if(this.kind() === 'PlusToken' || this.kind() === 'MinusToken') {
            children.push(this.token());
            this.expect('QuestionToken');
        }
        else if(this.kind() === 'QuestionToken') {
            children.push(this.token());
        }
        if(this.kind() === 'ColonToken') {
            this.next();
            children.push(this.type());
        }
        if(this.kind() === 'SemicolonToken') {
            this.next();
        }
        this.expect('CloseBraceToken');
        return this.make('MappedType', pos, children);
    }
    indexSignatureAhead(): boolean {
        const state = this.mark();
        this.next();
        this.next();
        const result = this.kind() === 'ColonToken';
        this.rewind(state);
        return result;
    }
    typeLiteral(): number {
        const pos = this.scanner.fullStart;
        this.expect('OpenBraceToken');
        const members: number[] = [];
        while(this.kind() !== 'CloseBraceToken') {
            const start = this.scanner.fullStart;
            const children: number[] = [];
            if(
                this.kind() === 'ReadonlyKeyword' &&
                (this.peek() === 'Identifier' ||
                    this.peek().endsWith('Keyword') ||
                    this.peek() === 'OpenBracketToken' ||
                    this.peek() === 'StringLiteral' ||
                    this.peek() === 'NumericLiteral')
            ) {
                children.push(this.token());
            }
            let kind = 'PropertySignature';
            if(this.kind() === 'NewKeyword' || this.kind() === 'OpenParenToken' || this.kind() === 'LessThanToken') {
                const constructor = this.kind() === 'NewKeyword';
                if(constructor) {
                    this.next();
                }
                for(const type of this.typeParameters()) {
                    children.push(type);
                }
                for(const parameter of this.parameters()) {
                    children.push(parameter);
                }
                kind = constructor ? 'ConstructSignature' : 'CallSignature';
            }
            else if(this.kind() === 'OpenBracketToken' && this.indexSignatureAhead()) {
                this.next();
                const parameterPos = this.scanner.fullStart;
                const parameter = [this.identifier()];
                this.expect('ColonToken');
                parameter.push(this.type());
                children.push(this.make('Parameter', parameterPos, parameter));
                this.expect('CloseBracketToken');
                kind = 'IndexSignature';
            }
            else {
                if(
                    (this.kind() === 'GetKeyword' || this.kind() === 'SetKeyword') &&
                    this.peek() !== 'OpenParenToken' &&
                    this.peek() !== 'LessThanToken' &&
                    this.peek() !== 'ColonToken' &&
                    this.peek() !== 'QuestionToken'
                ) {
                    kind = this.kind() === 'GetKeyword' ? 'GetAccessor' : 'SetAccessor';
                    this.next();
                }
                children.push(this.propertyName());
                if(this.kind() === 'QuestionToken') {
                    children.push(this.token());
                }
                if(this.kind() === 'OpenParenToken' || this.kind() === 'LessThanToken') {
                    if(kind === 'PropertySignature') {
                        kind = 'MethodSignature';
                    }
                    for(const type of this.typeParameters()) {
                        children.push(type);
                    }
                    for(const parameter of this.parameters()) {
                        children.push(parameter);
                    }
                }
            }
            if(this.kind() === 'ColonToken') {
                this.next();
                children.push(this.returnType());
            }
            if(this.kind() === 'SemicolonToken' || this.kind() === 'CommaToken') {
                this.next();
            }
            members.push(this.make(kind, start, children));
        }
        this.expect('CloseBraceToken');
        return this.make('TypeLiteral', pos, members);
    }
    parameters(): number[] {
        this.expect('OpenParenToken');
        const result: number[] = [];
        while(this.kind() !== 'CloseParenToken') {
            const pos = this.scanner.fullStart;
            const children: number[] = [];
            while(this.kind() === 'AtToken') {
                children.push(this.statements().decorator());
            }
            while(
                this.kind() === 'PublicKeyword' ||
                this.kind() === 'PrivateKeyword' ||
                this.kind() === 'ProtectedKeyword' ||
                this.kind() === 'ReadonlyKeyword' ||
                this.kind() === 'OverrideKeyword'
            ) {
                if(!this.nextIdentifierSameLine()) {
                    break;
                }
                children.push(this.token());
            }
            if(this.kind() === 'DotDotDotToken') {
                children.push(this.token());
            }
            children.push(this.bindingName());
            if(this.kind() === 'QuestionToken') {
                children.push(this.token());
            }
            if(this.kind() === 'ColonToken') {
                this.next();
                children.push(this.type());
            }
            if(this.kind() === 'EqualsToken') {
                this.next();
                children.push(this.rootAssignment());
            }
            result.push(this.make('Parameter', pos, children));
            if(this.kind() !== 'CommaToken') {
                break;
            }
            this.next();
        }
        this.expect('CloseParenToken');
        return result;
    }
    bindingName(): number {
        if(this.kind() !== 'OpenBracketToken' && this.kind() !== 'OpenBraceToken') {
            return this.identifier();
        }
        const pos = this.scanner.fullStart;
        const object = this.kind() === 'OpenBraceToken';
        this.next();
        const children: number[] = [];
        while(this.kind() !== (object ? 'CloseBraceToken' : 'CloseBracketToken')) {
            const start = this.scanner.fullStart;
            if(this.kind() === 'CommaToken') {
                children.push(this.make('BindingElement', start));
                this.next();
                continue;
            }
            const element: number[] = [];
            if(this.kind() === 'DotDotDotToken') {
                element.push(this.token());
            }
            let name = object ? this.propertyName() : this.bindingName();
            if(object && this.kind() === 'ColonToken') {
                element.push(name);
                this.next();
                name = this.bindingName();
            }
            element.push(name);
            if(this.kind() === 'EqualsToken') {
                this.next();
                element.push(this.rootAssignment());
            }
            children.push(this.make('BindingElement', start, element));
            if(this.kind() !== 'CommaToken') {
                break;
            }
            this.next();
        }
        this.expect(object ? 'CloseBraceToken' : 'CloseBracketToken');
        return this.make(object ? 'ObjectBindingPattern' : 'ArrayBindingPattern', pos, children);
    }
    decoratorContext = false;
    decorator(): number {
        const pos = this.scanner.fullStart;
        this.expect('AtToken');
        const saved = this.decoratorContext;
        this.decoratorContext = true;
        const expression = this.suffix(this.primary(), true);
        this.decoratorContext = saved;
        return this.make('Decorator', pos, [expression]);
    }
    awaitContext = false;
    yieldContext = false;
    expressionDepth = 0;
    rootAssignment(): number {
        this.expressionDepth++;
        const result = this.assignment();
        this.expressionDepth--;
        if(this.expressionDepth === 0) {
            this.roots.push(result);
        }
        return result;
    }
    rootExpression(): number {
        this.expressionDepth++;
        const result = this.expression();
        this.expressionDepth--;
        if(this.expressionDepth === 0) {
            this.roots.push(result);
        }
        return result;
    }
    arrow(allowReturn: boolean): number {
        const pos = this.scanner.fullStart;
        const children: number[] = [];
        let async = false;
        if(this.kind() === 'AsyncKeyword') {
            children.push(this.token());
            async = true;
        }
        const previousAwait = this.awaitContext;
        const previousYield = this.yieldContext;
        this.awaitContext = async;
        this.yieldContext = false;
        const oldIn = this.disallowIn;
        this.disallowIn = false;
        for(const parameter of this.typeParameters()) {
            children.push(parameter);
        }
        if(this.kind() === 'OpenParenToken') {
            for(const parameter of this.parameters()) {
                children.push(parameter);
            }
        }
        else {
            const name = this.identifier();
            children.push(this.make('Parameter', this.node(name).pos, [name]));
        }
        if(this.kind() === 'ColonToken') {
            this.next();
            children.push(this.returnType());
        }
        children.push(this.token());
        children.push(this.kind() === 'OpenBraceToken' ? this.block() : this.assignment(allowReturn));
        this.disallowIn = oldIn;
        this.awaitContext = previousAwait;
        this.yieldContext = previousYield;
        return this.make('ArrowFunction', pos, children);
    }
    returnType(): number {
        const pos = this.scanner.fullStart;
        const children: number[] = [];
        if(this.kind() === 'AssertsKeyword') {
            children.push(this.token());
        }
        if((this.kind() === 'Identifier' || this.kind().endsWith('Keyword')) && this.peek() === 'IsKeyword') {
            if(this.kind() === 'ThisKeyword') {
                this.next();
                children.push(this.make('ThisType', pos));
            }
            else {
                children.push(this.identifier());
            }
            this.next();
            children.push(this.type());
            return this.make('TypePredicate', pos, children);
        }
        if(children.length > 0) {
            children.push(this.identifier());
            return this.make('TypePredicate', pos, children);
        }
        return this.type();
    }
    functionExpression(): number {
        const pos = this.scanner.fullStart;
        const children: number[] = [];
        let async = false;
        if(this.kind() === 'AsyncKeyword') {
            async = true;
            children.push(this.token());
        }
        this.expect('FunctionKeyword');
        const generator = this.kind() === 'AsteriskToken';
        if(generator) {
            children.push(this.token());
        }
        if(this.kind() === 'Identifier' || this.kind().endsWith('Keyword')) {
            children.push(this.identifier());
        }
        const oldAwait = this.awaitContext;
        const oldYield = this.yieldContext;
        this.awaitContext = async;
        this.yieldContext = generator;
        const oldIn = this.disallowIn;
        this.disallowIn = false;
        for(const parameter of this.typeParameters()) {
            children.push(parameter);
        }
        for(const parameter of this.parameters()) {
            children.push(parameter);
        }
        if(this.kind() === 'ColonToken') {
            this.next();
            children.push(this.returnType());
        }
        children.push(this.block());
        this.disallowIn = oldIn;
        this.awaitContext = oldAwait;
        this.yieldContext = oldYield;
        return this.make('FunctionExpression', pos, children);
    }
    disallowIn = false;
    statements(): Statements {
        return new Statements({
            scanner: this.scanner,
            path: this.path,
            roots: this.roots,
            kind: () => this.kind(),
            next: () => {
                this.next();
            },
            expect: (kind) => {
                this.expect(kind);
            },
            node: (index) => this.node(index),
            make: (kind, pos, children: number[] = []) => this.make(kind, pos, children),
            entityName: () => this.entityName(),
            identifier: () => this.identifier(),
            token: () => this.token(),
            type: (minimum = 0, conditional = true) => this.type(minimum, conditional),
            typeArguments: () => this.typeArguments(),
            typeParameters: () => this.typeParameters(),
            parameters: () => this.parameters(),
            returnType: () => this.returnType(),
            bindingName: () => this.bindingName(),
            rootAssignment: () => this.rootAssignment(),
            rootExpression: () => this.rootExpression(),
            peek: () => this.peek(),
            nextIdentifierSameLine: () => this.nextIdentifierSameLine(),
            propertyName: () => this.propertyName(),
            methodBody: (pos, prefix, kind, async, generator) => this.methodBody(pos, prefix, kind, async, generator),
            primary: () => this.primary(),
            suffix: (expression, call) => this.suffix(expression, call),
            typeLiteral: () => this.typeLiteral(),
            decorator: () => this.decorator(),
            getAwait: () => this.awaitContext,
            setAwait: (value) => {
                this.awaitContext = value;
            },
            getYield: () => this.yieldContext,
            setYield: (value) => {
                this.yieldContext = value;
            },
            getIn: () => this.disallowIn,
            setIn: (value) => {
                this.disallowIn = value;
            },
            depth: () => this.expressionDepth,
        });
    }
    block(): number {
        return this.statements().block();
    }
    statement(): number {
        return this.statements().statement();
    }
    semicolon(): void {
        this.statements().semicolon();
    }
    classDeclaration(pos: number, prefix: readonly number[], expression: boolean): number {
        return this.statements().classDeclaration(pos, prefix, expression);
    }
    lastTrailing = false;
    peek(): string {
        const pos = this.scanner.pos;
        const start = this.scanner.start;
        const fullStart = this.scanner.fullStart;
        const value = this.scanner.value;
        const kind = this.kind();
        const flags = this.scanner.flags;
        const errors = this.scanner.errors.length;
        this.next();
        const result = this.kind();
        this.scanner.pos = pos;
        this.scanner.start = start;
        this.scanner.fullStart = fullStart;
        this.scanner.value = value;
        this.scanner.kind = kind;
        this.scanner.flags = flags;
        this.scanner.errors.splice(errors);
        return result;
    }
    optionalChain(index: number): boolean {
        const node = this.node(index);
        if(node.optional) {
            return true;
        }
        if(node.kind !== 'NonNullExpression') {
            return false;
        }
        const child = node.children[0] ?? panic('non-null expression without child');
        if(this.optionalChain(child)) {
            node.optional = true;
            return true;
        }
        return false;
    }
    spread(): number {
        const pos = this.scanner.fullStart;
        this.next();
        const expression = this.assignment();
        return this.make('SpreadElement', pos, [expression]);
    }
    arguments(): number[] {
        const oldIn = this.disallowIn;
        this.disallowIn = false;
        this.expect('OpenParenToken');
        const result: number[] = [];
        let trailing = false;
        while(this.kind() !== 'CloseParenToken') {
            result.push(this.kind() === 'DotDotDotToken' ? this.spread() : this.allowInAssignment());
            trailing = this.kind() === 'CommaToken';
            if(!trailing) {
                break;
            }
            this.next();
        }
        this.expect('CloseParenToken');
        this.lastTrailing = trailing;
        this.disallowIn = oldIn;
        return result;
    }
    suffix(expression: number, allowCall: boolean): number {
        let left = expression;
        const pos = this.node(left).pos;
        while(true) {
            if(this.kind() === 'LessThanToken' && typeArgumentsAhead(this.scanner)) {
                const types = this.typeArguments();
                const children = [left];
                for(const type of types) {
                    children.push(type);
                }
                left = this.make('ExpressionWithTypeArguments', pos, children);
                continue;
            }
            let question = -1;
            if(this.kind() === 'QuestionDotToken') {
                const after = this.peek();
                if(!allowCall && after === 'OpenParenToken') {
                    break;
                }
                question = this.token();
            }
            const optionalTypes: number[] = [];
            if(question >= 0 && this.kind() === 'LessThanToken') {
                for(const type of this.typeArguments()) {
                    optionalTypes.push(type);
                }
            }
            if(
                this.kind() === 'DotToken' ||
                (question >= 0 &&
                    (this.kind() === 'Identifier' ||
                        this.kind().endsWith('Keyword') ||
                        this.kind() === 'PrivateIdentifier'))
            ) {
                if(question < 0) {
                    this.next();
                }
                const children = [left];
                if(question >= 0) {
                    children.push(question);
                }
                if(this.kind() === 'PrivateIdentifier') {
                    const id = this.primary();
                    children.push(id);
                }
                else {
                    children.push(this.identifier());
                }
                const optional = question >= 0 || this.optionalChain(left);
                left = this.make('PropertyAccessExpression', pos, children);
                this.node(left).optional = optional;
                continue;
            }
            if(this.kind() === 'OpenBracketToken' && (question >= 0 || !this.decoratorContext)) {
                this.next();
                const argument = this.allowInExpression();
                this.expect('CloseBracketToken');
                const children = [left];
                if(question >= 0) {
                    children.push(question);
                }
                children.push(argument);
                const optional = question >= 0 || this.optionalChain(left);
                left = this.make('ElementAccessExpression', pos, children);
                this.node(left).optional = optional;
                continue;
            }
            if(allowCall && this.kind() === 'OpenParenToken') {
                const args = this.arguments();
                const trailing = this.lastTrailing;
                const original = this.node(left);
                const children =
                    original.kind === 'ExpressionWithTypeArguments' && question < 0
                        ? original.children.slice(0, 1)
                        : [left];
                if(question >= 0) {
                    children.push(question);
                }
                if(original.kind === 'ExpressionWithTypeArguments' && question < 0) {
                    for(const type of original.children.slice(1)) {
                        children.push(type);
                    }
                }
                for(const type of optionalTypes) {
                    children.push(type);
                }
                for(const argument of args) {
                    children.push(argument);
                }
                const optional = question >= 0 || this.optionalChain(children[0] ?? panic('missing callee'));
                left = this.make('CallExpression', pos, children);
                this.node(left).optional = optional;
                this.node(left).list = args.length;
                this.node(left).trailing = trailing;
                continue;
            }
            if(this.kind() === 'NoSubstitutionTemplateLiteral' || this.kind() === 'TemplateHead') {
                const template = this.kind() === 'TemplateHead' ? this.template() : this.literal();
                const original = this.node(left);
                const children =
                    original.kind === 'ExpressionWithTypeArguments' && question < 0
                        ? original.children.slice(0, 1)
                        : [left];
                if(question >= 0) {
                    children.push(question);
                }
                if(original.kind === 'ExpressionWithTypeArguments' && question < 0) {
                    for(const type of original.children.slice(1)) {
                        children.push(type);
                    }
                }
                children.push(template);
                const optional = question >= 0 || this.optionalChain(children[0] ?? panic('missing tag'));
                left = this.make('TaggedTemplateExpression', pos, children);
                this.node(left).optional = optional;
                continue;
            }
            if(question >= 0) {
                panic('optional chain missing member');
            }
            if(this.kind() === 'ExclamationToken' && (this.scanner.flags & 1) === 0) {
                this.next();
                left = this.make('NonNullExpression', pos, [left]);
                continue;
            }
            return left;
        }
        return left;
    }
    templatePart(): number {
        const id = this.make(this.kind(), this.scanner.fullStart);
        const node = this.node(id);
        node.text = this.scanner.value;
        node.literalFlags = this.scanner.flags & 7180;
        node.raw = this.scanner.text.slice(
            this.scanner.start + 1,
            this.scanner.pos - ((this.scanner.flags & 4) !== 0 ? 0 : this.kind() === 'TemplateTail' ? 1 : 2),
        );
        this.next();
        node.end = this.scanner.fullStart;
        return id;
    }
    templateType(): number {
        const pos = this.scanner.fullStart;
        const children = [this.templatePart()];
        while(true) {
            const start = this.scanner.fullStart;
            const type = this.type();
            this.scanner.rescanTemplate();
            const tail = this.kind() === 'TemplateTail';
            const literal = this.templatePart();
            children.push(this.make('TemplateLiteralTypeSpan', start, [type, literal]));
            if(tail) {
                break;
            }
        }
        return this.make('TemplateLiteralType', pos, children);
    }
    template(): number {
        const pos = this.scanner.fullStart;
        const children = [this.templatePart()];
        while(true) {
            const spanPos = this.scanner.fullStart;
            const expression = this.allowInExpression();
            if(this.kind() !== 'CloseBraceToken') {
                panic('template missing closing brace');
            }
            this.scanner.rescanTemplate();
            const tail = this.kind() === 'TemplateTail';
            const literal = this.templatePart();
            children.push(this.make('TemplateSpan', spanPos, [expression, literal]));
            if(tail) {
                break;
            }
        }
        return this.make('TemplateExpression', pos, children);
    }
    array(): number {
        const pos = this.scanner.fullStart;
        this.next();
        const multiLine = (this.scanner.flags & 1) !== 0;
        const children: number[] = [];
        let trailing = false;
        while(this.kind() !== 'CloseBracketToken') {
            children.push(
                this.kind() === 'CommaToken'
                    ? this.make('OmittedExpression', this.scanner.fullStart)
                    : this.kind() === 'DotDotDotToken'
                      ? this.spread()
                      : this.assignment(),
            );
            trailing = this.kind() === 'CommaToken';
            if(!trailing) {
                break;
            }
            this.next();
        }
        this.expect('CloseBracketToken');
        const id = this.make('ArrayLiteralExpression', pos, children);
        this.node(id).list = children.length;
        this.node(id).trailing = trailing;
        this.node(id).multiLine = multiLine;
        return id;
    }
    propertyName(): number {
        if(this.kind() === 'PrivateIdentifier') {
            return this.primary();
        }
        if(this.kind() === 'StringLiteral' || this.kind() === 'NumericLiteral') {
            return this.literal();
        }
        if(this.kind() === 'OpenBracketToken') {
            const pos = this.scanner.fullStart;
            this.next();
            const expression = this.rootExpression();
            this.expect('CloseBracketToken');
            return this.make('ComputedPropertyName', pos, [expression]);
        }
        return this.identifier();
    }
    methodBody(pos: number, prefix: readonly number[], kind: string, async: boolean, generator: boolean): number {
        const children = prefix.slice();
        const oldAwait = this.awaitContext;
        const oldYield = this.yieldContext;
        this.awaitContext = async;
        this.yieldContext = generator;
        const oldIn = this.disallowIn;
        this.disallowIn = false;
        for(const type of this.typeParameters()) {
            children.push(type);
        }
        for(const parameter of this.parameters()) {
            children.push(parameter);
        }
        if(this.kind() === 'ColonToken') {
            this.next();
            children.push(this.returnType());
        }
        if(this.kind() === 'OpenBraceToken') {
            children.push(this.block());
        }
        else {
            this.semicolon();
        }
        this.disallowIn = oldIn;
        this.awaitContext = oldAwait;
        this.yieldContext = oldYield;
        return this.make(kind, pos, children);
    }
    object(): number {
        const pos = this.scanner.fullStart;
        this.next();
        const multiLine = (this.scanner.flags & 1) !== 0;
        const properties: number[] = [];
        let trailing = false;
        while(this.kind() !== 'CloseBraceToken') {
            const start = this.scanner.fullStart;
            if(this.kind() === 'DotDotDotToken') {
                this.next();
                const expression = this.allowInAssignment();
                properties.push(this.make('SpreadAssignment', start, [expression]));
            }
            else {
                const children: number[] = [];
                let async = false;
                let methodKind = 'MethodDeclaration';
                if(
                    this.kind() === 'AsyncKeyword' &&
                    this.peek() !== 'ColonToken' &&
                    this.peek() !== 'EqualsToken' &&
                    this.peek() !== 'QuestionToken' &&
                    this.peek() !== 'ExclamationToken' &&
                    this.peek() !== 'CommaToken' &&
                    this.peek() !== 'CloseBraceToken' &&
                    this.peek() !== 'OpenParenToken'
                ) {
                    async = true;
                    children.push(this.token());
                }
                if(
                    (this.kind() === 'GetKeyword' || this.kind() === 'SetKeyword') &&
                    this.peek() !== 'ColonToken' &&
                    this.peek() !== 'EqualsToken' &&
                    this.peek() !== 'QuestionToken' &&
                    this.peek() !== 'ExclamationToken' &&
                    this.peek() !== 'CommaToken' &&
                    this.peek() !== 'OpenParenToken' &&
                    this.peek() !== 'CloseBraceToken'
                ) {
                    methodKind = this.kind() === 'GetKeyword' ? 'GetAccessor' : 'SetAccessor';
                    this.next();
                }
                const generator = this.kind() === 'AsteriskToken';
                if(generator) {
                    children.push(this.token());
                }
                children.push(this.propertyName());
                if(this.kind() === 'QuestionToken' || this.kind() === 'ExclamationToken') {
                    children.push(this.token());
                }
                if(this.kind() === 'OpenParenToken' || this.kind() === 'LessThanToken') {
                    properties.push(this.methodBody(start, children, methodKind, async, generator));
                }
                else {
                    const assignment = this.kind() === 'ColonToken';
                    if(assignment) {
                        this.next();
                        children.push(this.allowInAssignment());
                    }
                    else if(this.kind() === 'EqualsToken') {
                        children.push(this.token());
                        children.push(this.allowInAssignment());
                    }
                    properties.push(
                        this.make(assignment ? 'PropertyAssignment' : 'ShorthandPropertyAssignment', start, children),
                    );
                }
            }
            trailing = this.kind() === 'CommaToken';
            if(!trailing) {
                break;
            }
            this.next();
        }
        this.expect('CloseBraceToken');
        const id = this.make('ObjectLiteralExpression', pos, properties);
        this.node(id).list = properties.length;
        this.node(id).trailing = trailing;
        this.node(id).multiLine = multiLine;
        return id;
    }
    unary(): number {
        const pos = this.scanner.fullStart;
        const operator = this.kind();
        if(operator === 'LessThanToken' && this.jsx) {
            return this.suffix(this.jsxParser().element(true), true);
        }
        if(operator === 'LessThanToken') {
            this.next();
            const type = this.type();
            this.expect('GreaterThanToken');
            const expression = this.unary();
            return this.make('TypeAssertionExpression', pos, [type, expression]);
        }
        if(operator === 'AwaitKeyword' && (this.awaitContext || this.peek() === 'Identifier')) {
            this.next();
            const expression = this.unary();
            return this.make('AwaitExpression', pos, [expression]);
        }
        if(operator === 'YieldKeyword' && (this.yieldContext || this.yieldOperandAhead())) {
            this.next();
            const children: number[] = [];
            if(
                (this.scanner.flags & 1) === 0 &&
                this.kind() !== 'SemicolonToken' &&
                this.kind() !== 'CloseBraceToken'
            ) {
                if(this.kind() === 'AsteriskToken') {
                    children.push(this.token());
                }
                children.push(this.assignment());
            }
            return this.make('YieldExpression', pos, children);
        }
        if(
            operator === 'PlusToken' ||
            operator === 'MinusToken' ||
            operator === 'TildeToken' ||
            operator === 'ExclamationToken' ||
            operator === 'PlusPlusToken' ||
            operator === 'MinusMinusToken'
        ) {
            this.next();
            const operand = this.unary();
            const id = this.make('PrefixUnaryExpression', pos, [operand]);
            this.node(id).operator = operator;
            return id;
        }
        if(operator === 'DeleteKeyword' || operator === 'VoidKeyword' || operator === 'TypeOfKeyword') {
            this.next();
            const operand = this.unary();
            return this.make(
                operator === 'DeleteKeyword'
                    ? 'DeleteExpression'
                    : operator === 'VoidKeyword'
                      ? 'VoidExpression'
                      : 'TypeOfExpression',
                pos,
                [operand],
            );
        }
        const operand = this.suffix(this.primary(), true);
        const postfix = this.kind();
        if((postfix === 'PlusPlusToken' || postfix === 'MinusMinusToken') && (this.scanner.flags & 1) === 0) {
            this.next();
            const id = this.make('PostfixUnaryExpression', pos, [operand]);
            this.node(id).operator = postfix;
            return id;
        }
        return operand;
    }
    binary(minimum: number): number {
        const pos = this.scanner.fullStart;
        let left = this.unary();
        while(true) {
            this.scanner.rescanGreater();
            const operator = this.kind();
            const rank = precedence(operator);
            if(operator === 'InKeyword' && this.disallowIn) {
                break;
            }
            if(rank < minimum || (rank === minimum && operator !== 'AsteriskAsteriskToken')) {
                break;
            }
            if(operator === 'AsKeyword' || operator === 'SatisfiesKeyword') {
                if((this.scanner.flags & 1) !== 0) {
                    break;
                }
                this.next();
                const type = this.type();
                left = this.make(operator === 'AsKeyword' ? 'AsExpression' : 'SatisfiesExpression', pos, [left, type]);
                continue;
            }
            const token = this.token();
            const right = this.binary(rank);
            left = this.make('BinaryExpression', pos, [left, token, right]);
        }
        return left;
    }
    assignment(allowReturn = true): number {
        if((!this.jsx || jsxArrowAhead(this.scanner)) && arrowAhead(this.scanner, allowReturn)) {
            return this.arrow(allowReturn);
        }
        const pos = this.scanner.fullStart;
        let left = this.binary(0);
        const operator = this.kind();
        if(this.node(left).kind === 'Identifier' && operator === 'EqualsGreaterThanToken') {
            const parameter = this.make('Parameter', this.node(left).pos, [left]);
            const children = [parameter, this.token()];
            const oldAwait = this.awaitContext;
            const oldYield = this.yieldContext;
            this.awaitContext = false;
            this.yieldContext = false;
            children.push(this.kind() === 'OpenBraceToken' ? this.block() : this.assignment(allowReturn));
            this.awaitContext = oldAwait;
            this.yieldContext = oldYield;
            return this.make('ArrowFunction', pos, children);
        }
        if(operator === 'EqualsToken' || (operator.endsWith('EqualsToken') && precedence(operator) < 0)) {
            const token = this.token();
            const right = this.assignment();
            return this.make('BinaryExpression', pos, [left, token, right]);
        }
        if(operator === 'QuestionToken') {
            const question = this.token();
            const yes = this.allowInAssignment(false);
            if(this.kind() !== 'ColonToken') {
                panic('conditional missing colon');
            }
            const colon = this.token();
            const no = this.assignment();
            left = this.make('ConditionalExpression', pos, [left, question, yes, colon, no]);
        }
        return left;
    }
    allowInAssignment(allowReturn = true): number {
        const saved = this.disallowIn;
        this.disallowIn = false;
        const decorator = this.decoratorContext;
        this.decoratorContext = false;
        const result = this.assignment(allowReturn);
        this.decoratorContext = decorator;
        this.disallowIn = saved;
        return result;
    }
    allowInExpression(): number {
        const saved = this.disallowIn;
        this.disallowIn = false;
        const decorator = this.decoratorContext;
        this.decoratorContext = false;
        const result = this.expression();
        this.decoratorContext = decorator;
        this.disallowIn = saved;
        return result;
    }
    expression(): number {
        const pos = this.scanner.fullStart;
        let left = this.assignment();
        while(this.kind() === 'CommaToken') {
            const comma = this.token();
            const right = this.assignment();
            left = this.make('BinaryExpression', pos, [left, comma, right]);
        }
        return left;
    }
    file(): number {
        const parser = this.statements();
        const children: number[] = [];
        while(this.kind() !== 'EndOfFile') {
            children.push(parser.statement());
        }
        const eof = this.make('EndOfFile', this.scanner.fullStart);
        children.push(eof);
        this.node(eof).end = this.scanner.text.length;
        const root = this.make('SourceFile', 0, children);
        this.node(root).end = this.scanner.text.length;
        return root;
    }
}
