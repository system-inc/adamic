// JSX descent and scanner transitions follow typescript-go's parser.go.
import { panic } from 'adamic';
import type { Scanner } from '../scanner/scanner.ts';
import type { ParseNode } from './nodes.ts';

export interface JsxContextInterface {
    readonly scanner: Scanner;
    readonly path: string;
    readonly javascript: boolean;
    readonly kind: () => string;
    readonly next: () => void;
    readonly expect: (kind: string) => void;
    readonly node: (index: number) => ParseNode;
    readonly make: (kind: string, pos: number, children: number[]) => number;
    readonly identifier: () => number;
    readonly token: () => number;
    readonly literal: () => number;
    readonly expression: () => number;
    readonly typeArguments: () => number[];
    readonly missingGreater: () => void;
    readonly typeTrailing: () => boolean;
}

export class Jsx {
    readonly parser: JsxContextInterface;
    constructor(parser: JsxContextInterface) {
        this.parser = parser;
    }
    make(kind: string, pos: number, children: number[] = []): number {
        return this.parser.make(kind, pos, children);
    }
    identifier(): number {
        if(this.parser.kind() !== 'Identifier' && !this.parser.kind().endsWith('Keyword')) {
            panic(`JSX expected name at ${this.parser.scanner.start} in ${this.parser.path}`);
        }
        if((this.parser.scanner.flags & 1032) !== 0) {
            panic(`JSX Unicode escape not allowed at ${this.parser.scanner.start} in ${this.parser.path}`);
        }
        return this.parser.identifier();
    }
    name(tag: boolean): number {
        const pos = this.parser.scanner.fullStart;
        this.parser.scanner.scanJsxIdentifier();
        const isThis = tag && this.parser.kind() === 'ThisKeyword';
        let name = this.identifier();
        if(this.parser.kind() === 'ColonToken') {
            this.parser.next();
            this.parser.scanner.scanJsxIdentifier();
            const right = this.identifier();
            return this.make('JsxNamespacedName', pos, [name, right]);
        }
        if(isThis) {
            name = this.make('ThisKeyword', pos);
        }
        if(tag) {
            while(this.parser.kind() === 'DotToken') {
                this.parser.next();
                const right = this.identifier();
                name = this.make('PropertyAccessExpression', pos, [name, right]);
            }
        }
        return name;
    }
    text(): number {
        const pos = this.parser.scanner.fullStart;
        const whitespace = this.parser.kind() === 'JsxTextAllWhiteSpaces';
        const id = this.make('JsxText', pos);
        this.parser.node(id).text = this.parser.scanner.value;
        this.parser.node(id).semantic = whitespace ? '1' : '0';
        this.parser.scanner.scanJsx();
        this.parser.node(id).end = this.parser.scanner.fullStart;
        return id;
    }
    finishGreater(expression: boolean): void {
        if(this.parser.kind() !== 'GreaterThanToken') {
            // Go leaves the next token in place when the closing delimiter is missing.
            this.parser.missingGreater();
            return;
        }
        if(expression) {
            this.parser.next();
        }
        else {
            this.parser.scanner.scanJsx();
        }
    }
    expression(attribute: boolean): number {
        const pos = this.parser.scanner.fullStart;
        this.parser.expect('OpenBraceToken');
        const children: number[] = [];
        if(this.parser.kind() !== 'CloseBraceToken') {
            if(!attribute && this.parser.kind() === 'DotDotDotToken') {
                children.push(this.parser.token());
            }
            children.push(this.parser.expression());
        }
        if(attribute) {
            this.parser.expect('CloseBraceToken');
        }
        else {
            if(this.parser.kind() !== 'CloseBraceToken') {
                panic(`JSX expected } at ${this.parser.scanner.start} in ${this.parser.path}`);
            }
            this.parser.scanner.scanJsx();
        }
        return this.make('JsxExpression', pos, children);
    }
    attributes(): number {
        const pos = this.parser.scanner.fullStart;
        const attributes: number[] = [];
        while(this.parser.kind() !== 'GreaterThanToken' && this.parser.kind() !== 'SlashToken') {
            const start = this.parser.scanner.fullStart;
            if(this.parser.kind() === 'OpenBraceToken') {
                this.parser.next();
                this.parser.expect('DotDotDotToken');
                const expression = this.parser.expression();
                this.parser.expect('CloseBraceToken');
                attributes.push(this.make('JsxSpreadAttribute', start, [expression]));
                continue;
            }
            const children = [this.name(false)];
            if(this.parser.kind() === 'EqualsToken') {
                this.parser.scanner.scanJsxAttributeValue();
                if(this.parser.kind() === 'StringLiteral') {
                    children.push(this.parser.literal());
                }
                else if(this.parser.kind() === 'OpenBraceToken') {
                    children.push(this.expression(true));
                }
                else if(this.parser.kind() === 'LessThanToken') {
                    children.push(this.element(true));
                }
                else {
                    panic(`JSX expected attribute value at ${this.parser.scanner.start} in ${this.parser.path}`);
                }
            }
            attributes.push(this.make('JsxAttribute', start, children));
        }
        const id = this.make('JsxAttributes', pos, attributes);
        this.parser.node(id).list = attributes.length;
        return id;
    }
    tagKey(index: number): string {
        const node = this.parser.node(index);
        let key = `${node.kind}:${node.text}`;
        for(const child of node.children) {
            key += `:${this.tagKey(child)}`;
        }
        return key;
    }
    element(inExpression: boolean): number {
        const pos = this.parser.scanner.fullStart;
        this.parser.expect('LessThanToken');
        const fragment = this.parser.kind() === 'GreaterThanToken';
        const header: number[] = [];
        let typeCount = -1;
        let typeTrailing = false;
        if(!fragment) {
            header.push(this.name(true));
            if(!this.parser.javascript && this.parser.kind() === 'LessThanToken') {
                const types = this.parser.typeArguments();
                typeCount = types.length;
                typeTrailing = this.parser.typeTrailing();
                for(const type of types) {
                    header.push(type);
                }
            }
            header.push(this.attributes());
        }
        if(!fragment && this.parser.kind() === 'SlashToken') {
            this.parser.next();
            this.finishGreater(inExpression);
            const id = this.make('JsxSelfClosingElement', pos, header);
            this.parser.node(id).semantic = `${typeCount}:${typeTrailing ? 1 : 0}`;
            return id;
        }
        this.finishGreater(false);
        const opening = this.make(fragment ? 'JsxOpeningFragment' : 'JsxOpeningElement', pos, header);
        if(!fragment) {
            this.parser.node(opening).semantic = `${typeCount}:${typeTrailing ? 1 : 0}`;
        }
        const children = [opening];
        let count = 0;
        while(this.parser.kind() !== 'LessThanSlashToken') {
            switch(this.parser.kind()) {
                case 'JsxText':
                case 'JsxTextAllWhiteSpaces':
                    children.push(this.text());
                    break;
                case 'OpenBraceToken':
                    children.push(this.expression(false));
                    break;
                case 'LessThanToken':
                    children.push(this.element(false));
                    break;
                default:
                    panic(`JSX missing closing tag at ${this.parser.scanner.start} in ${this.parser.path}`);
            }
            count++;
        }
        const closePos = this.parser.scanner.fullStart;
        this.parser.next();
        const closeChildren: number[] = [];
        if(!fragment) {
            const name = this.name(true);
            if(this.tagKey(name) !== this.tagKey(header[0] ?? panic('missing opening name'))) {
                panic(`JSX mismatched closing tag at ${closePos} in ${this.parser.path}`);
            }
            closeChildren.push(name);
        }
        this.finishGreater(inExpression);
        children.push(this.make(fragment ? 'JsxClosingFragment' : 'JsxClosingElement', closePos, closeChildren));
        const id = this.make(fragment ? 'JsxFragment' : 'JsxElement', pos, children);
        this.parser.node(id).list = count;
        return id;
    }
}
