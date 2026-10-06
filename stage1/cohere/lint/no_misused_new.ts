import type { Batch5Context } from './batch5_context.ts';
import { first, name, returnType } from './batch5_helpers.ts';
import { ruleMessage } from './batch5_messages.ts';
function returnName(context: Batch5Context, member: number): string {
    const type = returnType(context, member);
    if(type < 0 || context.node(type).kind !== 'TypeReference') {
        return '';
    }
    const binding = first(context, type);
    return binding >= 0 && context.node(binding).kind === 'Identifier' ? context.node(binding).text : '';
}
export function visit(context: Batch5Context, index: number): void {
    const rule = '@typescript-eslint/no-misused-new';
    if(!context.enabled(rule)) {
        return;
    }
    const kind = context.node(index).kind;
    if(kind === 'MethodSignature') {
        const key = name(context, index);
        if(key >= 0 && context.node(key).kind === 'Identifier' && context.node(key).text === 'constructor') {
            context.report(key, rule, 'interfaceConstructor', ruleMessage('interfaceConstructor'), [], []);
        }
        return;
    }
    if(!['InterfaceDeclaration', 'ClassDeclaration', 'ClassExpression'].includes(kind)) {
        return;
    }
    const binding = name(context, index);
    if(binding < 0 || context.node(binding).kind !== 'Identifier') {
        return;
    }
    const className = context.node(binding).text;
    const members: number[] = [];
    for(const child of context.node(index).children) {
        if(context.node(child).kind === 'TypeLiteral') {
            for(const member of context.node(child).children) {
                members.push(member);
            }
        }
        else {
            members.push(child);
        }
    }
    for(const member of members) {
        const memberKind = context.node(member).kind;
        if(kind === 'InterfaceDeclaration') {
            if(memberKind === 'ConstructSignature' && returnName(context, member) === className) {
                context.report(
                    member,
                    rule,
                    'interfaceConstruct',
                    ruleMessage('interfaceConstruct'),
                    [],
                    [],
                    context.start(member) + 3,
                );
            }
        }
        else if(
            ['MethodDeclaration', 'GetAccessor'].includes(memberKind) &&
            !context.hasKind(member, 'Block') &&
            returnName(context, member) === className
        ) {
            const key = name(context, member);
            if(key >= 0 && context.node(key).kind === 'Identifier' && context.node(key).text === 'new') {
                context.report(key, rule, 'classNew', ruleMessage('classNew'), [], []);
            }
        }
    }
}
