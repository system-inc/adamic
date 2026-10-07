// Precedence and required parentheses over the parser's indexed tree.
import { panic } from 'adamic';
import type { Parser } from '../../typescript/parser/parser.ts';
import {
    unwrapped,
    syntaxNode,
    syntaxChild,
    syntaxOperator,
    syntaxAssignment,
    syntaxOwnOptional,
    syntaxMemberish,
    syntaxLeftmost,
    rank,
    flatten,
} from './syntax.ts';
export function parenthesize(
    parser: Parser,
    statementRoot: number,
    ancestors: readonly number[],
    optionalBoundaries: Set<number>,
    index: number,
    parent: number,
    role: string,
): boolean {
    const node = syntaxNode(parser, index);
    if(optionalBoundaries.has(index) && parent >= 0) {
        const outer = syntaxNode(parser, parent);
        if(
            outer.kind === 'NonNullExpression' ||
            role === 'tag' ||
            (role === 'object' && syntaxMemberish(parser, parent) && !syntaxOwnOptional(parser, parent)) ||
            (role === 'callee' &&
                (outer.kind === 'NewExpression' ||
                    (outer.kind === 'CallExpression' && !syntaxOwnOptional(parser, parent))))
        )
            return true;
    }
    if(node.kind === 'FunctionExpression') {
        const root = statementRoot >= 0 ? statementRoot : (ancestors[0] ?? index);
        return syntaxLeftmost(parser, root) === index || role === 'callee' || role === 'tag';
    }
    if(node.kind === 'ObjectLiteralExpression') {
        const root = statementRoot >= 0 ? statementRoot : (ancestors[0] ?? index);
        if(syntaxLeftmost(parser, root) === index) return true;
        for(let position = ancestors.length - 1; position >= 0; position--) {
            const ancestor = ancestors[position] ?? panic('missing object ancestor');
            const item = syntaxNode(parser, ancestor);
            if(item.kind !== 'ArrowFunction') continue;
            const body = unwrapped(parser, item.children[item.children.length - 1] ?? panic('missing arrow body'));
            return (
                !syntaxAssignment(parser, body) &&
                !(syntaxNode(parser, body).kind === 'BinaryExpression' && syntaxOperator(parser, body) === ',') &&
                syntaxLeftmost(parser, body) === index
            );
        }
        return false;
    }
    if(parent < 0) return false;
    const outer = syntaxNode(parser, parent);
    if(node.kind === 'AwaitExpression' || node.kind === 'YieldExpression')
        return (
            role === 'tag' ||
            role === 'object' ||
            role === 'callee' ||
            (outer.kind === 'ConditionalExpression' && role === 'test') ||
            (outer.kind === 'BinaryExpression' && !syntaxAssignment(parser, parent)) ||
            (node.kind === 'YieldExpression' && outer.kind === 'AwaitExpression') ||
            [
                'PrefixUnaryExpression',
                'DeleteExpression',
                'VoidExpression',
                'TypeOfExpression',
                'SpreadElement',
                'SpreadAssignment',
                'NonNullExpression',
            ].includes(outer.kind)
        );
    if(node.kind === 'ArrowFunction')
        return (
            role === 'callee' ||
            role === 'object' ||
            role === 'tag' ||
            (outer.kind === 'BinaryExpression' &&
                !syntaxAssignment(parser, parent) &&
                syntaxOperator(parser, parent) !== ',') ||
            (outer.kind === 'ConditionalExpression' && role === 'test') ||
            [
                'PrefixUnaryExpression',
                'PostfixUnaryExpression',
                'DeleteExpression',
                'VoidExpression',
                'TypeOfExpression',
                'NonNullExpression',
                'AwaitExpression',
            ].includes(outer.kind)
        );
    if(outer.kind === 'NewExpression' && role === 'callee') {
        let current = index;
        while(
            [
                'PropertyAccessExpression',
                'ElementAccessExpression',
                'NonNullExpression',
                'AwaitExpression',
                'TaggedTemplateExpression',
            ].includes(syntaxNode(parser, current).kind)
        )
            current = syntaxChild(parser, current, 0);
        if(syntaxNode(parser, current).kind === 'CallExpression') return true;
    }
    if(
        node.kind === 'Identifier' &&
        node.text === 'let' &&
        role === 'object' &&
        outer.kind === 'ElementAccessExpression' &&
        !outer.optional &&
        syntaxLeftmost(parser, statementRoot >= 0 ? statementRoot : (ancestors[0] ?? index)) === index
    )
        return true;
    if(node.kind === 'NumericLiteral')
        return role === 'object' && ['PropertyAccessExpression', 'ElementAccessExpression'].includes(outer.kind);
    const unary = [
        'PrefixUnaryExpression',
        'PostfixUnaryExpression',
        'DeleteExpression',
        'VoidExpression',
        'TypeOfExpression',
    ].includes(node.kind);
    if(unary) {
        const operator = syntaxOperator(parser, index);
        const update = node.kind === 'PostfixUnaryExpression' || ['++', '--'].includes(operator);
        if(outer.kind === 'PrefixUnaryExpression' && !['++', '--'].includes(syntaxOperator(parser, parent))) {
            if(update)
                return (
                    node.kind !== 'PostfixUnaryExpression' &&
                    ((operator === '++' && syntaxOperator(parser, parent) === '+') ||
                        (operator === '--' && syntaxOperator(parser, parent) === '-'))
                );
            return operator === syntaxOperator(parser, parent) && ['+', '-'].includes(operator);
        }
        if(outer.kind === 'BinaryExpression')
            return (
                role === 'left' &&
                (syntaxOperator(parser, parent) === '**' ||
                    (!update && ['in', 'instanceof'].includes(syntaxOperator(parser, parent))))
            );
        return role === 'object' || role === 'callee' || role === 'tag' || outer.kind === 'NonNullExpression';
    }
    if(node.kind === 'ConditionalExpression')
        return (
            role === 'object' ||
            role === 'callee' ||
            role === 'tag' ||
            (outer.kind === 'ConditionalExpression' && role === 'test') ||
            (outer.kind === 'BinaryExpression' &&
                !syntaxAssignment(parser, parent) &&
                syntaxOperator(parser, parent) !== ',') ||
            [
                'PrefixUnaryExpression',
                'PostfixUnaryExpression',
                'DeleteExpression',
                'VoidExpression',
                'TypeOfExpression',
                'NonNullExpression',
                'AwaitExpression',
                'SpreadElement',
                'SpreadAssignment',
            ].includes(outer.kind)
        );
    if(node.kind === 'BinaryExpression') {
        if(syntaxAssignment(parser, index)) return !syntaxAssignment(parser, parent);
        if(syntaxOperator(parser, index) === ',') return true;
        if(syntaxAssignment(parser, parent)) return false;
        if(outer.kind === 'ConditionalExpression') return syntaxOperator(parser, index) === '??';
        if(outer.kind === 'BinaryExpression') {
            const operator = syntaxOperator(parser, index);
            const other = syntaxOperator(parser, parent);
            if(['??', '||', '&&'].includes(operator) && ['??', '||', '&&'].includes(other)) return operator !== other;
            if(rank(other) > rank(operator)) return true;
            if(rank(other) === rank(operator) && (role === 'right' || !flatten(other, operator))) return true;
            if(operator === '%' && (other === '+' || other === '-')) return true;
            return ['|', '^', '&', '<<', '>>', '>>>'].includes(other);
        }
        return (
            role === 'object' ||
            role === 'callee' ||
            role === 'tag' ||
            [
                'PrefixUnaryExpression',
                'PostfixUnaryExpression',
                'DeleteExpression',
                'VoidExpression',
                'TypeOfExpression',
                'NonNullExpression',
                'AwaitExpression',
                'SpreadElement',
                'SpreadAssignment',
            ].includes(outer.kind)
        );
    }
    return false;
}
