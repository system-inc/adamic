import type { RuleContext } from '../../context.ts';
import type { ParseNode } from '../../../../typescript/parser/nodes.ts';
import { messageVars } from './messages.ts';

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) {
        this.context = context;
    }
    // A `var` list reports unless it is one of the leading statements of its function, static block or file,
    // after any directive prologue and imports.
    visit(node: ParseNode, index: number, parent: number): void {
        if(node.semantic !== '0') {
            return;
        }
        const subject = parent >= 0 && this.context.node(parent).kind === 'VariableStatement' ? parent : index;
        const container = this.context.parent(subject);
        if(container < 0) {
            return;
        }
        const owner = this.context.parent(container);
        const staticBlock = owner >= 0 && this.context.node(owner).kind === 'ClassStaticBlockDeclaration';
        if(
            this.context.node(container).kind === 'SourceFile' ||
            (this.context.node(container).kind === 'Block' &&
                owner >= 0 &&
                (this.context.functionLike(owner) || staticBlock))
        ) {
            let prefix = !staticBlock;
            for(const statement of this.context.node(container).children) {
                const statementNode = this.context.node(statement);
                const expression = statementNode.children[0] ?? -1;
                const directive =
                    statementNode.kind === 'ExpressionStatement' &&
                    expression >= 0 &&
                    this.context.node(expression).kind === 'StringLiteral';
                if(
                    prefix &&
                    (directive ||
                        statementNode.kind === 'ImportDeclaration' ||
                        statementNode.kind === 'ImportEqualsDeclaration')
                ) {
                    continue;
                }
                prefix = false;
                if(statementNode.kind !== 'VariableStatement') {
                    break;
                }
                if(statement === subject) {
                    return;
                }
            }
        }
        this.context.report(subject, 'vars-on-top', 'top', messageVars, '', '', '');
    }
}

export function create(context: RuleContext): Rule {
    return new Rule(context);
}
