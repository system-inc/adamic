import { panic } from 'adamic';
import type { Arena } from './arena.ts';
import type { ParseNode } from '../../typescript/parser/nodes.ts';
import { absent, boolValue, childValue, listValue, stringValue } from './values.ts';
import { jsxText } from './jsxEntities.ts';
export function convertJsx(
    arena: Arena,
    nodes: readonly ParseNode[],
    offsets: readonly number[],
    id: number,
    convert: (id: number) => number,
): number {
    const source = nodes[id] ?? panic('missing JSX source');
    const first = source.children[0] ?? -1;
    const last = source.children[source.children.length - 1] ?? -1;
    const type =
        source.kind === 'JsxTypeArguments'
            ? 'TSTypeParameterInstantiation'
            : source.kind === 'JsxName'
              ? 'JSXIdentifier'
              : source.kind === 'JsxMemberName'
                ? 'JSXMemberExpression'
                : source.kind === 'JsxNamespacedName'
                  ? 'JSXNamespacedName'
                  : source.kind === 'JsxString'
                    ? 'Literal'
                    : source.kind.slice(0, 3).toUpperCase() + source.kind.slice(3);
    const result = arena.newNode(
        type,
        offsets[source.pos] ?? panic('JSX start outside source'),
        offsets[source.end] ?? panic('JSX end outside source'),
    );
    const node = arena.node(result);
    if(source.kind === 'JsxTypeArguments') {
        const params: number[] = [];
        for(const parameter of source.children) {
            params.push(convert(parameter));
        }
        node.set('params', listValue(params));
    }
    else if(source.kind === 'JsxName') {
        node.set('name', stringValue(source.text));
    }
    else if(source.kind === 'JsxMemberName') {
        node.set('object', childValue(convert(first)));
        node.set('property', childValue(convert(last)));
    }
    else if(source.kind === 'JsxNamespacedName') {
        node.set('name', childValue(convert(last)));
        node.set('namespace', childValue(convert(first)));
    }
    else if(source.kind === 'JsxString') {
        node.set('raw', stringValue(source.raw));
        node.set('value', stringValue(jsxText(source.text)));
    }
    else if(source.kind === 'JsxText') {
        node.set('raw', stringValue(source.text));
        node.set('value', stringValue(jsxText(source.text)));
    }
    else if(source.kind === 'JsxExpression' || source.kind === 'JsxSpreadChild') {
        node.type = source.kind === 'JsxExpression' ? 'JSXExpressionContainer' : 'JSXSpreadChild';
        const empty =
            first < 0
                ? arena.newNode('JSXEmptyExpression', (offsets[source.pos] ?? -1) + 1, (offsets[source.end] ?? -1) - 1)
                : -1;
        node.set('expression', childValue(first < 0 ? empty : convert(first)));
    }
    else if(source.kind === 'JsxSpreadAttribute') {
        node.set('argument', childValue(convert(first)));
    }
    else if(source.kind === 'JsxAttribute') {
        node.set('name', childValue(convert(first)));
        node.set('value', source.children.length < 2 ? absent() : childValue(convert(last)));
    }
    else if(source.kind === 'JsxOpeningElement') {
        const attributes: number[] = [];
        for(const attribute of source.children.slice(source.list + 1)) {
            attributes.push(convert(attribute));
        }
        node.set('attributes', listValue(attributes));
        node.set('name', childValue(convert(first)));
        node.set('selfClosing', boolValue(source.optional));
        node.set('typeArguments', source.list === 0 ? absent() : childValue(convert(source.children[1] ?? -1)));
    }
    else if(source.kind === 'JsxClosingElement') {
        node.set('name', childValue(convert(first)));
    }
    else if(source.kind === 'JsxElement' || source.kind === 'JsxFragment') {
        const children: number[] = [];
        for(const child of source.children.slice(1, source.children.length > 1 ? -1 : 1)) {
            children.push(convert(child));
        }
        node.set('children', listValue(children));
        node.set(
            source.kind === 'JsxFragment' ? 'closingFragment' : 'closingElement',
            source.children.length > 1 ? childValue(convert(last)) : absent(),
        );
        node.set(source.kind === 'JsxFragment' ? 'openingFragment' : 'openingElement', childValue(convert(first)));
    }
    else if(source.kind !== 'JsxOpeningFragment' && source.kind !== 'JsxClosingFragment') {
        return panic(`JSX conversion missing ${source.kind}`);
    }
    return result;
}
