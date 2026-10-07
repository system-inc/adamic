import type { RuleContext } from '../../context.ts';
import type { ParseNode } from '../../../../typescript/parser/nodes.ts';
import { messageYield } from './messages.ts';

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) {
        this.context = context;
    }
    visit(node: ParseNode, index: number): void {
        if(!this.context.has(index, 'AsteriskToken')) {
            return;
        }
        for(const child of node.children) {
            if(
                this.context.node(child).kind === 'Block' &&
                this.context.node(child).children.length > 0 &&
                !this.containsYield(child)
            ) {
                this.context.report(index, 'require-yield', 'missingYield', messageYield, '', '', '');
            }
        }
    }
    // containsYield stops at nested functions and classes, whose yields belong to them.
    containsYield(index: number): boolean {
        if(this.context.node(index).kind === 'YieldExpression') {
            return true;
        }
        for(const child of this.context.node(index).children) {
            if(
                [
                    'FunctionDeclaration',
                    'FunctionExpression',
                    'ArrowFunction',
                    'MethodDeclaration',
                    'ClassDeclaration',
                    'ClassExpression',
                ].includes(this.context.node(child).kind)
            ) {
                continue;
            }
            if(this.containsYield(child)) {
                return true;
            }
        }
        return false;
    }
}

export function create(context: RuleContext): Rule {
    return new Rule(context);
}
