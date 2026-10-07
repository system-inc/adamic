import type { RuleContext } from '../../context.ts';
import { varMessage } from './messages.ts';

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) {
        this.context = context;
    }
    visit(index: number, parent: number): void {
        const node = this.context.node(index);
        if(node.kind === 'VariableStatement') {
            for(const child of node.children) {
                if(this.context.node(child).kind === 'DeclareKeyword') {
                    return;
                }
            }
            if(parent >= 0 && this.context.node(parent).kind === 'ModuleBlock') {
                const grandparent = this.context.parents[parent] ?? -1;
                if(grandparent >= 0 && this.context.node(grandparent).operator === 'GlobalKeyword') {
                    return;
                }
            }
        }
        for(const child of node.children) {
            const list = this.context.node(child);
            if(list.kind === 'VariableDeclarationList' && list.semantic === '0') {
                this.context.report(child, 'no-var', 'unexpectedVar', varMessage, '', '', '');
            }
        }
    }
}

export function create(context: RuleContext): Rule {
    return new Rule(context);
}
