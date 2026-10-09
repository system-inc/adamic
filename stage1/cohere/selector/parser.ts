/* eslint-disable nexus/correctness-no-caller-data-mutation -- As in cohere, append and newNode take ownership of fresh nodes in the parser arena. */
// Cohere's parser.go, postcss-selector-parser 2.2.3 dist/parser.js.
// Lossless defaults only, as Prettier calls it. Methods are callee first;
// parentheses and parse share one recursion (GraphQL's forward-method gap).
import { panic } from 'adamic';
import { convertSourceIndexes, SelectorNode } from './nodes.ts';
import { Source } from './source.ts';
import { type Token, tokenize } from './tokenize.ts';

export const loopsForever =
    'postcss-selector-parser 2.2.3 never returns on this selector: a namespace bar it does not consume';

function added(left: number | undefined, right: number): number {
    return left === undefined ? NaN : left + right;
}

class Parser {
    position = 0;
    readonly nodes: SelectorNode[] = [];
    readonly tokens: readonly Token[];
    current = 1;
    spaces = '';
    constructor(css: string) {
        this.tokens = tokenize(css);
        const root = new SelectorNode('root', undefined, undefined, undefined);
        const selector = new SelectorNode('selector', undefined, undefined, undefined);
        selector.parent = 0;
        root.children.push(1);
        this.nodes.push(root);
        this.nodes.push(selector);
    }
    node(index: number): SelectorNode {
        return this.nodes[index] ?? panic('missing node');
    }
    token(index: number): Token | undefined {
        return this.tokens[index];
    }
    curr(): Token {
        return this.tokens[this.position] ?? panic('missing token');
    }
    // Caller supplies a fresh node; the parser takes ownership in its arena.
    append(container: number, node: SelectorNode): number {
        const index = this.nodes.length;
        node.parent = container;
        this.node(container).children.push(index);
        this.nodes.push(node);
        return index;
    }
    // The caller hands over a fresh node, as in the original parser.
    newNode(node: SelectorNode, namespace: string | undefined): void {
        if(namespace !== undefined) {
            if(namespace !== '') {
                node.namespace = namespace;
            }
            else {
                node.namespaceTrue = true;
            }
        }
        if(this.spaces !== '') {
            node.before = this.spaces;
            this.spaces = '';
        }
        this.append(this.current, node);
    }
    last(): SelectorNode | undefined {
        const children = this.node(this.current).children;
        const index = children[children.length - 1];
        return index === undefined ? undefined : this.node(index);
    }
    simple(type: string, long: boolean): void {
        const token = this.curr();
        this.newNode(
            new SelectorNode(
                type,
                token.value,
                new Source(token.at(2), token.at(3), token.at(long ? 4 : 2), token.at(long ? 5 : 3)),
                token.at(long ? 6 : 4),
            ),
            undefined,
        );
        this.position++;
    }
    attribute(): void {
        let str = '';
        const starting = this.curr();
        this.position++;
        while(this.position < this.tokens.length && this.curr().kind !== ']') {
            str += this.curr().value;
            this.position++;
        }
        if(this.position === this.tokens.length && !str.includes(']')) {
            throw new Error('Expected a closing square bracket.');
        }
        // The Go's sole regexp, attributeOperator, matched natively by Adamic.
        const parts = str.split(/([*~^$|]?=)([\s\S]*)/);
        const namespace = (parts[0] ?? '').split('|');
        const closing = this.token(this.position);
        if(closing === undefined) {
            throw new Error("Cannot read properties of undefined (reading '2')");
        }
        const attr = new SelectorNode(
            'attribute',
            parts[2],
            new Source(starting.at(2), starting.at(3), closing.at(2), closing.at(3)),
            starting.at(4),
        );
        attr.operator = parts[1];
        attr.attribute = namespace.length > 1 ? (namespace[1] ?? '') : (parts[0] ?? '');
        if(namespace.length > 1) {
            attr.namespaceTrue = namespace[0] === '';
            if(!attr.namespaceTrue) {
                attr.namespace = namespace[0];
            }
        }
        const value = parts[2];
        if(value !== undefined && value !== '') {
            // Go walks backwards to express this JS regexp's whitespace class.
            const insensitive = value.split(/(\s+i\s*?)$/);
            attr.value = insensitive[0] ?? '';
            const suffix = insensitive[1];
            if(suffix !== undefined && suffix !== '') {
                attr.insensitive = true;
                attr.rawInsensitive = suffix;
            }
            const trimmed = (insensitive[0] ?? '').trim();
            attr.quoted = trimmed.startsWith("'") || trimmed.startsWith('"');
            attr.unquoted = attr.quoted ? trimmed.slice(1, -1) : trimmed;
        }
        this.newNode(attr, undefined);
        this.position++;
    }
    splitWord(namespace: string | undefined, pseudo: string | undefined, starting: Token): void {
        let next = this.token(this.position + 1);
        let word = this.curr().value;
        while(next !== undefined && next.kind === 'word') {
            this.position++;
            const current = this.curr().value;
            word += current;
            if(current.lastIndexOf('\\') === current.length - 1) {
                const following = this.token(this.position + 1);
                if(following !== undefined && following.kind === 'space') {
                    word += following.value;
                    this.position++;
                }
            }
            next = this.token(this.position + 1);
        }
        // indexes-of includes escaped dots too, and excludes Sass #{ hashes.
        const indices: number[] = [0];
        for(let index = 1; index < word.length; index++) {
            if(word[index] === '.' || (word[index] === '#' && word[index + 1] !== '{')) {
                indices.push(index);
            }
        }
        for(let partIndex = 0; partIndex < indices.length; partIndex++) {
            const ind = indices[partIndex] ?? panic('missing word index');
            const end = indices[partIndex + 1] ?? word.length;
            const value = word.slice(ind, end);
            const token = this.curr();
            if(partIndex === 0 && pseudo !== undefined) {
                this.newNode(
                    new SelectorNode(
                        'pseudo',
                        pseudo + value,
                        new Source(starting.at(2), starting.at(3), token.at(4), token.at(5)),
                        starting.at(4),
                    ),
                    undefined,
                );
                if(indices.length > 1 && next !== undefined && next.kind === '(') {
                    throw new Error('Misplaced parenthesis.');
                }
                continue;
            }
            const isClass = word[ind] === '.';
            const isID = word[ind] === '#' && word[ind + 1] !== '{';
            this.newNode(
                new SelectorNode(
                    isClass ? 'class' : isID ? 'id' : 'tag',
                    isClass || isID ? value.slice(1) : value,
                    new Source(token.at(2), added(token.at(3), ind), token.at(4), added(token.at(3), end - 1)),
                    added(token.at(6), ind),
                ),
                namespace,
            );
        }
        this.position++;
    }
    word(namespace: string | undefined): void {
        const next = this.token(this.position + 1);
        if(next !== undefined && next.value === '|') {
            const before = this.curr().value;
            this.position += 2;
            const following = this.token(this.position);
            if(following === undefined) {
                throw new Error("Cannot read properties of undefined (reading '0')");
            }
            if(following.kind !== 'word' && following.kind !== '*') {
                throw new Error(loopsForever);
            }
            this.word(before);
            return;
        }
        if(this.curr().kind === '*') {
            const token = this.curr();
            this.newNode(
                new SelectorNode(
                    'universal',
                    '*',
                    new Source(token.at(2), token.at(3), token.at(2), token.at(3)),
                    token.at(4),
                ),
                namespace,
            );
            this.position++;
        }
        else {
            this.splitWord(namespace, undefined, this.curr());
        }
    }
    combinator(): void {
        if(this.curr().value === '|') {
            const prev = this.token(this.position - 1);
            const before = prev === undefined ? '' : prev.value;
            this.position++;
            const next = this.token(this.position);
            if(next === undefined) {
                throw new Error("Cannot read properties of undefined (reading '0')");
            }
            if(next.kind !== 'word' && next.kind !== '*') {
                throw new Error(loopsForever);
            }
            this.word(before);
            return;
        }
        const token = this.curr();
        const source = new Source(token.at(2), token.at(3), token.at(2), token.at(3));
        const node = new SelectorNode('combinator', '', source, token.at(4));
        while(
            this.position < this.tokens.length &&
            (this.curr().kind === 'space' || this.curr().kind === 'combinator')
        ) {
            const next = this.token(this.position + 1);
            const prev = this.token(this.position - 1);
            if(next !== undefined && next.kind === 'combinator') {
                node.before = this.curr().value;
                source.startLine = next.at(2);
                source.startColumn = next.at(3);
                source.endLine = next.at(2);
                source.endColumn = next.at(3);
                node.sourceIndex = next.at(4);
            }
            else if(prev !== undefined && prev.kind === 'combinator') {
                node.after = this.curr().value;
            }
            else {
                node.value = this.curr().value;
            }
            this.position++;
        }
        this.newNode(node, undefined);
    }
    parse(throwOnParenthesis: boolean): void {
        const token = this.curr();
        switch(token.kind) {
            case 'comment':
                this.simple('comment', true);
                break;
            case 'string':
                this.simple('string', true);
                break;
            case '&':
                this.simple('nesting', false);
                break;
            case '[':
                this.attribute();
                break;
            case ']':
                throw new Error('Expected opening square bracket.');
            case ';':
                throw new Error('Expected a backslash preceding the semicolon.');
            case ')':
                if(throwOnParenthesis) {
                    throw new Error('Expected opening parenthesis.');
                }
                break;
            case 'word':
            case 'at-word':
            case '*':
                this.word(undefined);
                break;
            case 'combinator':
                this.combinator();
                break;
            case ',':
                if(this.position === this.tokens.length - 1) {
                    this.node(0).trailingComma = true;
                }
                else {
                    this.current = this.append(
                        this.node(this.current).parent,
                        new SelectorNode('selector', undefined, undefined, undefined),
                    );
                }
                this.position++;
                break;
            case 'space': {
                const prev = this.token(this.position - 1);
                const next = this.token(this.position + 1);
                if(prev === undefined || prev.kind === ',' || prev.kind === '(') {
                    this.spaces = token.value;
                    this.position++;
                }
                else if(next === undefined || next.kind === ',' || next.kind === ')') {
                    const last = this.last();
                    if(last === undefined) {
                        throw new Error("Cannot read properties of undefined (reading 'spaces')");
                    }
                    last.after = token.value;
                    this.position++;
                }
                else {
                    this.combinator();
                }
                break;
            }
            case ':': {
                let prefix = '';
                while(this.position < this.tokens.length && this.curr().kind === ':') {
                    prefix += this.curr().value;
                    this.position++;
                }
                if(this.position === this.tokens.length) {
                    throw new Error('Expected pseudo-class or pseudo-element');
                }
                if(this.curr().kind !== 'word') {
                    throw new Error(`Unexpected "${this.curr().kind}" found.`);
                }
                this.splitWord(undefined, prefix, token);
                break;
            }
            case '(': {
                const last = this.last();
                let balanced = 1;
                this.position++;
                if(last !== undefined && last.type === 'pseudo') {
                    const cache = this.current;
                    const parent = this.node(cache).children.at(-1) ?? panic('missing pseudo');
                    this.current = this.append(parent, new SelectorNode('selector', undefined, undefined, undefined));
                    while(this.position < this.tokens.length && balanced !== 0) {
                        if(this.curr().kind === '(') {
                            balanced++;
                        }
                        if(this.curr().kind === ')') {
                            balanced--;
                        }
                        if(balanced !== 0) {
                            this.parse(false);
                        }
                        else {
                            const source = last.source ?? panic('missing pseudo source');
                            source.endLine = this.curr().at(2);
                            source.endColumn = this.curr().at(3);
                            this.position++;
                        }
                    }
                    this.current = cache;
                }
                else {
                    if(last === undefined) {
                        throw new Error("Cannot read properties of undefined (reading 'value')");
                    }
                    last.value = `${last.value ?? 'undefined'}(`;
                    while(this.position < this.tokens.length && balanced !== 0) {
                        if(this.curr().kind === '(') {
                            balanced++;
                        }
                        if(this.curr().kind === ')') {
                            balanced--;
                        }
                        last.value = (last.value ?? 'undefined') + this.curr().value;
                        this.position++;
                    }
                }
                if(balanced !== 0) {
                    throw new Error('Expected closing parenthesis.');
                }
                break;
            }
        }
    }
    loop(): void {
        while(this.position < this.tokens.length) {
            this.parse(true);
        }
    }
}

export type ParseResultType =
    | { readonly kind: 'Parsed'; readonly nodes: readonly SelectorNode[] }
    | { readonly kind: 'Refused'; readonly message: string };
export function parse(css: string): ParseResultType {
    try {
        const parser = new Parser(css);
        parser.loop();
        convertSourceIndexes(parser.nodes, css);
        return { kind: 'Parsed', nodes: parser.nodes };
    }
    catch(error) {
        return { kind: 'Refused', message: error instanceof Error ? error.message : 'not an Error' };
    }
}
