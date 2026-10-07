// Speculation saves and restores scanner state without building a tree.
import type { ParseNode } from './nodes.ts';
import type { Scanner } from '../scanner/scanner.ts';
import { modifierKinds, precedence, reservedKinds, bindingStart } from './grammar.ts';

export interface ParserStateInterface {
    readonly pos: number;
    readonly start: number;
    readonly fullStart: number;
    readonly kind: string;
    readonly value: string;
    readonly flags: number;
    readonly errors: number;
    readonly nodes: number;
    readonly roots: number;
    readonly diagnostics: number;
}

function kind(scanner: Scanner): string {
    return scanner.kind;
}
class Speculation {
    readonly scanner: Scanner;
    readonly pos: number;
    readonly start: number;
    readonly fullStart: number;
    readonly kind: string;
    readonly value: string;
    readonly flags: number;
    readonly errors: number;
    braces = 0;
    readonly templates: number[] = [];
    previous = '';
    constructor(scanner: Scanner) {
        this.scanner = scanner;
        this.pos = scanner.pos;
        this.start = scanner.start;
        this.fullStart = scanner.fullStart;
        this.kind = scanner.kind;
        this.value = scanner.value;
        this.flags = scanner.flags;
        this.errors = scanner.errors.length;
    }
    prepare(): void {
        const current = kind(this.scanner);
        if(
            (current === 'SlashToken' || current === 'SlashEqualsToken') &&
            (this.previous === 'EqualsToken' ||
                this.previous === 'OpenParenToken' ||
                this.previous === 'CommaToken' ||
                this.previous === 'ColonToken' ||
                this.previous === 'QuestionToken' ||
                precedence(this.previous) >= 0)
        ) {
            this.scanner.rescanSlash();
        }
        if(
            kind(this.scanner) === 'CloseBraceToken' &&
            this.templates.length > 0 &&
            this.templates[this.templates.length - 1] === this.braces
        ) {
            this.scanner.rescanTemplate();
            if(kind(this.scanner) === 'TemplateTail') {
                this.templates.pop();
            }
        }
        if(kind(this.scanner) === 'TemplateHead') {
            this.templates.push(this.braces);
        }
        if(kind(this.scanner) === 'OpenBraceToken') {
            this.braces++;
        }
        if(kind(this.scanner) === 'CloseBraceToken') {
            this.braces--;
        }
    }
    next(): void {
        this.previous = kind(this.scanner);
        this.scanner.scan();
        this.prepare();
    }
    restore(): void {
        this.scanner.pos = this.pos;
        this.scanner.start = this.start;
        this.scanner.fullStart = this.fullStart;
        this.scanner.kind = this.kind;
        this.scanner.value = this.value;
        this.scanner.flags = this.flags;
        this.scanner.errors.splice(this.errors);
    }
}

function optionalParameterAhead(scanner: Scanner): boolean {
    const state = new Speculation(scanner);
    state.next();
    const after = kind(scanner);
    state.restore();
    return after === 'ColonToken' || after === 'CommaToken' || after === 'EqualsToken' || after === 'CloseParenToken';
}

