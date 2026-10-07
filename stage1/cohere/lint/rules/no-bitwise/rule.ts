import { panic } from 'adamic';
import type { RuleContext } from '../../context.ts';
import type { ParseNode } from '../../../../typescript/parser/nodes.ts';
import { messageBitwise } from './messages.ts';

// bitwiseOperator is a hot kind test, so it switches rather than allocating a lookup table.
function bitwiseOperator(kind: string): string {
    switch(kind) {
        case 'CaretToken':
            return '^';
        case 'BarToken':
            return '|';
        case 'AmpersandToken':
            return '&';
        case 'LessThanLessThanToken':
            return '<<';
        case 'GreaterThanGreaterThanToken':
            return '>>';
        case 'GreaterThanGreaterThanGreaterThanToken':
            return '>>>';
        case 'CaretEqualsToken':
            return '^=';
        case 'BarEqualsToken':
            return '|=';
        case 'AmpersandEqualsToken':
            return '&=';
        case 'LessThanLessThanEqualsToken':
            return '<<=';
        case 'GreaterThanGreaterThanEqualsToken':
            return '>>=';
        case 'GreaterThanGreaterThanGreaterThanEqualsToken':
            return '>>>=';
        case 'TildeToken':
            return '~';
        default:
            return '';
    }
}

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) {
        this.context = context;
    }
    visit(node: ParseNode, index: number): void {
        const operator = bitwiseOperator(
            node.kind === 'BinaryExpression'
                ? this.context.node(node.children[1] ?? panic('operator')).kind
                : node.operator,
        );
        if(operator === '') {
            return;
        }
        const right = node.children[2] ?? -1;
        const hint =
            this.context.settings.read('int32hint', 'false') === 'true' &&
            operator === '|' &&
            right >= 0 &&
            this.context.node(right).kind === 'NumericLiteral' &&
            this.context.node(right).text === '0';
        if(!this.context.settings.list('allow', []).includes(operator) && !hint) {
            this.context.report(
                index,
                'no-bitwise',
                'unexpected',
                `Unexpected use of '${operator}'. ${messageBitwise}`,
                '',
                '',
                '',
            );
        }
    }
}

export function create(context: RuleContext): Rule {
    return new Rule(context);
}
