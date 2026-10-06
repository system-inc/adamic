import type { Batch5Context } from './batch5_context.ts';
export function first(context: Batch5Context, index: number): number {
    return context.node(index).children[0] ?? -1;
}
export function last(context: Batch5Context, index: number): number {
    const children = context.node(index).children;
    return children[children.length - 1] ?? -1;
}
export function typescriptFile(path: string): boolean {
    return ['.ts', '.tsx', '.mts', '.cts'].some((extension) => path.endsWith(extension));
}
export function name(context: Batch5Context, index: number): number {
    for(const child of context.node(index).children) {
        if(
            [
                'Identifier',
                'StringLiteral',
                'NumericLiteral',
                'PrivateIdentifier',
                'ComputedPropertyName',
                'ObjectBindingPattern',
                'ArrayBindingPattern',
            ].includes(context.node(child).kind)
        ) {
            return child;
        }
    }
    return -1;
}
export function initializer(context: Batch5Context, index: number): number {
    const children = context.node(index).children;
    for(let cursor = 1; cursor < children.length; cursor++) {
        const child = children[cursor] ?? -1;
        const previous = children[cursor - 1] ?? -1;
        context.scanAt(context.node(previous).end);
        if(context.scanner.kind === 'EqualsToken') {
            return child;
        }
    }
    return -1;
}
export function returnType(context: Batch5Context, index: number): number {
    let result = -1;
    for(const child of context.node(index).children) {
        if(
            !['Parameter', 'TypeParameter', 'Block', 'Identifier', 'QuestionToken'].includes(
                context.node(child).kind,
            ) &&
            !context.node(child).kind.endsWith('Keyword')
        ) {
            result = child;
        }
        else if(
            [
                'AnyKeyword',
                'VoidKeyword',
                'StringKeyword',
                'NumberKeyword',
                'BooleanKeyword',
                'UnknownKeyword',
                'NeverKeyword',
                'ObjectKeyword',
                'UndefinedKeyword',
                'ThisType',
            ].includes(context.node(child).kind)
        ) {
            result = child;
        }
    }
    return result;
}
export function assignmentOperator(kind: string): boolean {
    return [
        'EqualsToken',
        'PlusEqualsToken',
        'MinusEqualsToken',
        'AsteriskEqualsToken',
        'AsteriskAsteriskEqualsToken',
        'SlashEqualsToken',
        'PercentEqualsToken',
        'LessThanLessThanEqualsToken',
        'GreaterThanGreaterThanEqualsToken',
        'GreaterThanGreaterThanGreaterThanEqualsToken',
        'AmpersandEqualsToken',
        'BarEqualsToken',
        'CaretEqualsToken',
        'BarBarEqualsToken',
        'AmpersandAmpersandEqualsToken',
        'QuestionQuestionEqualsToken',
    ].includes(kind);
}
