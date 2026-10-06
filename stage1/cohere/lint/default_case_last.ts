import type { RuleContext } from './rule_context.ts';
const message =
    'This `default` clause is not the last one in its `switch`. A reader scanning the clauses in order expects the fallback at the bottom, and a `default` sitting in the middle also falls through into whatever case follows it when its body has no `break`, which is a control flow almost nobody writes on purpose. Move it to the end.';
export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('default-case-last') || ctx.node(index).kind !== 'SwitchStatement') return;
    const block = ctx.child(index, 'CaseBlock');
    if(block < 0) return;
    const clauses = ctx.node(block).children;
    for(let position = 0; position < clauses.length; position++) {
        const clause = clauses[position] ?? -1;
        if(clause < 0 || ctx.node(clause).kind !== 'DefaultClause') continue;
        if(position !== clauses.length - 1) ctx.report(clause, 'default-case-last', 'notLast', message);
        return;
    }
}
