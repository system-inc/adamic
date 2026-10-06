import type { Batch5Context } from './batch5_context.ts';
import { first } from './batch5_helpers.ts';
import { RepairEdit } from './repair_edit.ts';
import { RepairSuggestion } from './repair_suggestion.ts';
export function visit(context: Batch5Context, index: number): void {
    const rule = '@typescript-eslint/no-confusing-non-null-assertion';
    if(context.node(index).kind !== 'BinaryExpression' || !context.enabled(rule)) {
        return;
    }
    const operator = context.node(context.node(index).children[1] ?? -1).kind;
    const operators = ['EqualsToken', 'EqualsEqualsToken', 'EqualsEqualsEqualsToken', 'InKeyword', 'InstanceOfKeyword'];
    const spelling = ['=', '==', '===', 'in', 'instanceof'];
    const slot = operators.indexOf(operator);
    if(slot < 0) {
        return;
    }
    const text = spelling[slot] ?? '';
    const left = first(context, index);
    if(!context.text(left).endsWith('!')) {
        return;
    }
    const start = context.start(left);
    const end = context.node(left).end;
    const suggestions: RepairSuggestion[] = [];
    const wrap = new RepairSuggestion(
        'wrapUpLeft',
        `Wrap the left-hand side in parentheses to avoid confusion with "${text}" operator.`,
        [new RepairEdit(start, start, '('), new RepairEdit(end, end, ')')],
    );
    let id: string;
    let message: string;
    if(text === '=') {
        id = 'confusingAssign';
        message =
            'Confusing combination of non-null assertion and assignment like `a! = b`, which looks very similar to `a != b`.';
    }
    else if(text === '==' || text === '===') {
        id = 'confusingEqual';
        message =
            'Confusing combination of non-null assertion and equality test like `a! == b`, which looks very similar to `a !== b`.';
    }
    else {
        id = 'confusingOperator';
        message = `Confusing combination of non-null assertion and \`${text}\` operator like \`a! ${text} b\`, which might be misinterpreted as \`!(a ${text} b)\`.`;
    }
    if(context.node(left).kind !== 'NonNullExpression') {
        suggestions.push(wrap);
    }
    else if(text === '=') {
        suggestions.push(
            new RepairSuggestion(
                'notNeedInAssign',
                'Remove unnecessary non-null assertion (!) in assignment left-hand side.',
                [new RepairEdit(end - 1, end, '')],
            ),
        );
    }
    else if(text === '==' || text === '===') {
        suggestions.push(
            new RepairSuggestion('notNeedInEqualTest', 'Remove unnecessary non-null assertion (!) in equality test.', [
                new RepairEdit(end - 1, end, ''),
            ]),
        );
    }
    else {
        suggestions.push(
            new RepairSuggestion(
                'notNeedInOperator',
                `Remove possibly unnecessary non-null assertion (!) in the left operand of the \`${text}\` operator.`,
                [new RepairEdit(end - 1, end, '')],
            ),
        );
        suggestions.push(wrap);
    }
    context.report(index, rule, id, message, [], suggestions);
}
