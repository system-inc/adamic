import { panic } from 'adamic';
import type { RuleContext } from '../../context.ts';
import type { ParseNode } from '../../../../typescript/parser/nodes.ts';
import { messageArrowAssignment, messageReturnAssignment } from './messages.ts';

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) {
        this.context = context;
    }
    // An assignment reports at the return statement or concise arrow body it is the value of.
    visit(node: ParseNode, index: number): void {
        const operator = this.context.node(node.children[1] ?? panic('assignment operator')).kind;
        if(
            !operator.endsWith('EqualsToken') ||
            [
                'EqualsEqualsToken',
                'EqualsEqualsEqualsToken',
                'ExclamationEqualsToken',
                'ExclamationEqualsEqualsToken',
                'LessThanEqualsToken',
                'GreaterThanEqualsToken',
            ].includes(operator)
        ) {
            return;
        }
        this.context.scanner.pos = node.end;
        this.context.scanner.scan();
        if(
            this.context.settings.read('option', 'except-parens') !== 'always' &&
            this.context.source[node.pos - 1] === '(' &&
            this.context.source[this.context.scanner.start] === ')'
        ) {
            return;
        }
        let current = index;
        for(;;) {
            const parent = this.context.parent(current);
            if(parent < 0) {
                return;
            }
            const kind = this.context.node(parent).kind;
            if(
                kind.endsWith('Statement') ||
                kind === 'ArrowFunction' ||
                kind === 'FunctionExpression' ||
                kind === 'ClassExpression'
            ) {
                if(
                    kind === 'ReturnStatement' ||
                    (kind === 'ArrowFunction' &&
                        this.context.node(parent).children[this.context.node(parent).children.length - 1] === current)
                ) {
                    this.context.report(
                        parent,
                        'no-return-assign',
                        kind === 'ReturnStatement' ? 'returnAssignment' : 'arrowAssignment',
                        kind === 'ReturnStatement' ? messageReturnAssignment : messageArrowAssignment,
                        '',
                        '',
                        '',
                    );
                }
                return;
            }
            current = parent;
        }
    }
}

export function create(context: RuleContext): Rule {
    return new Rule(context);
}
