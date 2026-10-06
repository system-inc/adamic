// Port of pinned Go cohere's for-direction.
import type { RuleContext } from './rule_context.ts';
import { text, child, last, kind, named } from './batch4_helpers.ts';
function sign(ctx: RuleContext, index: number): number {
    const current = ctx.unwrap(index);
    if(current < 0) return 0;
    if(['NumericLiteral', 'BigIntLiteral'].includes(kind(ctx, current))) {
        for(const character of ctx.node(current).text) if(character !== '0' && character !== '.') return 1;
        return 0;
    }
    if(kind(ctx, current) === 'TrueKeyword') return 1;
    if(kind(ctx, current) === 'PrefixUnaryExpression') {
        if(ctx.node(current).operator === 'PlusToken') return sign(ctx, child(ctx, current, 0));
        if(ctx.node(current).operator === 'MinusToken') return -sign(ctx, child(ctx, current, 0));
    }
    return 0;
}
function collect(ctx: RuleContext, index: number, name: string, result: number[]): void {
    const current = ctx.unwrap(index);
    if(current < 0) return;
    if(
        ['PrefixUnaryExpression', 'PostfixUnaryExpression'].includes(kind(ctx, current)) &&
        ['PlusPlusToken', 'MinusMinusToken'].includes(ctx.node(current).operator) &&
        named(ctx, ctx.unwrap(child(ctx, current, 0)), name)
    )
        result.push(current);
    if(kind(ctx, current) === 'BinaryExpression') {
        const operator = kind(ctx, child(ctx, current, 1));
        if(operator === 'CommaToken') {
            collect(ctx, child(ctx, current, 0), name, result);
            collect(ctx, child(ctx, current, 2), name, result);
        }
        else if(ctx.assignment(current) && named(ctx, ctx.unwrap(child(ctx, current, 0)), name)) result.push(current);
    }
}
function direction(ctx: RuleContext, index: number): number {
    if(kind(ctx, index) === 'BinaryExpression') {
        const operator = kind(ctx, child(ctx, index, 1));
        if(operator === 'PlusEqualsToken') return sign(ctx, child(ctx, index, 2));
        if(operator === 'MinusEqualsToken') return -sign(ctx, child(ctx, index, 2));
        return 0;
    }
    return ctx.node(index).operator === 'PlusPlusToken' ? 1 : ctx.node(index).operator === 'MinusMinusToken' ? -1 : 0;
}
export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('for-direction')) return;
    const rule = 'for-direction';
    if(kind(ctx, index) !== 'ForStatement') return;
    const parts = ctx.node(index).children;
    if(parts.length < 3) return;
    const body = last(ctx, index);
    let condition = -1;
    let update = -1;
    let semicolons = 0;
    let cursor = ctx.node(index).pos;
    for(const part of parts) {
        if(part === body) break;
        const position = ctx.start(part);
        ctx.scanner.pos = cursor;
        while(ctx.scanner.scan() !== 'EndOfFile' && ctx.scanStart() < position) {
            if(ctx.scanner.kind === 'SemicolonToken') semicolons++;
        }
        if(semicolons === 1) condition = ctx.unwrap(part);
        if(semicolons === 2) update = part;
        cursor = ctx.node(part).end;
    }
    if(kind(ctx, condition) !== 'BinaryExpression' || update < 0) return;
    const comparison = ctx.comparison(condition);
    if(!['<', '<=', '>', '>='].includes(comparison)) return;
    for(let side = 0; side < 2; side++) {
        const operand = ctx.unwrap(child(ctx, condition, side === 0 ? 0 : 2));
        if(kind(ctx, operand) !== 'Identifier') continue;
        const modifications: number[] = [];
        collect(ctx, update, ctx.node(operand).text, modifications);
        if(modifications.length !== 1) continue;
        const change = modifications[0] ?? -1;
        const wrong = comparison === '<' || comparison === '<=' ? (side === 0 ? -1 : 1) : side === 0 ? 1 : -1;
        if(direction(ctx, change) === wrong) {
            ctx.report(index, rule, 'incorrectDirection', text('for_direction'));
            return;
        }
    }
}
