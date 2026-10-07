import { panic } from 'adamic';
import { ParseNode } from '../../typescript/parser/nodes.ts';
import type { Scanner } from '../../typescript/scanner/scanner.ts';
class JsxParser {
    readonly scanner: Scanner;
    readonly nodes: ParseNode[];
    readonly expression: () => number;
    readonly types: () => number[];
    constructor(scanner: Scanner, nodes: ParseNode[], expression: () => number, types: () => number[]) {
        this.scanner = scanner;
        this.nodes = nodes;
        this.expression = expression;
        this.types = types;
    }
    add(kind: string, start: number, end: number, children: number[] = [], text = ''): number {
        this.nodes.push(new ParseNode(kind, start, end, children));
        const id = this.nodes.length - 1;
        const node = this.nodes[id] ?? panic('missing JSX node');
        node.text = text;
        return id;
    }
    leaf(kind: string, start: number, end: number, text: string): number {
        const empty: number[] = [];
        return this.add(kind, start, end, empty, text);
    }
    kind(): string {
        return this.scanner.kind;
    }
    expect(kind: string): void {
        if(this.kind() !== kind) {
            panic(`JSX expected ${kind}, got ${this.kind()}`);
        }
    }
    identifier(): number {
        // eslint-disable-next-line nexus/consistency-no-property-alias -- Snapshot the token position before scans mutate it.
        const start = this.scanner.start;
        if(this.kind() !== 'Identifier' && !this.kind().endsWith('Keyword')) {
            return panic('JSX tag name expected');
        }
        let end = this.scanner.pos;
        while(this.scanner.text.slice(end, end + 1) === '-') {
            end++;
            while(/^[\p{ID_Continue}$]$/u.test(this.scanner.text.slice(end, end + 1))) {
                end++;
            }
        }
        this.scanner.pos = end;
        const text = this.scanner.text.slice(start, end);
        if(text.includes('\\')) {
            return panic('Unicode escape in JSX name');
        }
        const name = this.leaf('JsxName', start, end, text);
        this.scanner.scan();
        return name;
    }
    name(): number {
        // eslint-disable-next-line nexus/consistency-no-property-alias -- Snapshot the token position before scans mutate it.
        const start = this.scanner.start;
        let name = this.identifier();
        if(this.kind() === 'ColonToken') {
            this.scanner.scan();
            const right = this.identifier();
            return this.add('JsxNamespacedName', start, this.nodes[right]?.end ?? -1, [name, right]);
        }
        while(this.kind() === 'DotToken') {
            this.scanner.scan();
            const right = this.identifier();
            name = this.add('JsxMemberName', start, this.nodes[right]?.end ?? -1, [name, right]);
        }
        return name;
    }
    container(child: boolean, attributeSpread = false): number {
        // eslint-disable-next-line nexus/consistency-no-property-alias -- Snapshot the token position before scans mutate it.
        const start = this.scanner.start;
        this.expect('OpenBraceToken');
        this.scanner.scan();
        const spread = this.kind() === 'DotDotDotToken';
        if(!child && spread !== attributeSpread) {
            return panic('invalid JSX attribute expression');
        }
        if(spread) {
            this.scanner.scan();
        }
        const parts: number[] = [];
        if(this.kind() !== 'CloseBraceToken') {
            parts.push(this.expression());
        }
        this.expect('CloseBraceToken');
        const end = this.scanner.pos;
        if(child) {
            this.scanner.scanJsx();
        }
        else {
            this.scanner.scan();
        }
        return this.add(
            spread ? (child ? 'JsxSpreadChild' : 'JsxSpreadAttribute') : 'JsxExpression',
            start,
            end,
            parts,
        );
    }
    parse(child: boolean): number {
        // eslint-disable-next-line nexus/consistency-no-property-alias -- Snapshot the token position before scans mutate it.
        const start = this.scanner.start;
        this.expect('LessThanToken');
        this.scanner.scan();
        const fragment = this.kind() === 'GreaterThanToken';
        const attributes: number[] = [];
        let typeCount = 0;
        if(!fragment) {
            attributes.push(this.name());
            if(this.kind() === 'LessThanToken') {
                const typeStart = this.scanner.start;
                const types = this.types();
                attributes.push(
                    types.length === 1 && this.nodes[types[0] ?? -1]?.kind === 'EmptyTypeArguments'
                        ? (types[0] ?? -1)
                        : this.add('JsxTypeArguments', typeStart, this.scanner.fullStart, types),
                );
                typeCount = 1;
            }
            while(this.kind() !== 'SlashToken' && this.kind() !== 'GreaterThanToken') {
                if(this.kind() === 'EndOfFile') {
                    return panic('unterminated JSX opening element');
                }
                if(this.kind() === 'OpenBraceToken') {
                    attributes.push(this.container(false, true));
                    continue;
                }
                const attributeStart = this.scanner.start;
                const name = this.name();
                const fields = [name];
                if(this.kind() === 'EqualsToken') {
                    const errors = this.scanner.errors.length;
                    this.scanner.scan();
                    if(this.kind() === 'StringLiteral') {
                        this.scanner.errors.splice(errors);
                        const begin = this.scanner.start;
                        const quote = this.scanner.text.slice(begin, begin + 1);
                        const end = this.scanner.text.indexOf(quote, begin + 1);
                        if(end < 0) {
                            return panic('unterminated JSX attribute');
                        }
                        const literal = this.leaf('JsxString', begin, end + 1, this.scanner.text.slice(begin + 1, end));
                        const literalNode = this.nodes[literal] ?? panic('missing JSX literal');
                        literalNode.raw = this.scanner.text.slice(begin, end + 1);
                        fields.push(literal);
                        this.scanner.pos = end + 1;
                        this.scanner.scan();
                    }
                    else if(this.kind() === 'OpenBraceToken') {
                        fields.push(this.container(false));
                    }
                    else if(this.kind() === 'LessThanToken') {
                        fields.push(this.parse(false));
                    }
                    else {
                        return panic('JSX attribute value expected');
                    }
                }
                const last = fields[fields.length - 1] ?? -1;
                attributes.push(this.add('JsxAttribute', attributeStart, this.nodes[last]?.end ?? -1, fields));
            }
        }
        const selfClosing = this.kind() === 'SlashToken';
        if(selfClosing) {
            this.scanner.scan();
        }
        this.expect('GreaterThanToken');
        const openEnd = this.scanner.pos;
        const opening = this.add(fragment ? 'JsxOpeningFragment' : 'JsxOpeningElement', start, openEnd, attributes);
        const openingNode = this.nodes[opening] ?? panic('missing opening');
        openingNode.optional = selfClosing;
        openingNode.list = typeCount;
        if(selfClosing) {
            if(child) {
                this.scanner.scanJsx();
            }
            else {
                this.scanner.scan();
            }
            return this.add('JsxElement', start, openEnd, [opening]);
        }
        this.scanner.scanJsx();
        const children = [opening];
        while(this.kind() !== 'LessThanSlashToken') {
            if(this.kind() === 'EndOfFile') {
                return panic('unterminated JSX element');
            }
            if(this.kind() === 'LessThanToken') {
                children.push(this.parse(true));
            }
            else if(this.kind() === 'OpenBraceToken') {
                children.push(this.container(true));
            }
            else {
                children.push(this.leaf('JsxText', this.scanner.start, this.scanner.pos, this.scanner.value));
                this.scanner.scanJsx();
            }
        }
        const closeStart = this.scanner.start;
        this.scanner.scan();
        const closingParts: number[] = [];
        if(!fragment) {
            closingParts.push(this.name());
        }
        this.expect('GreaterThanToken');
        const end = this.scanner.pos;
        children.push(this.add(fragment ? 'JsxClosingFragment' : 'JsxClosingElement', closeStart, end, closingParts));
        if(child) {
            this.scanner.scanJsx();
        }
        else {
            this.scanner.scan();
        }
        return this.add(fragment ? 'JsxFragment' : 'JsxElement', start, end, children);
    }
}
export function readJsx(scanner: Scanner, nodes: ParseNode[], expression: () => number, types: () => number[]): number {
    return new JsxParser(scanner, nodes, expression, types).parse(false);
}
