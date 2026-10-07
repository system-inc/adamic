import type { RuleContext } from '../../context.ts';
import { blockMessage, switchMessage } from './messages.ts';

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) {
        this.context = context;
    }
    visit(index: number, parent: number): void {
        const node = this.context.node(index);
        const parentKind = parent < 0 ? '' : this.context.node(parent).kind;
        if(node.kind === 'Block' && node.children.length === 0) {
            const functionBody = [
                'FunctionDeclaration',
                'FunctionExpression',
                'ArrowFunction',
                'MethodDeclaration',
                'Constructor',
                'GetAccessor',
                'SetAccessor',
            ].includes(parentKind);
            if(
                !functionBody &&
                !(this.context.allowCatch && parentKind === 'CatchClause') &&
                !this.context.comment(index)
            ) {
                this.context.report(index, 'no-empty', 'unexpectedBlock', blockMessage, '', '', '');
            }
        }
        if(node.kind === 'SwitchStatement') {
            for(const child of node.children) {
                if(
                    this.context.node(child).kind === 'CaseBlock' &&
                    this.context.node(child).children.length === 0 &&
                    !this.context.comment(child)
                ) {
                    this.context.report(index, 'no-empty', 'unexpectedSwitch', switchMessage, '', '', '');
                }
            }
        }
    }
}

export function create(context: RuleContext): Rule {
    return new Rule(context);
}
