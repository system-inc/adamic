// The corpus carries decoded Go options as JSON. Reuse the source parser for
// this restricted data grammar; never execute option text.
import { panic } from 'adamic';
import { Parser } from '../../typescript/parser/parser.ts';

export class Settings {
    text = '';
    readonly values = new Map<string, string[]>();
    load(text: string): void {
        this.text = text;
        if(text === '' || text === 'null') {
            return;
        }
        const parser = new Parser(`(${text});`, 'decoded rule options');
        const root = parser.file();
        const statement = parser.node(root).children[0] ?? panic('missing option statement');
        const outer = parser.node(statement).children[0] ?? panic('missing options');
        const object = parser.node(outer).children[0] ?? panic('missing option object');
        if(parser.node(object).kind === 'StringLiteral') {
            this.values.set('option', [parser.node(object).text]);
            return;
        }
        if(parser.node(object).kind !== 'ObjectLiteralExpression') {
            this.values.set('option', [text]);
            return;
        }
        for(const property of parser.node(object).children) {
            const children = parser.node(property).children;
            const name = parser.node(children[0] ?? panic('missing option key')).text.toLowerCase();
            const value = children[1] ?? panic('missing option value');
            if(parser.node(value).kind === 'NullKeyword') {
                continue;
            }
            const values: string[] = [];
            const elements =
                parser.node(value).kind === 'ArrayLiteralExpression' ? parser.node(value).children : [value];
            for(const element of elements) {
                const node = parser.node(element);
                const scalar = ['TrueKeyword', 'FalseKeyword', 'StringLiteral', 'NumericLiteral'].includes(node.kind);
                if(
                    !scalar &&
                    ![
                        'ObjectLiteralExpression',
                        'ArrayLiteralExpression',
                        'PrefixUnaryExpression',
                        'NullKeyword',
                    ].includes(node.kind)
                ) {
                    panic('unsupported option expression');
                }
                // Structured values stay as data for a rule's own decoder.
                values.push(
                    scalar
                        ? node.kind === 'TrueKeyword'
                            ? 'true'
                            : node.kind === 'FalseKeyword'
                              ? 'false'
                              : node.text
                        : `(${text});`.slice(node.pos, node.end).trim(),
                );
            }
            this.values.set(name, values);
        }
    }
    read(name: string, fallback: string): string {
        const values = this.values.get(name);
        if(values === undefined) {
            return fallback;
        }
        return values[0] ?? fallback;
    }
    list(name: string, fallback: string[]): string[] {
        return this.values.get(name) ?? fallback;
    }
}
