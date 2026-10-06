import type { Batch5Context } from './batch5_context.ts';
import { first } from './batch5_helpers.ts';
import { ruleMessage } from './batch5_messages.ts';
import { RepairEdit } from './repair_edit.ts';
export function visit(context: Batch5Context, index: number): void {
    const rule = '@typescript-eslint/no-extra-non-null-assertion';
    if(context.node(index).kind !== 'NonNullExpression' || !context.enabled(rule)) {
        return;
    }
    let outer = index;
    let owner = context.parent(outer);
    while(owner >= 0 && context.node(owner).kind === 'ParenthesizedExpression') {
        outer = owner;
        owner = context.parent(outer);
    }
    if(owner < 0) {
        return;
    }
    const kind = context.node(owner).kind;
    if(
        kind !== 'NonNullExpression' &&
        !(
            ['PropertyAccessExpression', 'ElementAccessExpression', 'CallExpression'].includes(kind) &&
            context.hasKind(owner, 'QuestionDotToken') &&
            first(context, owner) === outer
        )
    ) {
        return;
    }
    const end = context.node(index).end;
    context.report(
        index,
        rule,
        'noExtraNonNullAssertion',
        ruleMessage('noExtraNonNullAssertion'),
        [new RepairEdit(end - 1, end, '')],
        [],
    );
}