// 0 rejects an arrow, 1 commits to recovery, and 2 requires a complete speculative signature.
export function arrowHead(scanner: Scanner): number {
    if(kind(scanner) === 'EqualsGreaterThanToken') {
        return 1;
    }
    const state = new Speculation(scanner);
    if(kind(scanner) === 'AsyncKeyword') {
        state.next();
        if((scanner.flags & 1) !== 0) {
            state.restore();
            return 0;
        }
        if(
            kind(scanner) === 'Identifier' ||
            (kind(scanner).endsWith('Keyword') && !reservedKinds.includes(kind(scanner)))
        ) {
            state.next();
            const result = kind(scanner) === 'EqualsGreaterThanToken' && (scanner.flags & 1) === 0 ? 1 : 0;
            state.restore();
            return result;
        }
    }
    const generic = kind(scanner) === 'LessThanToken';
    if(!generic && kind(scanner) !== 'OpenParenToken') {
        state.restore();
        return 0;
    }
    state.next();
    const first = kind(scanner);
    const identifier = first === 'Identifier' || (first.endsWith('Keyword') && !reservedKinds.includes(first));
    if(generic) {
        state.restore();
        return identifier || first === 'ConstKeyword' ? 2 : 0;
    }
    if(first === 'CloseParenToken') {
        state.next();
        const result = ['EqualsGreaterThanToken', 'ColonToken', 'OpenBraceToken'].includes(kind(scanner)) ? 1 : 0;
        state.restore();
        return result;
    }
    if(['OpenBracketToken', 'OpenBraceToken', 'DotDotDotToken'].includes(first)) {
        state.restore();
        return first === 'DotDotDotToken' ? 1 : 2;
    }
    state.next();
    const second = kind(scanner);
    let result = 0;
    if(
        modifierKinds.includes(first) &&
        first !== 'AsyncKeyword' &&
        second !== 'AsKeyword' &&
        (second === 'Identifier' || (second.endsWith('Keyword') && !reservedKinds.includes(second)))
    ) {
        result = 1;
    }
    else if(identifier || first === 'ThisKeyword') {
        if(second === 'ColonToken' || (second === 'QuestionToken' && optionalParameterAhead(scanner))) {
            result = 1;
        }
        else if(['CommaToken', 'EqualsToken', 'CloseParenToken'].includes(second)) {
            result = 2;
        }
    }
    state.restore();
    return result;
}

// A modifier is a declaration start only when its following tokens commit to
// a declaration. Source-list recovery skips a stranded export token.
export function declarationAhead(scanner: Scanner): boolean {
    const state = new Speculation(scanner);
    let result = false;
    while(kind(scanner) !== 'EndOfFile') {
        const current = kind(scanner);
        if(
            ['VarKeyword', 'LetKeyword', 'ConstKeyword', 'FunctionKeyword', 'ClassKeyword', 'EnumKeyword'].includes(
                current,
            )
        ) {
            result = true;
            break;
        }
        if(['InterfaceKeyword', 'TypeKeyword', 'DeferKeyword', 'ModuleKeyword', 'NamespaceKeyword'].includes(current)) {
            state.next();
            result =
                (scanner.flags & 1) === 0 &&
                (kind(scanner) === 'Identifier' ||
                    (kind(scanner).endsWith('Keyword') && !reservedKinds.includes(kind(scanner))) ||
                    ((current === 'ModuleKeyword' || current === 'NamespaceKeyword') &&
                        kind(scanner) === 'StringLiteral'));
            break;
        }
        if(current === 'ExportKeyword') {
            state.next();
            if(
                ['EqualsToken', 'AsteriskToken', 'OpenBraceToken', 'DefaultKeyword', 'AsKeyword', 'AtToken'].includes(
                    kind(scanner),
                )
            ) {
                result = true;
                break;
            }
            if(kind(scanner) === 'TypeKeyword') {
                state.next();
                result =
                    kind(scanner) === 'AsteriskToken' ||
                    kind(scanner) === 'OpenBraceToken' ||
                    ((kind(scanner) === 'Identifier' ||
                        (kind(scanner).endsWith('Keyword') && !reservedKinds.includes(kind(scanner)))) &&
                        (scanner.flags & 1) === 0);
                break;
            }
            continue;
        }
        if(current === 'ImportKeyword') {
            state.next();
            result =
                ['StringLiteral', 'AsteriskToken', 'OpenBraceToken', 'Identifier'].includes(kind(scanner)) ||
                kind(scanner).endsWith('Keyword');
            break;
        }
        if(current === 'GlobalKeyword') {
            state.next();
            result = ['OpenBraceToken', 'Identifier', 'ExportKeyword'].includes(kind(scanner));
            break;
        }
        if(
            [
                'AbstractKeyword',
                'AccessorKeyword',
                'AsyncKeyword',
                'DeclareKeyword',
                'PrivateKeyword',
                'ProtectedKeyword',
                'PublicKeyword',
                'ReadonlyKeyword',
                'StaticKeyword',
            ].includes(current)
        ) {
            state.next();
            if(current !== 'StaticKeyword' && (scanner.flags & 1) !== 0) {
                break;
            }
            if(current === 'DeclareKeyword' && kind(scanner) === 'TypeKeyword') {
                result = true;
                break;
            }
            continue;
        }
        break;
    }
    state.restore();
    return result;
}

