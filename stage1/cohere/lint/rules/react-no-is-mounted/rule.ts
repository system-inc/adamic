import type { RuleContext } from '../../context.ts';
import { messageNoIsMounted } from './messages.ts';
export function noIsMounted(context: RuleContext, index: number): void {
    if(context.kind(index) !== 'CallExpression') {
        return;
    }
    const callee = context.child(index, 0);
    if(!['PropertyAccessExpression', 'ElementAccessExpression'].includes(context.kind(callee))) {
        return;
    }
    if(context.kind(context.child(callee, 0)) !== 'ThisKeyword') {
        return;
    }
    const name = context.property(callee);
    if(context.node(name).text !== 'isMounted') {
        return;
    }
    if(
        context.kind(callee) === 'ElementAccessExpression' &&
        !['StringLiteral', 'NoSubstitutionTemplateLiteral'].includes(context.kind(name))
    ) {
        return;
    }
    let ancestor = context.parent(index);
    while(ancestor >= 0) {
        if(['PropertyAssignment', 'MethodDeclaration', 'GetAccessor', 'SetAccessor'].includes(context.kind(ancestor))) {
            context.reportNode(index, 'react/no-is-mounted', 'noIsMounted', messageNoIsMounted);
            return;
        }
        ancestor = context.parent(ancestor);
    }
}

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) {
        this.context = context;
    }
    visit(index: number): void {
        noIsMounted(this.context, index);
    }
}

export function create(context: RuleContext): Rule {
    return new Rule(context);
}
