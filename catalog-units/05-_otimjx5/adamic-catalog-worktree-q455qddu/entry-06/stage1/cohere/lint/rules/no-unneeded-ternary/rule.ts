import { panic } from 'adamic';
import type { RuleContext } from '../../context.ts';
import type { ParseNode } from '../../../../typescript/parser/nodes.ts';
import {
    messageNoUnneededTernaryConditionalExpression,
    messageNoUnneededTernaryConditionalAssignment,
} from './messages.ts';

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) {
        this.context = context;
    }
    visit(node: ParseNode, index: number): void {
        const test = node.children[0] ?? panic('test');
        const yes = node.children[2] ?? panic('consequent');
        const no = node.children[4] ?? panic('alternate');
        const yesKind = this.context.node(this.context.unwrap(yes)).kind;
        const noKind = this.context.node(this.context.unwrap(no)).kind;
        if(['TrueKeyword', 'FalseKeyword'].includes(yesKind) && ['TrueKeyword', 'FalseKeyword'].includes(noKind)) {
            let repair = 'fix';
            let replacement: string;
            if(yesKind === noKind) {
                replacement = yesKind === 'TrueKeyword' ? 'true' : 'false';
                if(this.context.node(this.context.unwrap(test)).kind !== 'Identifier') {
                    repair = '';
                    replacement = '';
                }
            }
            else if(noKind === 'TrueKeyword') {
                replacement = this.invert(test);
            }
            else {
                replacement = this.alwaysBoolean(test) ? this.context.raw(test) : `!${this.invert(test)}`;
            }
            this.context.report(
                index,
                'no-unneeded-ternary',
                'unnecessaryConditionalExpression',
                messageNoUnneededTernaryConditionalExpression,
                repair,
                replacement,
                '',
            );
            return;
        }
        const testNode = this.context.node(this.context.unwrap(test));
        const yesNode = this.context.node(this.context.unwrap(yes));
        if(
            this.context.settings.read('defaultassignment', 'true') === 'false' &&
            testNode.kind === 'Identifier' &&
            yesNode.kind === 'Identifier' &&
            testNode.text === yesNode.text
        ) {
            let alternate = this.context.raw(no);
            const noNode = this.context.node(no);
            const coalesce =
                noNode.kind === 'BinaryExpression' &&
                this.context.raw(noNode.children[1] ?? panic('operator')) === '??';
            if(noNode.kind !== 'ParenthesizedExpression' && (this.precedence(no) < 4 || coalesce)) {
                alternate = `(${alternate})`;
            }
            this.context.report(
                index,
                'no-unneeded-ternary',
                'unnecessaryConditionalAssignment',
                messageNoUnneededTernaryConditionalAssignment,
                'fix',
                `${this.context.raw(test)} || ${alternate}`,
                '',
            );
        }
    }
    precedence(index: number): number {
        const node = this.context.node(index);
        if(node.kind === 'ParenthesizedExpression') {
            return this.precedence(node.children[0] ?? panic('operand'));
        }
        if(node.kind === 'BinaryExpression') {
            const operator = this.context.raw(node.children[1] ?? panic('operator'));
            if(operator === ',') {
                return 0;
            }
            if(
                [
                    '=',
                    '+=',
                    '-=',
                    '*=',
                    '**=',
                    '/=',
                    '%=',
                    '<<=',
                    '>>=',
                    '>>>=',
                    '&=',
                    '|=',
                    '^=',
                    '&&=',
                    '||=',
                    '??=',
                ].includes(operator)
            ) {
                return 1;
            }
            if(['??', '||'].includes(operator)) {
                return 4;
            }
            if(operator === '&&') {
                return 5;
            }
            if(operator === '|') {
                return 6;
            }
            if(operator === '^') {
                return 7;
            }
            if(operator === '&') {
                return 8;
            }
            if(['==', '!=', '===', '!=='].includes(operator)) {
                return 9;
            }
            if(['<', '<=', '>', '>=', 'in', 'instanceof'].includes(operator)) {
                return 10;
            }
            if(['<<', '>>', '>>>'].includes(operator)) {
                return 11;
            }
            if(['+', '-'].includes(operator)) {
                return 12;
            }
            if(['*', '/', '%'].includes(operator)) {
                return 13;
            }
            if(operator === '**') {
                return 15;
            }
            return -1;
        }
        if(node.kind === 'ConditionalExpression') {
            return 3;
        }
        if(['ArrowFunction', 'YieldExpression'].includes(node.kind)) {
            return 1;
        }
        if(
            [
                'PrefixUnaryExpression',
                'TypeOfExpression',
                'VoidExpression',
                'DeleteExpression',
                'AwaitExpression',
            ].includes(node.kind)
        ) {
            return 16;
        }
        if(node.kind === 'PostfixUnaryExpression') {
            return 17;
        }
        if(node.kind === 'CallExpression') {
            return 18;
        }
        if(node.kind === 'NewExpression') {
            return 19;
        }
        if(['AsExpression', 'SatisfiesExpression'].includes(node.kind)) {
            return 2;
        }
        if(
            [
                'Identifier',
                'ThisKeyword',
                'SuperKeyword',
                'StringLiteral',
                'NumericLiteral',
                'BigIntLiteral',
                'TrueKeyword',
                'FalseKeyword',
                'NullKeyword',
                'RegularExpressionLiteral',
                'NoSubstitutionTemplateLiteral',
                'TemplateExpression',
                'ArrayLiteralExpression',
                'ObjectLiteralExpression',
                'PropertyAccessExpression',
                'ElementAccessExpression',
                'TaggedTemplateExpression',
                'FunctionExpression',
                'ClassExpression',
            ].includes(node.kind)
        ) {
            return 20;
        }
        return -1;
    }
    invert(index: number): string {
        if(this.context.node(index).kind === 'BinaryExpression') {
            const operator = this.context.node(index).children[1] ?? panic('operator');
            const spelling = this.context.raw(operator);
            const inverse =
                spelling === '=='
                    ? '!='
                    : spelling === '!='
                      ? '=='
                      : spelling === '==='
                        ? '!=='
                        : spelling === '!=='
                          ? '==='
                          : '';
            if(inverse !== '') {
                return (
                    this.context.source.slice(this.context.start(index), this.context.start(operator)) +
                    inverse +
                    this.context.source.slice(this.context.node(operator).end, this.context.node(index).end)
                );
            }
        }
        return this.precedence(index) < 16 ? `!(${this.context.raw(index)})` : `!${this.context.raw(index)}`;
    }
    alwaysBoolean(index: number): boolean {
        const node = this.context.node(this.context.unwrap(index));
        return node.kind === 'PrefixUnaryExpression'
            ? node.operator === 'ExclamationToken'
            : node.kind === 'BinaryExpression' &&
                  ['==', '!=', '===', '!==', '<', '<=', '>', '>=', 'in', 'instanceof'].includes(
                      this.context.raw(node.children[1] ?? panic('operator')),
                  );
    }
}

export function create(context: RuleContext): Rule {
    return new Rule(context);
}
