import type { Batch5Context } from './batch5_context.ts';
import { first, last, name, initializer } from './batch5_helpers.ts';
import { ruleMessage } from './batch5_messages.ts';
import { RepairEdit } from './repair_edit.ts';
function same(context: Batch5Context, type: number, value: number): boolean {
    if(type < 0 || value < 0 || context.node(type).kind !== 'LiteralType') {
        return false;
    }
    const literal = first(context, type);
    return (
        literal >= 0 &&
        ['StringLiteral', 'NumericLiteral'].includes(context.node(literal).kind) &&
        context.node(literal).kind === context.node(value).kind &&
        context.node(literal).text === context.node(value).text
    );
}
export function visit(context: Batch5Context, index: number): void {
    const rule = '@typescript-eslint/prefer-as-const';
    if(!context.enabled(rule)) {
        return;
    }
    const kind = context.node(index).kind;
    const edits: RepairEdit[] = [];
    let type = -1;
    if(kind === 'AsExpression') {
        type = last(context, index);
        if(!same(context, type, first(context, index))) {
            return;
        }
        edits.push(new RepairEdit(context.start(type), context.node(type).end, 'const'));
    }
    else if(['VariableDeclaration', 'PropertyDeclaration'].includes(kind)) {
        if(context.hasKind(index, 'ExclamationToken') || context.hasKind(index, 'QuestionToken')) {
            return;
        }
        const binding = name(context, index);
        const value = initializer(context, index);
        if(binding < 0 || value < 0) {
            return;
        }
        for(const child of context.node(index).children) {
            if(child !== value && context.node(child).kind === 'LiteralType') {
                type = child;
            }
        }
        if(!same(context, type, value)) {
            return;
        }
        if(kind === 'PropertyDeclaration' || context.node(binding).kind === 'Identifier') {
            const colon = context.scanAt(context.node(binding).end).start;
            edits.push(new RepairEdit(colon, context.node(type).end, ''));
            const end = context.node(value).end;
            edits.push(new RepairEdit(end, end, ' as const'));
        }
    }
    else {
        return;
    }
    context.report(type, rule, 'preferAsConst', ruleMessage('preferAsConst'), edits, []);
}
