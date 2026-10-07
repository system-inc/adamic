import type { RuleContext } from '../../context.ts';
import type { ParseNode } from '../../../../typescript/parser/nodes.ts';
import { messageAwait } from './messages.ts';

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) {
        this.context = context;
    }
    visit(node: ParseNode, index: number, parent: number): void {
        if(node.kind === 'ForOfStatement' && !this.context.has(index, 'AwaitKeyword')) {
            return;
        }
        // An `await using` declaration list awaits its disposal.
        if(node.kind === 'VariableDeclarationList' && node.semantic !== '6') {
            return;
        }
        if(!this.awaitedLoop(index)) {
            return;
        }
        const subject =
            node.kind === 'VariableDeclarationList' &&
            parent >= 0 &&
            this.context.node(parent).kind === 'VariableStatement'
                ? parent
                : index;
        this.context.report(subject, 'no-await-in-loop', 'unexpectedAwait', messageAwait, '', '', '');
    }
    // awaitedLoop is whether the node repeats with a loop, stopping at the function, static block or
    // for-await that owns it.
    awaitedLoop(index: number): boolean {
        let cursor = index;
        for(;;) {
            const parent = this.context.parent(cursor);
            if(parent < 0) {
                return false;
            }
            const node = this.context.node(parent);
            if(
                this.context.functionLike(parent) ||
                node.kind === 'ClassStaticBlockDeclaration' ||
                (node.kind === 'ForOfStatement' && this.context.has(parent, 'AwaitKeyword'))
            ) {
                return false;
            }
            if(node.kind === 'WhileStatement' || node.kind === 'DoStatement') {
                return true;
            }
            if(node.kind === 'ForOfStatement' || node.kind === 'ForInStatement') {
                if(cursor === node.children[node.children.length - 1] || this.context.node(cursor).semantic === '6') {
                    return true;
                }
            }
            if(node.kind === 'ForStatement') {
                // Before the initializer there is no semicolon token. Every later
                // immediate child follows one, including omitted-header forms.
                this.context.scanner.pos = this.context.start(parent);
                while(
                    this.context.scanner.scan() !== 'EndOfFile' &&
                    this.context.scanner.start < this.context.node(cursor).pos
                ) {
                    if(this.context.scanner.kind === 'SemicolonToken') {
                        return true;
                    }
                }
            }
            cursor = parent;
        }
    }
}

export function create(context: RuleContext): Rule {
    return new Rule(context);
}
