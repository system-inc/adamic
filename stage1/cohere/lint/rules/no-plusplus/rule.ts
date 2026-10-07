import { panic } from 'adamic';
import type { RuleContext } from '../../context.ts';
import type { ParseNode } from '../../../../typescript/parser/nodes.ts';
import { messageUnexpectedUnaryOp } from './messages.ts';

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) {
        this.context = context;
    }
    visit(node: ParseNode, index: number): void {
        if(node.operator !== 'PlusPlusToken' && node.operator !== 'MinusMinusToken') {
            return;
        }
        if(this.context.settings.read('allowforloopafterthoughts', 'false') === 'true' && this.afterthought(index)) {
            return;
        }
        this.context.report(
            index,
            'no-plusplus',
            'unexpectedUnaryOp',
            messageUnexpectedUnaryOp(node.operator === 'MinusMinusToken' ? '--' : '++'),
            '',
            '',
            '',
        );
    }
    // afterthought is whether the update is a for statement's incrementor, alone or in a comma list.
    afterthought(index: number): boolean {
        let current = index;
        for(;;) {
            const parent = this.context.parent(current);
            if(parent < 0) {
                return false;
            }
            const node = this.context.node(parent);
            if(
                node.kind === 'ParenthesizedExpression' ||
                (node.kind === 'BinaryExpression' &&
                    this.context.node(node.children[1] ?? panic('operator')).kind === 'CommaToken')
            ) {
                current = parent;
                continue;
            }
            if(node.kind !== 'ForStatement') {
                return false;
            }
            // Header-level semicolon count distinguishes incrementor from initializer and test.
            const limit = this.context.start(current);
            const scanner = this.context.scanner;
            scanner.pos = this.context.start(parent);
            let depth = 0;
            let count = 0;
            while(scanner.scan() !== 'EndOfFile' && scanner.start < limit) {
                if(scanner.kind === 'OpenParenToken') {
                    depth++;
                }
                if(scanner.kind === 'CloseParenToken') {
                    depth--;
                }
                if(scanner.kind === 'SemicolonToken' && depth === 1) {
                    count++;
                }
            }
            return count === 2 && node.children[node.children.length - 1] !== current;
        }
    }
}

export function create(context: RuleContext): Rule {
    return new Rule(context);
}
