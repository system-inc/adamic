// Operator precedence and reserved words follow typescript-go.
import { panic } from 'adamic';
export function precedence(kind: string): number {
    switch(kind) {
        case 'QuestionQuestionToken':
            return 5;
        case 'BarBarToken':
            return 5;
        case 'AmpersandAmpersandToken':
            return 6;
        case 'BarToken':
            return 7;
        case 'CaretToken':
            return 8;
        case 'AmpersandToken':
            return 9;
        case 'EqualsEqualsToken':
        case 'ExclamationEqualsToken':
        case 'EqualsEqualsEqualsToken':
        case 'ExclamationEqualsEqualsToken':
            return 10;
        case 'LessThanToken':
        case 'GreaterThanToken':
        case 'LessThanEqualsToken':
        case 'GreaterThanEqualsToken':
        case 'InstanceOfKeyword':
        case 'InKeyword':
        case 'AsKeyword':
        case 'SatisfiesKeyword':
            return 11;
        case 'LessThanLessThanToken':
        case 'GreaterThanGreaterThanToken':
        case 'GreaterThanGreaterThanGreaterThanToken':
            return 12;
        case 'PlusToken':
        case 'MinusToken':
            return 13;
        case 'AsteriskToken':
        case 'SlashToken':
        case 'PercentToken':
            return 14;
        case 'AsteriskAsteriskToken':
            return 15;
        default:
            return -1;
    }
}

export const reservedKinds: readonly string[] = [
    'BreakKeyword',
    'CaseKeyword',
    'CatchKeyword',
    'ClassKeyword',
    'ConstKeyword',
    'ContinueKeyword',
    'DebuggerKeyword',
    'DefaultKeyword',
    'DeleteKeyword',
    'DoKeyword',
    'ElseKeyword',
    'EnumKeyword',
    'ExportKeyword',
    'ExtendsKeyword',
    'FalseKeyword',
    'FinallyKeyword',
    'ForKeyword',
    'FunctionKeyword',
    'IfKeyword',
    'ImportKeyword',
    'InKeyword',
    'InstanceOfKeyword',
    'NewKeyword',
    'NullKeyword',
    'ReturnKeyword',
    'SuperKeyword',
    'SwitchKeyword',
    'ThisKeyword',
    'ThrowKeyword',
    'TrueKeyword',
    'TryKeyword',
    'TypeOfKeyword',
    'VarKeyword',
    'VoidKeyword',
    'WhileKeyword',
    'WithKeyword',
];

// Spellings used by parseExpected, following Go's scanner.TokenToString.
export function tokenSpelling(kind: string): string {
    if(kind.endsWith('Keyword')) {
        return kind.slice(0, -7).toLowerCase();
    }
    switch(kind) {
        case 'OpenBraceToken':
            return '{';
        case 'CloseBraceToken':
            return '}';
        case 'OpenBracketToken':
            return '[';
        case 'CloseBracketToken':
            return ']';
        case 'OpenParenToken':
            return '(';
        case 'CloseParenToken':
            return ')';
        case 'LessThanToken':
            return '<';
        case 'GreaterThanToken':
            return '>';
        case 'ColonToken':
            return ':';
        case 'SemicolonToken':
            return ';';
        case 'CommaToken':
            return ',';
        case 'QuestionToken':
            return '?';
        case 'EqualsToken':
            return '=';
        case 'EqualsGreaterThanToken':
            return '=>';
        case 'AtToken':
            return '@';
        default:
            return panic('missing expected-token spelling');
    }
}

export const modifierKinds: readonly string[] = [
    'AbstractKeyword',
    'AccessorKeyword',
    'AsyncKeyword',
    'ConstKeyword',
    'DeclareKeyword',
    'DefaultKeyword',
    'ExportKeyword',
    'InKeyword',
    'PrivateKeyword',
    'ProtectedKeyword',
    'PublicKeyword',
    'ReadonlyKeyword',
    'OutKeyword',
    'OverrideKeyword',
    'StaticKeyword',
];

export const expressionKinds: readonly string[] = [
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
    'AtToken',
];

// Property-name and parameter token inventories are shared by list recovery.
export function propertyNameStart(kind: string): boolean {
    return (
        kind === 'Identifier' ||
        kind.endsWith('Keyword') ||
        ['StringLiteral', 'NumericLiteral', 'BigIntLiteral'].includes(kind)
    );
}

const parameterKinds: readonly string[] = [
    'DotDotDotToken',
    'OpenBraceToken',
    'OpenBracketToken',
    'PrivateIdentifier',
    'ThisKeyword',
    'AtToken',
];

export function bindingStart(kind: string): boolean {
    return kind === 'Identifier' || (kind.endsWith('Keyword') && !reservedKinds.includes(kind));
}

export function expressionTokenStart(kind: string, disallowIn: boolean): boolean {
    return (
        bindingStart(kind) ||
        (precedence(kind) >= 0 && !(kind === 'InKeyword' && disallowIn)) ||
        expressionKinds.includes(kind)
    );
}

export function parameterTokenStart(kind: string): boolean {
    return modifierKinds.includes(kind) || parameterKinds.includes(kind);
}
