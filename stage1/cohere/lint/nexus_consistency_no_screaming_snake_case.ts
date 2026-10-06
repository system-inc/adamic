// Port of pinned Go cohere's nexus/consistency-no-screaming-snake-case.
import type { RuleContext } from './rule_context.ts';
import { child, kind, named, text } from './batch4_helpers.ts';
function shouting(name: string): boolean {
    if(!name.includes('_') || name.endsWith('_')) return false;
    const first = name[0] ?? '';
    if(first < 'A' || first > 'Z') return false;
    for(const character of name)
        if(character !== '_' && !(character >= 'A' && character <= 'Z') && !(character >= '0' && character <= '9'))
            return false;
    return true;
}
function environment(ctx: RuleContext, index: number): boolean {
    let current = index;
    if(
        kind(ctx, current) === 'BinaryExpression' &&
        ['QuestionQuestionToken', 'BarBarToken'].includes(kind(ctx, child(ctx, current, 1)))
    )
        current = child(ctx, current, 0);
    if(kind(ctx, current) !== 'PropertyAccessExpression') return false;
    let saw = false;
    while(kind(ctx, current) === 'PropertyAccessExpression') {
        if(ctx.accessName(current) === 'env') saw = true;
        current = child(ctx, current, 0);
    }
    return saw && named(ctx, current, 'process');
}
function convert(name: string, exported: boolean): string {
    const parts = name.split('_').filter((part) => part !== '');
    let result = '';
    for(let i = 0; i < parts.length; i++) {
        const part = parts[i] ?? '';
        result += i === 0 && !exported ? part.toLowerCase() : (part[0] ?? '') + part.slice(1).toLowerCase();
    }
    return result;
}
export function visit(ctx: RuleContext, index: number): void {
    if(!ctx.enabled('nexus/consistency-no-screaming-snake-case') || kind(ctx, index) !== 'VariableDeclaration') return;
    const name = child(ctx, index, 0);
    if(kind(ctx, name) !== 'Identifier') return;
    const spelling = ctx.node(name).text;
    if(
        !shouting(spelling) ||
        ctx.settings.list('allow', []).includes(spelling) ||
        environment(ctx, ctx.initializer(index))
    )
        return;
    const list = ctx.parent(index);
    if(kind(ctx, list) !== 'VariableDeclarationList' || !['2', '6'].includes(ctx.node(list).semantic)) return;
    const statement = ctx.parent(list);
    const exported = statement >= 0 && ctx.hasKind(statement, 'ExportKeyword');
    const id = exported ? 'noScreamingSnakeCaseExported' : 'noScreamingSnakeCaseLocal';
    ctx.report(
        name,
        'nexus/consistency-no-screaming-snake-case',
        id,
        text(`nexus_consistency_no_screaming_snake_case:${id}`)
            .replaceAll('{{name}}', spelling)
            .replaceAll('{{suggestion}}', convert(spelling, exported)),
    );
}
