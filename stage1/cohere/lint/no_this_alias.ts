import type { Batch5Context } from './batch5_context.ts';
import { first, last, name, initializer, assignmentOperator, typescriptFile } from './batch5_helpers.ts';
import { ruleMessage } from './batch5_messages.ts';
export function visit(context: Batch5Context, index: number): void {
    const rule = '@typescript-eslint/no-this-alias';
    if(!context.enabled(rule) || !typescriptFile(context.parser.path)) {
        return;
    }
    const kind = context.node(index).kind;
    let target = -1;
    let value = -1;
    if(kind === 'VariableDeclaration') {
        target = name(context, index);
        value = initializer(context, index);
    }
    else if(
        kind === 'BinaryExpression' &&
        assignmentOperator(context.node(context.node(index).children[1] ?? -1).kind)
    ) {
        target = first(context, index);
        value = last(context, index);
        if(!['ObjectLiteralExpression', 'ArrayLiteralExpression'].includes(context.node(target).kind)) {
            while(
                [
                    'ParenthesizedExpression',
                    'AsExpression',
                    'SatisfiesExpression',
                    'TypeAssertionExpression',
                    'NonNullExpression',
                ].includes(context.node(target).kind)
            ) {
                target =
                    context.node(target).kind === 'TypeAssertionExpression'
                        ? last(context, target)
                        : first(context, target);
            }
            if(context.node(target).kind !== 'Identifier') {
                return;
            }
        }
    }
    if(target < 0 || value < 0 || context.node(value).kind !== 'ThisKeyword') {
        return;
    }
    const identifier = context.node(target).kind === 'Identifier';
    if(
        identifier
            ? context.settings.list('allowednames', []).includes(context.node(target).text)
            : context.settings.read('reportdestructuring', 'false') !== 'true'
    ) {
        return;
    }
    const id = identifier ? 'thisAssignment' : 'thisDestructure';
    context.report(target, rule, id, ruleMessage(id), [], []);
}
