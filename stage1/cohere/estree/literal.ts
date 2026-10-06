import type { ParseNode } from '../../typescript/parser/nodes.ts';
import { Property } from './property.ts';
import { Value, absent, boolValue, stringValue, numberValue, goNumber, bigintDecimal } from './values.ts';
export const literalKinds = [
    'StringLiteral',
    'NumericLiteral',
    'BigIntLiteral',
    'RegularExpressionLiteral',
    'TrueKeyword',
    'FalseKeyword',
    'NullKeyword',
];
export function literalFields(node: ParseNode, raw: string): Property[] {
    const properties: Property[] = [];

    if(node.kind === 'BigIntLiteral') {
        properties.push(new Property('bigint', stringValue(bigintDecimal(raw))));
    }
    properties.push(new Property('raw', stringValue(raw)));
    if(node.kind === 'RegularExpressionLiteral') {
        const value = new Value('regex');
        const slash = raw.lastIndexOf('/');
        value.text = raw.slice(1, slash);
        value.cooked = raw.slice(slash + 1);
        properties.push(new Property('regex', value));
    }
    properties.push(
        new Property(
            'value',
            node.kind === 'StringLiteral'
                ? stringValue(node.text)
                : node.kind === 'NumericLiteral'
                  ? numberValue(goNumber(raw))
                  : node.kind === 'TrueKeyword' || node.kind === 'FalseKeyword'
                    ? boolValue(node.kind === 'TrueKeyword')
                    : absent(),
        ),
    );

    return properties;
}