export function statementAhead(scanner: Scanner): boolean {
    const current = kind(scanner);
    if(current === 'ExportKeyword' || current === 'ConstKeyword') {
        return declarationAhead(scanner);
    }
    if(current === 'ImportKeyword') {
        const state = new Speculation(scanner);
        state.next();
        const expression = ['OpenParenToken', 'LessThanToken', 'DotToken'].includes(kind(scanner));
        state.restore();
        return expression || declarationAhead(scanner);
    }
    if(
        [
            'AccessorKeyword',
            'PublicKeyword',
            'PrivateKeyword',
            'ProtectedKeyword',
            'StaticKeyword',
            'ReadonlyKeyword',
        ].includes(current)
    ) {
        if(declarationAhead(scanner)) {
            return true;
        }
        const state = new Speculation(scanner);
        state.next();
        const identifier =
            (kind(scanner) === 'Identifier' || kind(scanner).endsWith('Keyword')) && (scanner.flags & 1) === 0;
        state.restore();
        return !identifier;
    }
    return (
        [
            'AtToken',
            'SemicolonToken',
            'OpenBraceToken',
            'VarKeyword',
            'LetKeyword',
            'UsingKeyword',
            'FunctionKeyword',
            'ClassKeyword',
            'EnumKeyword',
            'IfKeyword',
            'DoKeyword',
            'WhileKeyword',
            'ForKeyword',
            'ContinueKeyword',
            'BreakKeyword',
            'ReturnKeyword',
            'WithKeyword',
            'SwitchKeyword',
            'ThrowKeyword',
            'TryKeyword',
            'DebuggerKeyword',
            'CatchKeyword',
            'FinallyKeyword',
            'ThisKeyword',
            'SuperKeyword',
            'NullKeyword',
            'TrueKeyword',
            'FalseKeyword',
            'NumericLiteral',
            'BigIntLiteral',
            'StringLiteral',
            'NoSubstitutionTemplateLiteral',
            'TemplateHead',
            'OpenParenToken',
            'OpenBracketToken',
            'NewKeyword',
            'SlashToken',
            'SlashEqualsToken',
            'PlusToken',
            'MinusToken',
            'TildeToken',
            'ExclamationToken',
            'DeleteKeyword',
            'TypeOfKeyword',
            'VoidKeyword',
            'PlusPlusToken',
            'MinusMinusToken',
            'LessThanToken',
            'AwaitKeyword',
            'YieldKeyword',
            'PrivateIdentifier',
        ].includes(current) ||
        precedence(current) >= 0 ||
        current === 'Identifier' ||
        (current.endsWith('Keyword') && !reservedKinds.includes(current))
    );
}

export function typeMemberAhead(scanner: Scanner): boolean {
    if(['OpenParenToken', 'LessThanToken', 'GetKeyword', 'SetKeyword'].includes(kind(scanner))) {
        return true;
    }
    const state = new Speculation(scanner);
    let identifier = false;
    while(modifierKinds.includes(kind(scanner))) {
        identifier = true;
        state.next();
    }
    if(kind(scanner) === 'OpenBracketToken') {
        state.restore();
        return true;
    }
    if(
        kind(scanner) === 'Identifier' ||
        kind(scanner).endsWith('Keyword') ||
        ['StringLiteral', 'NumericLiteral', 'BigIntLiteral'].includes(kind(scanner))
    ) {
        identifier = true;
        state.next();
    }
    const result =
        identifier &&
        ([
            'OpenParenToken',
            'LessThanToken',
            'QuestionToken',
            'ColonToken',
            'CommaToken',
            'SemicolonToken',
            'CloseBraceToken',
            'EndOfFile',
        ].includes(kind(scanner)) ||
            (scanner.flags & 1) !== 0);
    state.restore();
    return result;
}

