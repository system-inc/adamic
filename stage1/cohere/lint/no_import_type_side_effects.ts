import type { Batch5Context } from './batch5_context.ts';
import { first } from './batch5_helpers.ts';
import { ruleMessage } from './batch5_messages.ts';
import { RepairEdit } from './repair_edit.ts';
export function visit(context: Batch5Context, index: number): void {
    const rule = '@typescript-eslint/no-import-type-side-effects';
    if(context.node(index).kind !== 'ImportDeclaration' || !context.enabled(rule)) {
        return;
    }
    let bindings = -1;
    for(const child of context.node(index).children) {
        if(context.node(child).kind !== 'ImportClause') {
            continue;
        }
        if(context.node(child).semantic === 'TypeKeyword') {
            return;
        }
        for(const part of context.node(child).children) {
            if(context.node(part).kind !== 'NamedImports') {
                return;
            }
            bindings = part;
        }
    }
    if(bindings < 0 || context.node(bindings).children.length === 0) {
        return;
    }
    const edits: RepairEdit[] = [];
    for(const specifier of context.node(bindings).children) {
        if(context.node(specifier).kind !== 'ImportSpecifier' || context.node(specifier).semantic !== '1') {
            return;
        }
        edits.push(new RepairEdit(context.start(specifier), context.start(first(context, specifier)), ''));
    }
    const insertion = context.start(index) + 6;
    edits.push(new RepairEdit(insertion, insertion, ' type'));
    context.report(index, rule, 'useTopLevelQualifier', ruleMessage('useTopLevelQualifier'), edits, []);
}
