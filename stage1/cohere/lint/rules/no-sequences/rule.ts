import { panic } from 'adamic';
import type { RuleContext } from '../../context.ts';
import type { ParseNode } from '../../../../typescript/parser/nodes.ts';
import { messageUnexpectedCommaExpression } from './messages.ts';

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) {
        this.context = context;
    }
    // One finding per comma chain, at the chain's first comma, so only the outermost comma operator reports.
    visit(node: ParseNode, index: number, parent: number): void {
        const operator = node.children[1] ?? panic('operator');
        if(this.context.node(operator).kind === 'CommaToken' && !this.comma(parent) && !this.forSlot(index)) {
            this.chain(index, parent);
        }
    }
    chain(index: number, parent: number): void {
        const grandparent = this.context.parent(parent);
        const parens =
            parent >= 0 &&
            this.context.node(parent).kind === 'ParenthesizedExpression' &&
            !(grandparent >= 0 && this.context.node(grandparent).kind === 'ArrowFunction');
        if(this.context.settings.read('allowinparentheses', 'true') === 'true' && parens) {
            return;
        }
        let cursor = index;
        while(this.comma(this.context.node(cursor).children[0] ?? -1)) {
            cursor = this.context.node(cursor).children[0] ?? panic('left');
        }
        this.context.report(
            this.context.node(cursor).children[1] ?? panic('comma'),
            'no-sequences',
            'unexpectedCommaExpression',
            messageUnexpectedCommaExpression,
            '',
            '',
            '',
        );
    }
    comma(index: number): boolean {
        return (
            index >= 0 &&
            this.context.node(index).kind === 'BinaryExpression' &&
            this.context.node(this.context.node(index).children[1] ?? panic('operator')).kind === 'CommaToken'
        );
    }
    // forSlot is whether the expression is a for statement's initializer or update, where a comma is the
    // idiom rather than a hidden second statement.
    forSlot(index: number): boolean {
        let cursor = index;
        let parent = this.context.parent(cursor);
        while(parent >= 0 && this.context.node(parent).kind === 'ParenthesizedExpression') {
            cursor = parent;
            parent = this.context.parent(cursor);
        }
        if(parent < 0 || this.context.node(parent).kind !== 'ForStatement') {
            return false;
        }
        const children = this.context.node(parent).children;
        if(children[children.length - 1] === cursor) {
            return false;
        }
        // Count only header-level semicolons; nested function bodies and regex
        // literal spans cannot masquerade as separators. The literal map is built before the scanner is
        // positioned, because building it moves the scanner.
        const literalEnds = this.context.literalEnds();
        const limit = this.context.start(cursor);
        const scanner = this.context.scanner;
        scanner.pos = this.context.start(parent);
        let parens = 0;
        let brackets = 0;
        let braces = 0;
        let separators = 0;
        while(scanner.scan() !== 'EndOfFile' && scanner.start < limit) {
            const literalEnd = literalEnds[scanner.start] ?? -1;
            if(literalEnd >= 0) {
                scanner.pos = literalEnd;
                continue;
            }
            if(scanner.kind === 'OpenParenToken') {
                parens++;
            }
            if(scanner.kind === 'CloseParenToken') {
                parens--;
            }
            if(scanner.kind === 'OpenBracketToken') {
                brackets++;
            }
            if(scanner.kind === 'CloseBracketToken') {
                brackets--;
            }
            if(scanner.kind === 'OpenBraceToken') {
                braces++;
            }
            if(scanner.kind === 'CloseBraceToken') {
                braces--;
            }
            if(scanner.kind === 'SemicolonToken' && parens === 1 && brackets === 0 && braces === 0) {
                separators++;
            }
        }
        return separators === 0 || separators === 2;
    }
}

export function create(context: RuleContext): Rule {
    return new Rule(context);
}