// First tokens shared by normal type parsing and parameter recovery.
export function typeTokenStart(tokenKind: string): boolean {
    return (
        tokenKind === 'Identifier' ||
        (tokenKind.endsWith('Keyword') && !reservedKinds.includes(tokenKind)) ||
        [
            'AnyKeyword',
            'UnknownKeyword',
            'StringKeyword',
            'NumberKeyword',
            'BigIntKeyword',
            'BooleanKeyword',
            'ReadonlyKeyword',
            'SymbolKeyword',
            'UniqueKeyword',
            'VoidKeyword',
            'UndefinedKeyword',
            'NullKeyword',
            'ThisKeyword',
            'TypeOfKeyword',
            'NeverKeyword',
            'OpenBraceToken',
            'OpenBracketToken',
            'LessThanToken',
            'BarToken',
            'AmpersandToken',
            'NewKeyword',
            'StringLiteral',
            'NumericLiteral',
            'BigIntLiteral',
            'TrueKeyword',
            'FalseKeyword',
            'ObjectKeyword',
            'AsteriskToken',
            'QuestionToken',
            'ExclamationToken',
            'DotDotDotToken',
            'InferKeyword',
            'ImportKeyword',
            'AssertsKeyword',
            'NoSubstitutionTemplateLiteral',
            'TemplateHead',
            'FunctionKeyword',
        ].includes(tokenKind)
    );
}

export function modifierAhead(scanner: Scanner, permitConst: boolean): boolean {
    const current = kind(scanner);
    if(!modifierKinds.includes(current)) {
        return false;
    }
    const state = new Speculation(scanner);
    state.next();
    let result: boolean;
    if(current === 'DefaultKeyword' || (current === 'ExportKeyword' && kind(scanner) === 'DefaultKeyword')) {
        if(current === 'ExportKeyword') {
            state.next();
        }
        result = ['ClassKeyword', 'FunctionKeyword', 'InterfaceKeyword', 'AtToken'].includes(kind(scanner));
        if(kind(scanner) === 'AbstractKeyword' || kind(scanner) === 'AsyncKeyword') {
            const next = kind(scanner) === 'AbstractKeyword' ? 'ClassKeyword' : 'FunctionKeyword';
            state.next();
            result = kind(scanner) === next && (scanner.flags & 1) === 0;
        }
    }
    else {
        if(current === 'ExportKeyword' && kind(scanner) === 'TypeKeyword') {
            state.next();
        }
        const follow =
            kind(scanner) === 'Identifier' ||
            kind(scanner).endsWith('Keyword') ||
            [
                'OpenBracketToken',
                'OpenBraceToken',
                'AsteriskToken',
                'DotDotDotToken',
                'StringLiteral',
                'NumericLiteral',
                'BigIntLiteral',
            ].includes(kind(scanner));
        result =
            current === 'ExportKeyword'
                ? kind(scanner) === 'AtToken' ||
                  (!['AsteriskToken', 'AsKeyword', 'OpenBraceToken'].includes(kind(scanner)) && follow)
                : current === 'ConstKeyword' && !permitConst
                  ? kind(scanner) === 'EnumKeyword'
                  : follow && (current === 'StaticKeyword' || (scanner.flags & 1) === 0);
    }
    state.restore();
    return result;
}

