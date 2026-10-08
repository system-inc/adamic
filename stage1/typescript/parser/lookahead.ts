// Speculation saves and restores scanner state without building a tree.
import type { Scanner } from '../scanner/scanner.ts';
import { isModifierKind, precedence, isReservedKind, bindingStart, propertyNameStart } from './grammar.ts';

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

function declarationModifierStart(tokenKind: string): boolean {
    switch(tokenKind) {
        case 'AbstractKeyword':
        case 'AccessorKeyword':
        case 'AsyncKeyword':
        case 'DeclareKeyword':
        case 'PrivateKeyword':
        case 'ProtectedKeyword':
        case 'PublicKeyword':
        case 'ReadonlyKeyword':
        case 'StaticKeyword':
            return true;
        default:
            return false;
    }
}

function typeKeywordStart(tokenKind: string): boolean {
    switch(tokenKind) {
        case 'AnyKeyword':
        case 'UnknownKeyword':
        case 'StringKeyword':
        case 'NumberKeyword':
        case 'BigIntKeyword':
        case 'BooleanKeyword':
        case 'ReadonlyKeyword':
        case 'SymbolKeyword':
        case 'UniqueKeyword':
        case 'VoidKeyword':
        case 'UndefinedKeyword':
        case 'NullKeyword':
        case 'ThisKeyword':
        case 'TypeOfKeyword':
        case 'NeverKeyword':
        case 'OpenBraceToken':
        case 'OpenBracketToken':
        case 'LessThanToken':
        case 'BarToken':
        case 'AmpersandToken':
        case 'NewKeyword':
        case 'StringLiteral':
        case 'NumericLiteral':
        case 'BigIntLiteral':
        case 'TrueKeyword':
        case 'FalseKeyword':
        case 'ObjectKeyword':
        case 'AsteriskToken':
        case 'QuestionToken':
        case 'ExclamationToken':
        case 'DotDotDotToken':
        case 'InferKeyword':
        case 'ImportKeyword':
        case 'AssertsKeyword':
        case 'NoSubstitutionTemplateLiteral':
        case 'TemplateHead':
        case 'FunctionKeyword':
            return true;
        default:
            return false;
    }
}

