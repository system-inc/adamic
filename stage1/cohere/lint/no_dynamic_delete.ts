import type { Batch5Context } from './batch5_context.ts';
import { first, last } from './batch5_helpers.ts';
import { ruleMessage } from './batch5_messages.ts';
export function visit(context: Batch5Context, index: number): void {
    const rule = '@typescript-eslint/no-dynamic-delete';
    if(context.node(index).kind !== 'DeleteExpression' || !context.enabled(rule)) {
        return;
    }
    const operand = context.unwrap(first(context, index));
    if(context.node(operand).kind !== 'ElementAccessExpression' || context.hasKind(operand, 'QuestionDotToken')) {
        return;
    }
    const key = context.unwrap(last(context, operand));
    const kind = context.node(key).kind;
    if(['StringLiteral', 'NumericLiteral'].includes(kind)) {
        return;
    }
    if(
        kind === 'PrefixUnaryExpression' &&
        context.node(key).operator === 'MinusToken' &&
        context.node(context.unwrap(first(context, key))).kind === 'NumericLiteral'
    ) {
        return;
    }
    context.report(key, rule, 'dynamicDelete', ruleMessage('dynamicDelete'), [], []);
}