// Recognition precedes descent so invalid class tokens cannot create missing
// properties forever. Keep the same recovery boundary as typescript-go.
export function classMemberAhead(scanner: Scanner): boolean {
    if(kind(scanner) === 'AtToken') {
        return true;
    }
    const state = new Speculation(scanner);
    let name = '';
    while(modifierKinds.includes(kind(scanner))) {
        name = kind(scanner);
        if(
            [
                'PublicKeyword',
                'PrivateKeyword',
                'ProtectedKeyword',
                'StaticKeyword',
                'ReadonlyKeyword',
                'OverrideKeyword',
                'AccessorKeyword',
            ].includes(name)
        ) {
            state.restore();
            return true;
        }
        state.next();
    }
    if(kind(scanner) === 'AsteriskToken') {
        state.restore();
        return true;
    }
    if(
        ['Identifier', 'PrivateIdentifier'].includes(kind(scanner)) ||
        kind(scanner).endsWith('Keyword') ||
        ['StringLiteral', 'NumericLiteral', 'BigIntLiteral'].includes(kind(scanner))
    ) {
        name = kind(scanner);
        state.next();
    }
    const result =
        kind(scanner) === 'OpenBracketToken' ||
        (name !== '' &&
            (!name.endsWith('Keyword') ||
                ['GetKeyword', 'SetKeyword'].includes(name) ||
                [
                    'OpenParenToken',
                    'LessThanToken',
                    'ExclamationToken',
                    'ColonToken',
                    'EqualsToken',
                    'QuestionToken',
                    'SemicolonToken',
                    'CloseBraceToken',
                    'EndOfFile',
                ].includes(kind(scanner)) ||
                (scanner.flags & 1) !== 0));
    state.restore();
    return result;
}

export function heritageElementAhead(
    scanner: Scanner,
    recovery: boolean,
    awaitContext: boolean,
    yieldContext: boolean,
): boolean {
    const current = kind(scanner);
    const state = new Speculation(scanner);
    state.next();
    const after = kind(scanner);
    state.next();
    const following = kind(scanner);
    state.restore();
    if(current === 'OpenBraceToken') {
        return (
            after !== 'CloseBraceToken' ||
            ['CommaToken', 'OpenBraceToken', 'ExtendsKeyword', 'ImplementsKeyword'].includes(following)
        );
    }
    const identifier =
        (current === 'Identifier' || (current.endsWith('Keyword') && !reservedKinds.includes(current))) &&
        !(current === 'AwaitKeyword' && awaitContext) &&
        !(current === 'YieldKeyword' && yieldContext);
    const leftHandSide =
        identifier ||
        [
            'ThisKeyword',
            'SuperKeyword',
            'NullKeyword',
            'TrueKeyword',
            'FalseKeyword',
            'NumericLiteral',
            'BigIntLiteral',
            'StringLiteral',
            'NoSubstitutionTemplateLiteral',
            'TemplateHead',
            'OpenParenToken',
            'OpenBracketToken',
            'FunctionKeyword',
            'ClassKeyword',
            'NewKeyword',
            'SlashToken',
            'SlashEqualsToken',
        ].includes(current) ||
        (current === 'ImportKeyword' && ['OpenParenToken', 'LessThanToken', 'DotToken'].includes(after));
    const clause =
        ['ExtendsKeyword', 'ImplementsKeyword'].includes(current) &&
        (after === 'Identifier' ||
            (after.endsWith('Keyword') && !reservedKinds.includes(after)) ||
            [
                'ThisKeyword',
                'SuperKeyword',
                'NullKeyword',
                'TrueKeyword',
                'FalseKeyword',
                'NumericLiteral',
                'BigIntLiteral',
                'StringLiteral',
                'NoSubstitutionTemplateLiteral',
                'TemplateHead',
                'OpenParenToken',
                'OpenBracketToken',
                'OpenBraceToken',
                'FunctionKeyword',
                'ClassKeyword',
                'NewKeyword',
                'SlashToken',
                'SlashEqualsToken',
                'PlusToken',
                'MinusToken',
                'TildeToken',
                'ExclamationToken',
                'DeleteKeyword',
                'TypeOfKeyword',
                'VoidKeyword',
                'PlusPlusToken',
                'MinusMinusToken',
                'LessThanToken',
                'AwaitKeyword',
                'YieldKeyword',
                'PrivateIdentifier',
            ].includes(after));
    return (recovery ? identifier : leftHandSide) && !clause;
}

