import { leftHandSide } from './sourceLookahead.ts';
import { reservedKinds } from './sourceGrammar.ts';
import type { ParseNode } from '../../typescript/parser/nodes.ts';

export function syntaxError(nodes: readonly ParseNode[]): string {
    const parents: number[] = [];
    for(let remaining = nodes.length; remaining > 0; remaining--) {
        parents.push(-1);
    }
    for(let index = 0; index < nodes.length; index++) {
        const node = nodes[index];
        if(node === undefined) {
            continue;
        }
        for(const child of node.children) {
            parents[child] = index;
        }
    }
    for(let index = 0; index < nodes.length; index++) {
        const node = nodes[index];
        if(node === undefined) {
            continue;
        }
        const first = nodes[node.children[0] ?? -1];
        const parent = nodes[parents[index] ?? -1];

        if(
            node.kind === 'QualifiedName' &&
            nodes[node.children[1] ?? -1]?.kind === 'PrivateIdentifier' &&
            parent?.kind === 'TypeReference'
        ) {
            return 'private name in a type reference';
        }
        if(
            ['UnionType', 'IntersectionType'].includes(node.kind) &&
            node.children.some((id) => ['FunctionType', 'ConstructorType'].includes(nodes[id]?.kind ?? ''))
        ) {
            return 'function or constructor type must be parenthesized in a union or intersection';
        }
        if(node.kind === 'TypeAssertionExpression' && nodes[node.children[1] ?? -1]?.kind === 'YieldExpression') {
            return 'yield is not a simple unary expression after a type assertion';
        }
        if(node.kind === 'ImportSpecifier') {
            const name = nodes[node.children[node.children.length - 1] ?? -1];
            if(
                name?.kind === 'StringLiteral' ||
                reservedKinds.some((kind) => kind.slice(0, -7).toLowerCase() === name?.text)
            ) {
                return 'an import binding must be an identifier';
            }
        }
        if(
            node.kind === 'PrefixUnaryExpression' &&
            ['PlusPlusToken', 'MinusMinusToken'].includes(node.operator) &&
            !leftHandSide(first?.kind ?? '')
        ) {
            return 'update operand must be a left hand side expression';
        }
        if(
            node.kind === 'BinaryExpression' &&
            nodes[node.children[1] ?? -1]?.kind === 'AsteriskAsteriskToken' &&
            first !== undefined
        ) {
            if(
                [
                    'DeleteExpression',
                    'TypeOfExpression',
                    'VoidExpression',
                    'AwaitExpression',
                    'TypeAssertionExpression',
                ].includes(first.kind) ||
                (first.kind === 'PrefixUnaryExpression' &&
                    !['PlusPlusToken', 'MinusMinusToken'].includes(first.operator))
            ) {
                return 'unparenthesized unary expression before exponentiation';
            }
        }
        if(node.kind === 'PropertyAccessExpression') {
            if(first?.kind === 'ExpressionWithTypeArguments') {
                return 'property access after an instantiation expression';
            }
            if(node.optional && nodes[node.children[node.children.length - 1] ?? -1]?.kind === 'PrivateIdentifier') {
                return 'private identifier in an optional chain';
            }
        }
        if(node.kind === 'ExpressionWithTypeArguments' && first?.kind === 'SuperKeyword') {
            return 'super may not use type arguments';
        }
        if(['CallExpression', 'TaggedTemplateExpression'].includes(node.kind) && first?.kind === 'SuperKeyword') {
            const count = node.kind === 'CallExpression' ? node.list : 1;
            if(node.children.length > count + 1) {
                return 'super may not use type arguments';
            }
        }
        if(
            node.kind === 'SuperKeyword' &&
            (parent === undefined ||
                ![
                    'CallExpression',
                    'PropertyAccessExpression',
                    'ElementAccessExpression',
                    'ExpressionWithTypeArguments',
                    'NewExpression',
                ].includes(parent.kind))
        ) {
            return 'super requires arguments or member access';
        }
        if(node.kind === 'NewExpression' && first?.optional === true) {
            return 'invalid optional chain from new expression';
        }
        if(
            ['NoSubstitutionTemplateLiteral', 'TemplateHead', 'TemplateMiddle', 'TemplateTail'].includes(node.kind) &&
            (node.literalFlags & 2048) !== 0
        ) {
            let ownerId = parents[index] ?? -1;
            if(nodes[ownerId]?.kind === 'TemplateSpan') {
                ownerId = parents[ownerId] ?? -1;
            }
            if(nodes[ownerId]?.kind === 'TemplateExpression') {
                ownerId = parents[ownerId] ?? -1;
            }
            const owner = nodes[ownerId];
            if(owner?.kind !== 'TaggedTemplateExpression') {
                return 'invalid escape in an untagged template';
            }
        }
    }
    return '';
}
