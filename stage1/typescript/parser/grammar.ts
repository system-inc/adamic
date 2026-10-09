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

export function isReservedKind(kind: string): boolean {
    switch(kind) {
        case 'BreakKeyword':
        case 'CaseKeyword':
        case 'CatchKeyword':
        case 'ClassKeyword':
        case 'ConstKeyword':
        case 'ContinueKeyword':
        case 'DebuggerKeyword':
        case 'DefaultKeyword':
        case 'DeleteKeyword':
        case 'DoKeyword':
        case 'ElseKeyword':
        case 'EnumKeyword':
        case 'ExportKeyword':
        case 'ExtendsKeyword':
        case 'FalseKeyword':
        case 'FinallyKeyword':
        case 'ForKeyword':
        case 'FunctionKeyword':
        case 'IfKeyword':
        case 'ImportKeyword':
        case 'InKeyword':
        case 'InstanceOfKeyword':
        case 'NewKeyword':
        case 'NullKeyword':
        case 'ReturnKeyword':
        case 'SuperKeyword':
        case 'SwitchKeyword':
        case 'ThisKeyword':
        case 'ThrowKeyword':
        case 'TrueKeyword':
        case 'TryKeyword':
        case 'TypeOfKeyword':
        case 'VarKeyword':
        case 'VoidKeyword':
        case 'WhileKeyword':
        case 'WithKeyword':
            return true;
        default:
            return false;
    }
}

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

export function isModifierKind(kind: string): boolean {
    switch(kind) {
        case 'AbstractKeyword':
        case 'AccessorKeyword':
        case 'AsyncKeyword':
        case 'ConstKeyword':
        case 'DeclareKeyword':
        case 'DefaultKeyword':
        case 'ExportKeyword':
        case 'InKeyword':
        case 'PrivateKeyword':
        case 'ProtectedKeyword':
        case 'PublicKeyword':
        case 'ReadonlyKeyword':
        case 'OutKeyword':
        case 'OverrideKeyword':
        case 'StaticKeyword':
            return true;
        default:
            return false;
    }
}

export function isExpressionKind(kind: string): boolean {
    switch(kind) {
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
        case 'OpenBraceToken':
        case 'FunctionKeyword':
        case 'ClassKeyword':
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
        case 'AtToken':
            return true;
        default:
            return false;
    }
}

// Property-name and parameter token inventories are shared by list recovery.
export function propertyNameStart(kind: string): boolean {
    return (
        kind === 'Identifier' ||
        kind === 'PrivateIdentifier' ||
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
    return kind === 'Identifier' || (kind.endsWith('Keyword') && !isReservedKind(kind));
}

export function expressionTokenStart(kind: string, disallowIn: boolean): boolean {
    return (
        bindingStart(kind) || (precedence(kind) >= 0 && !(kind === 'InKeyword' && disallowIn)) || isExpressionKind(kind)
    );
}

export function parameterTokenStart(kind: string): boolean {
    return isModifierKind(kind) || parameterKinds.includes(kind);
}

export function leftHandSideKind(kind: string): boolean {
    switch(kind) {
        case 'PropertyAccessExpression':
        case 'ElementAccessExpression':
        case 'NewExpression':
        case 'CallExpression':
        case 'JsxElement':
        case 'JsxSelfClosingElement':
        case 'JsxFragment':
        case 'TaggedTemplateExpression':
        case 'ArrayLiteralExpression':
        case 'ParenthesizedExpression':
        case 'ObjectLiteralExpression':
        case 'ClassExpression':
        case 'FunctionExpression':
        case 'Identifier':
        case 'PrivateIdentifier':
        case 'RegularExpressionLiteral':
        case 'NumericLiteral':
        case 'BigIntLiteral':
        case 'StringLiteral':
        case 'NoSubstitutionTemplateLiteral':
        case 'TemplateExpression':
        case 'FalseKeyword':
        case 'NullKeyword':
        case 'ThisKeyword':
        case 'TrueKeyword':
        case 'SuperKeyword':
        case 'NonNullExpression':
        case 'ExpressionWithTypeArguments':
        case 'MetaProperty':
        case 'ImportKeyword':
        case 'MissingDeclaration':
            return true;
        default:
            return false;
    }
}

export function heritageTokenStart(kind: string): boolean {
    return kind !== 'ImplementsKeyword' && (bindingStart(kind) || [
        'ThisKeyword', 'SuperKeyword', 'NullKeyword', 'TrueKeyword', 'FalseKeyword',
        'NumericLiteral', 'BigIntLiteral', 'StringLiteral', 'NoSubstitutionTemplateLiteral',
        'TemplateHead', 'OpenParenToken', 'OpenBracketToken', 'FunctionKeyword',
        'ClassKeyword', 'NewKeyword', 'SlashToken', 'SlashEqualsToken',
    ].includes(kind));
}
