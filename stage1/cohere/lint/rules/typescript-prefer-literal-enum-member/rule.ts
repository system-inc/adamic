import { panic } from 'adamic';
import type { RuleContext } from '../../context.ts';
import type { ParseNode } from '../../../../typescript/parser/nodes.ts';
import {
    messagePreferLiteralEnumMemberNotLiteral,
    messagePreferLiteralEnumMemberNotLiteralOrBitwise,
} from './messages.ts';

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) {
        this.context = context;
    }
    // An initialized member reports at its name unless the initializer is a literal, or with the option a
    // bitwise expression over literals and the enum's own members.
    visit(node: ParseNode, index: number): void {
        if(node.children.length <= 1) {
            return;
        }
        const owner = this.context.parent(index);
        const initializer = node.children[1] ?? panic('enum initializer');
        const bitwise = this.context.settings.read('allowbitwiseexpressions', 'false') === 'true';
        if(owner >= 0 && !this.enumAllowed(owner, initializer, false, bitwise)) {
            this.context.report(
                node.children[0] ?? panic('enum name'),
                '@typescript-eslint/prefer-literal-enum-member',
                bitwise ? 'notLiteralOrBitwiseExpression' : 'notLiteral',
                bitwise ? messagePreferLiteralEnumMemberNotLiteralOrBitwise : messagePreferLiteralEnumMemberNotLiteral,
                '',
                '',
                '',
            );
        }
    }
    // enumSelf is whether the expression names a member of the declaring enum, bare or through the enum.
    enumSelf(declaration: number, index: number): boolean {
        const node = this.context.node(index);
        let name = '';
        if(node.kind === 'Identifier') {
            name = node.text;
        }
        else if(node.kind === 'PropertyAccessExpression' || node.kind === 'ElementAccessExpression') {
            const receiver = node.children[0] ?? panic('enum receiver');
            const enumName = this.context.name(declaration);
            if(
                enumName < 0 ||
                this.context.node(receiver).kind !== 'Identifier' ||
                this.context.node(receiver).text !== this.context.node(enumName).text
            ) {
                return false;
            }
            const key = node.children[node.children.length - 1] ?? panic('enum key');
            const kind = this.context.node(key).kind;
            if(
                node.kind === 'PropertyAccessExpression'
                    ? kind !== 'Identifier'
                    : kind !== 'StringLiteral' && kind !== 'NoSubstitutionTemplateLiteral'
            ) {
                return false;
            }
            name = this.context.node(key).text;
        }
        if(name === '') {
            return false;
        }
        return this.context.node(declaration).children.some((child) => {
            if(this.context.node(child).kind !== 'EnumMember') {
                return false;
            }
            const key = this.context.node(child).children[0] ?? panic('enum member name');
            return (
                ['Identifier', 'StringLiteral', 'NoSubstitutionTemplateLiteral'].includes(
                    this.context.node(key).kind,
                ) && this.context.node(key).text === name
            );
        });
    }
    enumAllowed(declaration: number, index: number, part: boolean, bitwise: boolean): boolean {
        const unwrapped = this.context.unwrap(index);
        const node = this.context.node(unwrapped);
        if(part && this.enumSelf(declaration, unwrapped)) {
            return true;
        }
        if(
            [
                'StringLiteral',
                'NumericLiteral',
                'BigIntLiteral',
                'RegularExpressionLiteral',
                'NullKeyword',
                'TrueKeyword',
                'FalseKeyword',
                'NoSubstitutionTemplateLiteral',
            ].includes(node.kind)
        ) {
            return true;
        }
        if(node.kind === 'PrefixUnaryExpression') {
            const operand = node.children[0] ?? panic('enum operand');
            if(node.operator === 'PlusToken' || node.operator === 'MinusToken') {
                return this.enumAllowed(declaration, operand, part, bitwise);
            }
            return bitwise && node.operator === 'TildeToken' && this.enumAllowed(declaration, operand, true, bitwise);
        }
        if(node.kind === 'BinaryExpression') {
            const operator = this.context.node(node.children[1] ?? panic('enum operator')).kind;
            return (
                bitwise &&
                [
                    'AmpersandToken',
                    'CaretToken',
                    'LessThanLessThanToken',
                    'GreaterThanGreaterThanToken',
                    'GreaterThanGreaterThanGreaterThanToken',
                    'BarToken',
                ].includes(operator) &&
                this.enumAllowed(declaration, node.children[0] ?? panic('enum left'), true, bitwise) &&
                this.enumAllowed(declaration, node.children[2] ?? panic('enum right'), true, bitwise)
            );
        }
        return false;
    }
}

export function create(context: RuleContext): Rule {
    return new Rule(context);
}
