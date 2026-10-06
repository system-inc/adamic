// Port of pinned Go cohere's base/correctness-require-orm-column-declare.
import type { RuleContext } from './rule_context.ts';
import { text, child, kind, key } from './batch4_helpers.ts';

export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('base/correctness-require-orm-column-declare')) return;
    const rule = 'base/correctness-require-orm-column-declare';
    if(kind(ctx, index) !== 'PropertyDeclaration' || ctx.hasKind(index, 'DeclareKeyword')) return;
    let decoratorName = '';
    for(const part of ctx.node(index).children) {
        if(kind(ctx, part) !== 'Decorator') continue;
        let expression = child(ctx, part, 0);
        if(kind(ctx, expression) === 'CallExpression') expression = child(ctx, expression, 0);
        if(
            kind(ctx, expression) === 'Identifier' &&
            [
                'OrmColumn',
                'OrmPrimaryKey',
                'OrmPrimaryAutoColumn',
                'OrmCreateDateColumn',
                'OrmUpdateDateColumn',
                'OrmColumnIndex',
                'OrmColumnUnique',
                'OrmColumnUniqueIndex',
            ].includes(ctx.node(expression).text)
        ) {
            decoratorName = ctx.node(expression).text;
            break;
        }
    }
    if(decoratorName === '') return;
    let name = key(ctx, index);
    if(kind(ctx, name) === 'ComputedPropertyName') name = child(ctx, name, 0);
    if(name < 0) return;
    const propertyName = kind(ctx, name) === 'Identifier' ? ctx.node(name).text : '<computed>';
    ctx.report(
        name,
        rule,
        'missingDeclare',
        text('base_correctness_require_orm_column_declare:missingDeclare')
            .replaceAll('{{propertyName}}', propertyName)
            .replaceAll('{{decoratorName}}', decoratorName),
    );
}
