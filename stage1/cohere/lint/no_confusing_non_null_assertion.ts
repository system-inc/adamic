import type { RuleContext } from './rule_context.ts';
import { ExtraEdit } from './extra_edit.ts';
import { RepairSuggestion } from './repair_suggestion.ts';
export function visit(ctx: RuleContext, index: number): void {
    if(
        !ctx.enabled('@typescript-eslint/no-confusing-non-null-assertion') ||
        ctx.node(index).kind !== 'BinaryExpression'
    )
        return;
    const children = ctx.node(index).children;
    const left = children[0] ?? -1;
    const operator = children[1] ?? -1;
    if(left < 0 || operator < 0) return;
    const text = ctx.text(operator);
    if(!['=', '==', '===', 'in', 'instanceof'].includes(text) || !ctx.text(left).endsWith('!')) return;
    const start = ctx.start(left);
    const end = ctx.node(left).end;
    const wrap = new RepairSuggestion(
        'wrapUpLeft',
        `Wrap the left-hand side in parentheses to avoid confusion with "${text}" operator.`,
        [new ExtraEdit(start, start, '('), new ExtraEdit(end, end, ')')],
    );
    const assign = text === '=';
    const equal = text === '==' || text === '===';
    const id = assign ? 'confusingAssign' : equal ? 'confusingEqual' : 'confusingOperator';
    const message = assign
        ? 'Confusing combination of non-null assertion and assignment like `a! = b`, which looks very similar to `a != b`.'
        : equal
          ? 'Confusing combination of non-null assertion and equality test like `a! == b`, which looks very similar to `a !== b`.'
          : `Confusing combination of non-null assertion and \`${text}\` operator like \`a! ${text} b\`, which might be misinterpreted as \`!(a ${text} b)\`.`;
    let first = wrap;
    if(ctx.node(left).kind === 'NonNullExpression') {
        const suggestionId = assign ? 'notNeedInAssign' : equal ? 'notNeedInEqualTest' : 'notNeedInOperator';
        const description = assign
            ? 'Remove unnecessary non-null assertion (!) in assignment left-hand side.'
            : equal
              ? 'Remove unnecessary non-null assertion (!) in equality test.'
              : `Remove possibly unnecessary non-null assertion (!) in the left operand of the \`${text}\` operator.`;
        first = new RepairSuggestion(suggestionId, description, [new ExtraEdit(end - 1, end, '')]);
    }
    const edit = first.edits[0] ?? new ExtraEdit(start, start, '');
    const finding = ctx.report(
        index,
        '@typescript-eslint/no-confusing-non-null-assertion',
        id,
        message,
        'suggestion',
        edit.text,
        first.message,
    );
    finding.editStart = edit.start;
    finding.editEnd = edit.end;
    finding.suggestions.push(first);
    if(ctx.node(left).kind === 'NonNullExpression' && !assign && !equal) finding.suggestions.push(wrap);
}
