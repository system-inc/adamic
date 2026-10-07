// Speculation saves and restores scanner state without building a tree.
import type { Scanner } from '../scanner/scanner.ts';
import { modifierKinds, precedence, reservedKinds } from './grammar.ts';

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

export function arrowAhead(scanner: Scanner, allowReturn: boolean): boolean {
    if(kind(scanner) === 'EqualsGreaterThanToken') {
        return true;
    }
    const state = new Speculation(scanner);
    let result = false;
    if(kind(scanner) === 'AsyncKeyword') {
        state.next();
        if((scanner.flags & 1) !== 0) {
            state.restore();
            return false;
        }
        if(
            kind(scanner) === 'Identifier' ||
            (kind(scanner).endsWith('Keyword') && !reservedKinds.includes(kind(scanner)))
        ) {
            state.next();
            result = kind(scanner) === 'EqualsGreaterThanToken' && (scanner.flags & 1) === 0;
            state.restore();
            return result;
        }
    }
    const generic = kind(scanner) === 'LessThanToken';
    if(generic) {
        let depth = 0;
        while(kind(scanner) !== 'EndOfFile') {
            if(kind(scanner) === 'LessThanToken') {
                depth++;
            }
            if(kind(scanner) === 'GreaterThanToken') {
                depth--;
            }
            state.next();
            if(depth === 0) {
                break;
            }
        }
    }
    if(kind(scanner) === 'OpenParenToken') {
        const parameterState = new Speculation(scanner);
        parameterState.next();
        const first = kind(scanner);
        parameterState.next();
        const second = kind(scanner);
        parameterState.restore();
        if(
            !generic &&
            modifierKinds.includes(first) &&
            first !== 'AsyncKeyword' &&
            second !== 'AsKeyword' &&
            (second === 'Identifier' || (second.endsWith('Keyword') && !reservedKinds.includes(second)))
        ) {
            state.restore();
            return true;
        }
        const parameterHead =
            first === 'Identifier' ||
            first === 'ThisKeyword' ||
            (first.endsWith('Keyword') && !reservedKinds.includes(first));
        let depth = 0;
        let typed = false;
        let head = 0;
        while(kind(scanner) !== 'EndOfFile') {
            if(kind(scanner) === 'OpenParenToken') {
                depth++;
            }
            if(kind(scanner) === 'CloseParenToken') {
                depth--;
            }
            if(depth === 1 && kind(scanner) !== 'OpenParenToken') {
                head++;
                if(
                    (parameterHead &&
                        head === 2 &&
                        (kind(scanner) === 'ColonToken' ||
                            (kind(scanner) === 'QuestionToken' && optionalParameterAhead(scanner)))) ||
                    (head === 1 && first === 'DotDotDotToken')
                ) {
                    typed = true;
                }
            }
            state.next();
            if(depth === 0) {
                break;
            }
        }
        if(
            !generic &&
            (typed || (head === 0 && (kind(scanner) === 'ColonToken' || kind(scanner) === 'OpenBraceToken')))
        ) {
            state.restore();
            return true;
        }
        if(kind(scanner) === 'EqualsGreaterThanToken' && (scanner.flags & 1) === 0) {
            result = true;
        }
        else if(kind(scanner) === 'ColonToken' && (allowReturn || typed)) {
            state.next();
            let braces = 0;
            while(kind(scanner) !== 'EndOfFile' && kind(scanner) !== 'EqualsGreaterThanToken') {
                if(kind(scanner) === 'SemicolonToken' && braces === 0) {
                    break;
                }
                if(kind(scanner) === 'OpenBraceToken') {
                    braces++;
                }
                if(kind(scanner) === 'CloseBraceToken') {
                    if(braces === 0) {
                        break;
                    }
                    braces--;
                }
                state.next();
            }
            result = kind(scanner) === 'EqualsGreaterThanToken';
        }
    }
    state.restore();
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

export function typeArgumentsAhead(scanner: Scanner): boolean {
    const state = new Speculation(scanner);
    let depth = 0;
    let braces = 0;
    let parens = 0;
    let brackets = 0;
    let closed = false;
    let conditionalType = false;
    while(kind(scanner) !== 'EndOfFile') {
        if(
            (kind(scanner) === 'CloseParenToken' && parens === 0) ||
            (kind(scanner) === 'CloseBraceToken' && braces === 0) ||
            (kind(scanner) === 'CloseBracketToken' && brackets === 0)
        ) {
            break;
        }
        if(kind(scanner) === 'ExtendsKeyword') {
            conditionalType = true;
        }
        if(kind(scanner) === 'QuestionToken' && !conditionalType && braces === 0 && parens === 0 && brackets === 0) {
            break;
        }
        if(kind(scanner) === 'OpenBracketToken') {
            brackets++;
        }
        if(kind(scanner) === 'CloseBracketToken') {
            brackets--;
        }
        if(kind(scanner) === 'SemicolonToken' && braces === 0 && parens === 0) {
            break;
        }
        if(kind(scanner) === 'OpenBraceToken') {
            braces++;
        }
        if(kind(scanner) === 'CloseBraceToken') {
            braces--;
        }
        if(kind(scanner) === 'OpenParenToken') {
            parens++;
        }
        if(kind(scanner) === 'CloseParenToken') {
            parens--;
        }
        if(kind(scanner) === 'LessThanToken') {
            depth++;
        }
        if(kind(scanner) === 'GreaterThanToken') {
            if(depth === 1) {
                scanner.rescanGreater();
                if(kind(scanner) !== 'GreaterThanToken') {
                    break;
                }
            }
            depth--;
        }
        if(
            kind(scanner) === 'PlusToken' ||
            kind(scanner) === 'MinusToken' ||
            kind(scanner) === 'SlashToken' ||
            kind(scanner) === 'BarBarToken' ||
            kind(scanner) === 'AmpersandAmpersandToken'
        ) {
            break;
        }
        state.next();
        if(depth === 0) {
            closed = true;
            break;
        }
    }
    const after = kind(scanner);
    const result =
        closed &&
        (after === 'OpenParenToken' ||
            after === 'NoSubstitutionTemplateLiteral' ||
            after === 'TemplateHead' ||
            (after !== 'LessThanToken' &&
                after !== 'GreaterThanToken' &&
                after !== 'PlusToken' &&
                after !== 'MinusToken' &&
                ((scanner.flags & 1) !== 0 ||
                    precedence(after) >= 0 ||
                    after === 'SemicolonToken' ||
                    after === 'CloseParenToken' ||
                    after === 'CloseBracketToken' ||
                    after === 'CommaToken' ||
                    after === 'EndOfFile' ||
                    after === 'DotToken' ||
                    after === 'QuestionDotToken' ||
                    after === 'QuestionToken' ||
                    after === 'ColonToken' ||
                    after === 'CloseBraceToken' ||
                    after === 'EqualsToken')));
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
