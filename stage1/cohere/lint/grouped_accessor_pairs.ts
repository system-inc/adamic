// Port of pinned Go cohere's grouped-accessor-pairs.
import type { RuleContext } from './rule_context.ts';
import { child, kind, key, range } from './batch4_helpers.ts';
function staticName(ctx: RuleContext, index: number): string {
    let name = index;
    if(kind(ctx, name) === 'ComputedPropertyName') {
        name = ctx.unwrap(child(ctx, name, 0));
        if(!['StringLiteral', 'NoSubstitutionTemplateLiteral', 'NumericLiteral'].includes(kind(ctx, name))) return '';
    }
    if(
        [
            'Identifier',
            'PrivateIdentifier',
            'StringLiteral',
            'NumericLiteral',
            'NoSubstitutionTemplateLiteral',
        ].includes(kind(ctx, name))
    )
        return `static:${ctx.node(name).text}`;
    return '';
}
function label(ctx: RuleContext, index: number): string {
    const name = key(ctx, index);
    const prefix = ctx.hasKind(index, 'StaticKeyword') ? 'static ' : '';
    const accessor = kind(ctx, index) === 'GetAccessor' ? 'getter' : 'setter';
    if(kind(ctx, name) === 'PrivateIdentifier') return `${prefix}private ${accessor} ${ctx.node(name).text}`;
    const value = staticName(ctx, name);
    return prefix + accessor + (value === '' ? '' : ` '${value.slice(7)}'`);
}
function check(ctx: RuleContext, members: number[], filter: number, rule: string): void {
    const keys: string[] = [];
    const gets: number[] = [];
    const sets: number[] = [];
    const getCounts: number[] = [];
    const setCounts: number[] = [];
    for(let i = 0; i < members.length; i++) {
        const member = members[i] ?? -1;
        if(
            !['GetAccessor', 'SetAccessor'].includes(kind(ctx, member)) ||
            (filter >= 0 && ctx.hasKind(member, 'StaticKeyword') !== (filter === 1))
        )
            continue;
        const name = key(ctx, member);
        if(name < 0) continue;
        let value = staticName(ctx, name);
        if(kind(ctx, name) === 'PrivateIdentifier') value = `private:${ctx.node(name).text}`;
        if(value === '') value = `computed:${ctx.source.slice(ctx.start(name), ctx.node(name).end)}`;
        let entry = keys.indexOf(value);
        if(entry < 0) {
            entry = keys.length;
            keys.push(value);
            gets.push(-1);
            sets.push(-1);
            getCounts.push(0);
            setCounts.push(0);
        }
        if(kind(ctx, member) === 'GetAccessor') {
            gets[entry] = i;
            getCounts[entry] = (getCounts[entry] ?? 0) + 1;
        }
        else {
            sets[entry] = i;
            setCounts[entry] = (setCounts[entry] ?? 0) + 1;
        }
    }
    for(let i = 0; i < keys.length; i++) {
        if(getCounts[i] !== 1 || setCounts[i] !== 1) continue;
        const get = gets[i] ?? -1;
        const set = sets[i] ?? -1;
        const former = members[Math.min(get, set)] ?? -1;
        const latter = members[Math.max(get, set)] ?? -1;
        const order = ctx.settings.read('order', 'anyOrder');
        let id = '';
        let message = '';
        if(Math.abs(get - set) > 1) {
            id = 'notGrouped';
            message = `Accessor pair ${label(ctx, former)} and ${label(ctx, latter)} should be grouped.`;
        }
        else if((order === 'getBeforeSet' && get > set) || (order === 'setBeforeGet' && set > get)) {
            id = 'invalidOrder';
            message = `Expected ${label(ctx, latter)} to be before ${label(ctx, former)}.`;
        }
        if(id !== '') range(ctx, rule, id, message, ctx.start(latter), ctx.node(key(ctx, latter)).end);
    }
}
export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('grouped-accessor-pairs')) return;
    const rule = 'grouped-accessor-pairs';
    const nodeKind = kind(ctx, index);
    if(
        !['ObjectLiteralExpression', 'ClassDeclaration', 'ClassExpression'].includes(nodeKind) &&
        !(
            ctx.settings.read('enforcefortstypes', 'false') === 'true' &&
            ['InterfaceDeclaration', 'TypeLiteral'].includes(nodeKind)
        )
    )
        return;
    const members = ctx
        .node(index)
        .children.filter(
            (part) =>
                ![
                    'Identifier',
                    'TypeParameter',
                    'HeritageClause',
                    'ExportKeyword',
                    'DefaultKeyword',
                    'DeclareKeyword',
                    'AbstractKeyword',
                    'Decorator',
                ].includes(kind(ctx, part)),
        );
    const classBody = nodeKind === 'ClassDeclaration' || nodeKind === 'ClassExpression';
    check(ctx, members, classBody ? 0 : -1, rule);
    if(classBody) check(ctx, members, 1, rule);
}
