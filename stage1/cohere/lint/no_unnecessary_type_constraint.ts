// Port of pinned Go cohere's @typescript-eslint/no-unnecessary-type-constraint.
import type { RuleContext } from './rule_context.ts';

const messageNoUnnecessaryTypeConstraint =
    'A type parameter with no `extends` clause is already constrained to `unknown`, so writing `extends any` or `extends unknown` constrains nothing. The clause reads as a deliberate bound and communicates one, which makes it worse than noise: a reader looking for where `T` is restricted finds this and stops. Remove it.';
const suggestionRemoveTheConstraint = 'Remove the constraint that does nothing.';

export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('@typescript-eslint/no-unnecessary-type-constraint')) return;
    const node = ctx.node(index);
    if(node.kind !== 'TypeParameter') return;
    const name = node.children.find((child) => ctx.node(child).kind === 'Identifier') ?? -1;
    if(name < 0) return;
    const position = node.children.indexOf(name);
    const constraint = node.children[position + 1] ?? -1;
    if(constraint < 0 || !['AnyKeyword', 'UnknownKeyword'].includes(ctx.node(constraint).kind)) return;
    if(ctx.scanAt(ctx.node(name).end) !== 'ExtendsKeyword') return;
    let replacement = '';
    const parent = ctx.parent(index);
    if(
        parent >= 0 &&
        ctx.node(parent).kind === 'ArrowFunction' &&
        node.children.length === position + 2 &&
        ctx.node(parent).children.filter((child) => ctx.node(child).kind === 'TypeParameter').length === 1 &&
        ['.tsx', '.mts', '.cts'].some((suffix) => ctx.parser.path.endsWith(suffix))
    ) {
        if(ctx.scanAt(node.end) !== 'CommaToken') replacement = ',';
    }
    const finding = ctx.report(
        name,
        '@typescript-eslint/no-unnecessary-type-constraint',
        'noUnnecessaryTypeConstraint',
        messageNoUnnecessaryTypeConstraint,
        'suggestion',
        replacement,
        suggestionRemoveTheConstraint,
    );
    finding.editStart = ctx.node(name).end;
    finding.editEnd = ctx.node(constraint).end;
}