function statementModifierStart(tokenKind: string): boolean {
    switch(tokenKind) {
        case 'AccessorKeyword':
        case 'PublicKeyword':
        case 'PrivateKeyword':
        case 'ProtectedKeyword':
        case 'StaticKeyword':
        case 'ReadonlyKeyword':
            return true;
        default:
            return false;
    }
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
        if(kind(scanner) === 'Identifier' || (kind(scanner).endsWith('Keyword') && !isReservedKind(kind(scanner)))) {
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
            isModifierKind(first) &&
            first !== 'AsyncKeyword' &&
            second !== 'AsKeyword' &&
            (second === 'Identifier' || (second.endsWith('Keyword') && !isReservedKind(second)))
        ) {
            state.restore();
            return true;
        }
        const parameterHead =
            first === 'Identifier' || first === 'ThisKeyword' || (first.endsWith('Keyword') && !isReservedKind(first));
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
        if((kind(scanner) === 'EqualsGreaterThanToken' || kind(scanner) === 'OpenBraceToken') && (scanner.flags & 1) === 0) {
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

// Contextual modifiers remain names unless the next token can be a declaration
// member. Static permits a line break; the other modifiers use same-line lookahead.
export function declarationModifierAhead(scanner: Scanner): boolean {
    if(!declarationModifierStart(kind(scanner))) {
        return false;
    }
    const current = kind(scanner);
    const state = new Speculation(scanner);
    state.next();
    const next = kind(scanner);
    const result =
        (current === 'StaticKeyword' || (scanner.flags & 1) === 0) &&
        (next === 'Identifier' || next.endsWith('Keyword') ||
            ['OpenBracketToken', 'OpenBraceToken', 'AsteriskToken', 'DotDotDotToken',
                'StringLiteral', 'NumericLiteral', 'BigIntLiteral'].includes(next));
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
                    (kind(scanner).endsWith('Keyword') && !isReservedKind(kind(scanner))) ||
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
                        (kind(scanner).endsWith('Keyword') && !isReservedKind(kind(scanner)))) &&
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
        if(declarationModifierStart(current)) {
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

function statementTokenStart(tokenKind: string): boolean {
    switch(tokenKind) {
        case 'AtToken':
        case 'SemicolonToken':
        case 'OpenBraceToken':
        case 'VarKeyword':
        case 'LetKeyword':
        case 'UsingKeyword':
        case 'FunctionKeyword':
        case 'ClassKeyword':
        case 'EnumKeyword':
        case 'IfKeyword':
        case 'DoKeyword':
        case 'WhileKeyword':
        case 'ForKeyword':
        case 'ContinueKeyword':
        case 'BreakKeyword':
        case 'ReturnKeyword':
        case 'WithKeyword':
        case 'SwitchKeyword':
        case 'ThrowKeyword':
        case 'TryKeyword':
        case 'DebuggerKeyword':
        case 'CatchKeyword':
        case 'FinallyKeyword':
        case 'ThisKeyword':
        case 'SuperKeyword':
        case 'NullKeyword':
        case 'TrueKeyword':
        case 'FalseKeyword':
        case 'NumericLiteral':
        case 'BigIntLiteral':
        case 'StringLiteral':
        case 'NoSubstitutionTemplateLiteral':
        case 'TemplateHead':
        case 'OpenParenToken':
        case 'OpenBracketToken':
        case 'NewKeyword':
        case 'SlashToken':
        case 'SlashEqualsToken':
        case 'PlusToken':
        case 'MinusToken':
        case 'TildeToken':
        case 'ExclamationToken':
        case 'DeleteKeyword':
        case 'TypeOfKeyword':
        case 'VoidKeyword':
        case 'PlusPlusToken':
        case 'MinusMinusToken':
        case 'LessThanToken':
        case 'AwaitKeyword':
        case 'YieldKeyword':
        case 'PrivateIdentifier':
            return true;
        default:
            return false;
    }
}

export function statementAhead(scanner: Scanner): boolean {
    const current = kind(scanner);
    if(current === 'Identifier') {
        return true;
    }
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
    if(statementModifierStart(current)) {
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
        statementTokenStart(current) ||
        precedence(current) >= 0 ||
        current === 'Identifier' ||
        (current.endsWith('Keyword') && !isReservedKind(current))
    );
}

export function typeMemberAhead(scanner: Scanner): boolean {
    if(['OpenParenToken', 'LessThanToken', 'GetKeyword', 'SetKeyword'].includes(kind(scanner))) {
        return true;
    }
    const state = new Speculation(scanner);
    let identifier = false;
    while(isModifierKind(kind(scanner))) {
        identifier = true;
        state.next();
    }
    if(kind(scanner) === 'OpenBracketToken') {
        state.restore();
        return true;
    }
    if(propertyNameStart(kind(scanner))) {
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
        (tokenKind.endsWith('Keyword') && !isReservedKind(tokenKind)) ||
        typeKeywordStart(tokenKind)
    );
}

export function modifierAhead(scanner: Scanner, permitConst: boolean): boolean {
    const current = kind(scanner);
    if(!isModifierKind(current)) {
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
            propertyNameStart(kind(scanner)) ||
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

export function nextIdentifierSameLine(scanner: Scanner): boolean {
    const state = new Speculation(scanner);
    scanner.scan();
    const result = bindingStart(kind(scanner)) && (scanner.flags & 1) === 0;
    state.restore();
    return result;
}

export function classMemberAhead(scanner: Scanner): boolean {
    const state = new Speculation(scanner);
    let id = '';
    let result = kind(scanner) === 'AtToken';
    while(!result && isModifierKind(kind(scanner))) {
        id = kind(scanner);
        result = [
            'PublicKeyword', 'PrivateKeyword', 'ProtectedKeyword', 'StaticKeyword',
            'ReadonlyKeyword', 'OverrideKeyword', 'AccessorKeyword',
        ].includes(id);
        state.next();
    }
    if(!result && kind(scanner) === 'AsteriskToken') {
        result = true;
    }
    if(!result && propertyNameStart(kind(scanner))) {
        id = kind(scanner);
        state.next();
    }
    result = result || kind(scanner) === 'OpenBracketToken';
    if(!result && id !== '') {
        result =
            !id.endsWith('Keyword') || id === 'GetKeyword' || id === 'SetKeyword' ||
            [
                'OpenParenToken', 'LessThanToken', 'ExclamationToken', 'ColonToken',
                'EqualsToken', 'QuestionToken', 'SemicolonToken', 'CloseBraceToken', 'EndOfFile',
            ].includes(kind(scanner)) || (scanner.flags & 1) !== 0;
    }
    state.restore();
    return result;
}

export function accessorAhead(scanner: Scanner): boolean {
    const state = new Speculation(scanner);
    state.next();
    const next = kind(scanner);
    const result = next === 'OpenBracketToken' || propertyNameStart(next);
    state.restore();
    return result;
}

// 0 rules out an arrow, 1 commits to recovery, 2 requires a parsed signature.
export function arrowHeadKind(scanner: Scanner): number {
    if(
        kind(scanner) !== 'EqualsGreaterThanToken' && kind(scanner) !== 'OpenParenToken' &&
        kind(scanner) !== 'LessThanToken' && kind(scanner) !== 'AsyncKeyword'
    ) {
        return 0;
    }
    const state = new Speculation(scanner);
    if(kind(scanner) === 'EqualsGreaterThanToken') {
        return 1;
    }
    if(kind(scanner) === 'AsyncKeyword') {
        state.next();
        if((scanner.flags & 1) !== 0) {
            state.restore();
            return 0;
        }
    }
    const first = kind(scanner);
    state.next();
    const second = kind(scanner);
    let result = 0;
    if(first === 'LessThanToken') {
        result = bindingStart(second) || second === 'ConstKeyword' ? 2 : 0;
    }
    else if(first === 'OpenParenToken') {
        if(second === 'CloseParenToken') {
            state.next();
            result = ['EqualsGreaterThanToken', 'ColonToken', 'OpenBraceToken'].includes(kind(scanner)) ? 1 : 0;
        }
        else if(second === 'DotDotDotToken') {
            result = 1;
        }
        else if(second === 'OpenBracketToken' || second === 'OpenBraceToken') {
            result = 2;
        }
        else {
            state.next();
            const third = kind(scanner);
            if(isModifierKind(second) && second !== 'AsyncKeyword' && bindingStart(third) && third !== 'AsKeyword') {
                result = 1;
            }
            else if(bindingStart(second) || second === 'ThisKeyword') {
                if(third === 'ColonToken') {
                    result = 1;
                }
                else if(third === 'QuestionToken') {
                    state.next();
                    result = ['ColonToken', 'CommaToken', 'EqualsToken', 'CloseParenToken'].includes(kind(scanner)) ? 1 : 0;
                }
                else if(['CommaToken', 'EqualsToken', 'CloseParenToken'].includes(third)) {
                    result = 2;
                }
            }
        }
    }
    state.restore();
    return result;
}
