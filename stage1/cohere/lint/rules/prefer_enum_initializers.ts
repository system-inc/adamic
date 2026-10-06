import { panic } from 'adamic';
import type { RuleContext } from './context.ts';
import { Finding } from '../finding.ts';
import { RepairEdit } from '../repair_edit.ts';
import { suggest } from './shared.ts';

export function preferEnumInitializers(context: RuleContext, index: number): void {
    let ordinal = 0;
    for(const member of context.node(index).children) {
        const node = context.node(member);
        if(node.kind !== 'EnumMember') {
            continue;
        }
        const position = ordinal;
        ordinal++;
        if(node.children.length > 1) {
            continue;
        }
        const start = context.start(member);
        const name = context.source.slice(start, node.end);
        const values = [position.toString(), (position + 1).toString(), `'${name}'`];
        const first = values[0] ?? panic('missing enum value');
        const finding = new Finding(
            '@typescript-eslint/prefer-enum-initializers',
            'defineInitializer',
            `The value of the member '${name}' should be explicitly defined.`,
            start,
            node.end,
            'suggestion',
            `${name} = ${first}`,
            `Can be fixed to ${name} = ${first}`,
        );
        for(const value of values) {
            suggest(finding, 'defineInitializerSuggestion', `Can be fixed to ${name} = ${value}`, [
                new RepairEdit(start, node.end, `${name} = ${value}`),
            ]);
        }
        context.record(finding);
    }
}

export function visit(context: RuleContext, index: number): void {
    if(
        context.node(index).kind === 'EnumDeclaration' &&
        context.enabled('@typescript-eslint/prefer-enum-initializers')
    ) {
        preferEnumInitializers(context, index);
    }
}