// Contextual get/set require a property name, including across a line break.
export function accessorAhead(scanner: Scanner): boolean {
    if(!['GetKeyword', 'SetKeyword'].includes(kind(scanner))) {
        return false;
    }
    const state = new Speculation(scanner);
    state.next();
    const result =
        kind(scanner) === 'Identifier' ||
        kind(scanner).endsWith('Keyword') ||
        ['PrivateIdentifier', 'OpenBracketToken', 'StringLiteral', 'NumericLiteral', 'BigIntLiteral'].includes(
            kind(scanner),
        );
    state.restore();
    return result;
}

// A statement after an arrow recovers as a block with a missing opening brace.
export function arrowBlockAhead(scanner: Scanner, expressionStart: boolean): boolean {
    const current = kind(scanner);
    return (
        current === 'OpenBraceToken' ||
        (!['SemicolonToken', 'FunctionKeyword', 'ClassKeyword'].includes(current) &&
            statementAhead(scanner) &&
            (!expressionStart || current === 'AtToken'))
    );
}

export interface ExpressionSpeculationInterface {
    readonly scanner: Scanner;
    readonly mark: () => ParserStateInterface;
    readonly rewind: (state: ParserStateInterface) => void;
    readonly next: () => void;
    readonly kind: () => string;
    readonly peek: () => string;
    readonly typeParameters: () => number[];
    readonly speculativeParameters: () => boolean;
    readonly node: (index: number) => ParseNode;
    readonly diagnosticCount: () => number;
    readonly bindingName: () => number;
    readonly bindingIdentifier: () => boolean;
    readonly nextIdentifierSameLine: () => boolean;
    readonly returnType: () => number;
    readonly arrow: (allowReturn: boolean) => number;
    readonly typeArgumentList: () => number[];
    readonly expressionStart: () => boolean;
}

export function expressionTypeArguments(parser: ExpressionSpeculationInterface): number[] | undefined {
    const state = parser.mark();
    parser.next();
    const types = parser.typeArgumentList();
    parser.scanner.rescanGreater();
    if(parser.kind() === 'GreaterThanToken') {
        parser.next();
        const after = parser.kind();
        if(
            ['OpenParenToken', 'NoSubstitutionTemplateLiteral', 'TemplateHead'].includes(after) ||
            (!['LessThanToken', 'GreaterThanToken', 'PlusToken', 'MinusToken'].includes(after) &&
                ((parser.scanner.flags & 1) !== 0 || precedence(after) >= 0 || !parser.expressionStart()))
        ) {
            return types;
        }
    }
    parser.rewind(state);
    return undefined;
}

function arrowReturnTypeValid(parser: ExpressionSpeculationInterface, index: number): boolean {
    const node = parser.node(index);
    if(node.kind === 'TypeReference') {
        const name = parser.node(node.children[0] ?? -1);
        return name.pos !== name.end;
    }
    if(['ParenthesizedType', 'FunctionType', 'ConstructorType'].includes(node.kind)) {
        return arrowReturnTypeValid(parser, node.children.at(-1) ?? -1);
    }
    return true;
}

