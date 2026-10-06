import type { Batch5Context } from './batch5_context.ts';
import { ruleMessage } from './batch5_messages.ts';
import { RepairEdit } from './repair_edit.ts';

function empty(context: Batch5Context, index: number): boolean {
    if(context.node(index).kind !== 'ExportDeclaration' || context.hasKind(index, 'StringLiteral')) {
        return false;
    }
    for(const child of context.node(index).children) {
        if(context.node(child).kind === 'NamedExports') {
            return context.node(child).children.length === 0;
        }
    }
    return false;
}
function moduleStatement(context: Batch5Context, index: number): boolean {
    const kind = context.node(index).kind;
    if(kind === 'ImportDeclaration' || kind === 'ExportDeclaration') {
        return true;
    }
    if(kind === 'ExportAssignment') {
        return context.node(index).semantic !== '1';
    }
    if(kind === 'ImportEqualsDeclaration') {
        return context.hasKind(index, 'ExternalModuleReference');
    }
    return context.hasKind(index, 'ExportKeyword');
}
export function visit(context: Batch5Context, index: number): void {
    const rule = '@typescript-eslint/no-useless-empty-export';
    if(context.node(index).kind !== 'SourceFile' || !context.enabled(rule) || context.parser.path.endsWith('.d.ts')) {
        return;
    }
    const exports: number[] = [];
    let isModule = false;
    for(const statement of context.node(index).children) {
        if(empty(context, statement)) {
            exports.push(statement);
        }
        else if(moduleStatement(context, statement)) {
            isModule = true;
        }
    }
    if(!isModule) {
        return;
    }
    for(const statement of exports) {
        context.report(
            statement,
            rule,
            'uselessEmptyExport',
            ruleMessage('uselessEmptyExport'),
            [new RepairEdit(context.start(statement), context.node(statement).end, '')],
            [],
        );
    }
}
