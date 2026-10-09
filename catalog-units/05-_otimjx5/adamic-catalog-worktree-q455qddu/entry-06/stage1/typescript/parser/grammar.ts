// Operator precedence and reserved words follow typescript-go.
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
