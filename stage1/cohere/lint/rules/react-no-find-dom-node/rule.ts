import type { RuleContext } from '../../context.ts';
import { messageNoFindDOMNode } from './messages.ts';
export function noFindDomNode(context: RuleContext, index: number): void {
    if(context.kind(index) !== 'CallExpression') {
        return;
    }
    const callee = context.unwrap(context.child(index, 0));
    let name = -1;
    if(context.kind(callee) === 'Identifier') {
        name = callee;
    }
    else if(['PropertyAccessExpression', 'ElementAccessExpression'].includes(context.kind(callee))) {
        const receiver = context.unwrap(context.child(callee, 0));
        if(
            context.kind(receiver) !== 'Identifier' ||
            !['React', 'ReactDOM', 'ReactDom'].includes(context.node(receiver).text)
        ) {
            return;
        }
        name = context.property(callee);
        if(
            context.kind(callee) === 'ElementAccessExpression' &&
            !['StringLiteral', 'NoSubstitutionTemplateLiteral'].includes(context.kind(name))
        ) {
            return;
        }
    }
    if(name >= 0 && context.node(name).text === 'findDOMNode') {
        context.reportNode(name, 'react/no-find-dom-node', 'noFindDOMNode', messageNoFindDOMNode);
    }
}

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) {
        this.context = context;
    }
    visit(index: number): void {
        noFindDomNode(this.context, index);
    }
}

export function create(context: RuleContext): Rule {
    return new Rule(context);
}
