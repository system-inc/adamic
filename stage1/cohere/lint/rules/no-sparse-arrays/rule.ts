import { panic } from 'adamic';
import type { RuleContext } from '../../context.ts';
import type { ParseNode } from '../../../../typescript/parser/nodes.ts';
import { messageSparse } from './messages.ts';

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) {
        this.context = context;
    }
    visit(node: ParseNode, index: number): void {
        if(!this.assignmentTarget(index) && this.context.has(index, 'OmittedExpression')) {
            this.context.report(index, 'no-sparse-arrays', 'unexpectedSparseArray', messageSparse, '', '', '');
        }
    }
    // assignmentTarget is whether the array is written to rather than read, where a hole skips an element.
    assignmentTarget(index: number): boolean {
        let cursor = index;
        for(;;) {
            const parent = this.context.parent(cursor);
            if(parent < 0) {
                return false;
            }
            const node = this.context.node(parent);
            const first = node.children[0] ?? -1;
            switch(node.kind) {
                case 'BinaryExpression': {
                    const operator = this.context.node(node.children[1] ?? panic('missing operator')).kind;
                    return (
                        first === cursor &&
                        [
                            'EqualsToken',
                            'PlusEqualsToken',
                            'MinusEqualsToken',
                            'AsteriskEqualsToken',
                            'AsteriskAsteriskEqualsToken',
                            'SlashEqualsToken',
                            'PercentEqualsToken',
                            'LessThanLessThanEqualsToken',
                            'GreaterThanGreaterThanEqualsToken',
                            'GreaterThanGreaterThanGreaterThanEqualsToken',
                            'AmpersandEqualsToken',
                            'BarEqualsToken',
                            'CaretEqualsToken',
                            'BarBarEqualsToken',
                            'AmpersandAmpersandEqualsToken',
                            'QuestionQuestionEqualsToken',
                        ].includes(operator)
                    );
                }
                case 'PrefixUnaryExpression':
                case 'PostfixUnaryExpression':
                    return node.operator === 'PlusPlusToken' || node.operator === 'MinusMinusToken';
                case 'ForInStatement':
                case 'ForOfStatement':
                    return (
                        (first >= 0 && this.context.node(first).kind === 'AwaitKeyword' ? node.children[1] : first) ===
                        cursor
                    );
                case 'ParenthesizedExpression':
                case 'ArrayLiteralExpression':
                case 'SpreadElement':
                case 'NonNullExpression':
                    cursor = parent;
                    break;
                case 'SpreadAssignment':
                    cursor = this.context.parent(parent);
                    break;
                case 'ShorthandPropertyAssignment':
                    if(first !== cursor) {
                        return false;
                    }
                    cursor = this.context.parent(parent);
                    break;
                case 'PropertyAssignment':
                    if(first === cursor) {
                        return false;
                    }
                    cursor = this.context.parent(parent);
                    break;
                default:
                    return false;
            }
            if(cursor < 0) {
                return false;
            }
        }
    }
}

export function create(context: RuleContext): Rule {
    return new Rule(context);
}