export function possibleArrowAhead(parser: ExpressionSpeculationInterface, allowReturn: boolean): boolean {
    const state = parser.mark();
    if(parser.kind() === 'AsyncKeyword') {
        parser.next();
        if((parser.scanner.flags & 1) !== 0) {
            parser.rewind(state);
            return false;
        }
    }
    if(parser.kind() === 'LessThanToken') {
        parser.typeParameters();
    }
    const valid = parser.kind() === 'OpenParenToken';
    if(!valid) {
        parser.rewind(state);
        return false;
    }
    if(!parser.speculativeParameters()) {
        parser.rewind(state);
        return false;
    }
    const hasReturn = parser.kind() === 'ColonToken';
    if(hasReturn) {
        parser.next();
        if(!arrowReturnTypeValid(parser, parser.returnType())) {
            parser.rewind(state);
            return false;
        }
    }
    let result = ['EqualsGreaterThanToken', 'OpenBraceToken'].includes(parser.kind());
    parser.rewind(state);
    if(result && hasReturn && !allowReturn) {
        parser.arrow(allowReturn);
        result = parser.kind() === 'ColonToken';
        parser.rewind(state);
    }
    return result;
}

// Outside their contexts, await/yield still recover before an identifier or literal on the same line.
export function contextualExpressionAhead(scanner: Scanner, context: boolean): boolean {
    if(context) {
        return true;
    }
    const state = new Speculation(scanner);
    state.next();
    const after = kind(scanner);
    const result =
        (scanner.flags & 1) === 0 &&
        (after === 'Identifier' ||
            after.endsWith('Keyword') ||
            ['NumericLiteral', 'BigIntLiteral', 'StringLiteral'].includes(after));
    state.restore();
    return result;
}

export function functionTypeAhead(parser: ExpressionSpeculationInterface): boolean {
    if(parser.kind() === 'LessThanToken') {
        return true;
    }
    const saved = parser.mark();
    parser.next();
    let result = parser.kind() === 'CloseParenToken' || parser.kind() === 'DotDotDotToken';
    if(!result) {
        while(
            ['PublicKeyword', 'PrivateKeyword', 'ProtectedKeyword', 'ReadonlyKeyword', 'OverrideKeyword'].includes(
                parser.kind(),
            ) &&
            parser.nextIdentifierSameLine()
        ) {
            parser.next();
        }
        let parameter = false;
        if(parser.bindingIdentifier() || parser.kind() === 'ThisKeyword') {
            parser.next();
            parameter = true;
        }
        else if(parser.kind() === 'OpenBracketToken' || parser.kind() === 'OpenBraceToken') {
            parser.bindingName();
            parameter = parser.diagnosticCount() === saved.diagnostics;
        }
        if(parameter) {
            result = ['ColonToken', 'CommaToken', 'QuestionToken', 'EqualsToken'].includes(parser.kind());
            if(parser.kind() === 'CloseParenToken') {
                parser.next();
                result = parser.kind() === 'EqualsGreaterThanToken';
            }
        }
    }
    parser.rewind(saved);
    return result;
}

// JSX generic arrows require a comma, default or a real constraint.
export function jsxArrowAhead(scanner: Scanner): boolean {
    const state = new Speculation(scanner);
    if(kind(scanner) === 'AsyncKeyword') {
        state.next();
    }
    if(kind(scanner) !== 'LessThanToken') {
        state.restore();
        return true;
    }
    state.next();
    if(kind(scanner) === 'ConstKeyword') {
        state.next();
    }
    state.next();
    let result = kind(scanner) === 'CommaToken' || kind(scanner) === 'EqualsToken';
    if(kind(scanner) === 'ExtendsKeyword') {
        state.next();
        result =
            kind(scanner) !== 'EqualsToken' && kind(scanner) !== 'GreaterThanToken' && kind(scanner) !== 'SlashToken';
    }
    state.restore();
    return result;
}

export function tupleNameAhead(scanner: Scanner): boolean {
    if(!bindingStart(kind(scanner))) {
        return false;
    }
    const state = new Speculation(scanner);
    state.next();
    if(kind(scanner) === 'QuestionToken') {
        state.next();
    }
    const result = kind(scanner) === 'ColonToken';
    state.restore();
    return result;
}
export function nextIdentifierSameLine(scanner: Scanner): boolean {
    const state = new Speculation(scanner);
    state.next();
    const result = bindingStart(kind(scanner)) && (scanner.flags & 1) === 0;
    state.restore();
    return result;
}
