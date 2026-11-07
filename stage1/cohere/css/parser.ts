// Cohere's postcss 8.5.16 parser. Token arrays are copied where upstream slices.
import { panic } from 'adamic';
import { Input } from './input.ts';
import { CssNode, RawValue } from './nodes.ts';
import { Token, Tokenizer } from './tokenize.ts';

function tokenString(tokens: readonly Token[], from: number = 0, to: number = tokens.length): string {
    const parts: string[] = [];
    for(let index = from; index < to; index++) {
        parts.push((tokens[index] ?? panic('missing token')).value);
    }
    return parts.join('');
}
function lastPosition(tokens: readonly Token[]): number {
    for(let index = tokens.length - 1; index >= 0; index--) {
        const pos = (tokens[index] ?? panic('missing token')).endOrStart();
        if(pos > 0) {
            return pos;
        }
    }
    return -1;
}
function spacesEnd(tokens: Token[], comments: boolean = true): string {
    let result = '';
    while(tokens.length > 0) {
        const last = tokens[tokens.length - 1] ?? panic('missing token');
        if(last.kind !== 'space' && !(comments && last.kind === 'comment')) {
            break;
        }
        result = last.value + result;
        tokens.pop();
    }
    return result;
}
function spacesStart(tokens: Token[]): string {
    let result = '';
    while(tokens.length > 0) {
        const first = tokens[0] ?? panic('missing token');
        if(first.kind !== 'space' && first.kind !== 'comment') {
            break;
        }
        result += first.value;
        tokens.splice(0, 1);
    }
    return result;
}
export class Parser {
    readonly input: Input;
    readonly tokenizer: Tokenizer;
    readonly scss: boolean;
    readonly nodes: CssNode[] = [];
    current = 0;
    spaces = '';
    semicolon = false;
    constructor(input: Input, scss: boolean = false) {
        this.scss = scss;
        this.input = input;
        this.tokenizer = new Tokenizer(input, scss);
        const root = new CssNode('root');
        root.start = input.position(0);
        root.mark('source');
        this.nodes.push(root);
    }
    node(index: number): CssNode {
        return this.nodes[index] ?? panic('missing CSS node');
    }
    init(type: string, offset: number): number {
        const node = new CssNode(type);
        const index = this.nodes.length;
        node.parent = this.current;
        node.start = this.input.position(offset);
        node.mark('source');
        node.rawStrings.set('before', this.spaces);
        this.spaces = '';
        this.node(this.current).children.push(index);
        this.nodes.push(node);
        if(type !== 'comment') {
            this.semicolon = false;
        }
        return index;
    }
    finish(node: CssNode, offset: number, increment: number = 1): void {
        node.end = this.input.position(offset);
        node.end.offset += increment;
    }
    unknown(tokens: readonly Token[]): void {
        const first = tokens[0] ?? panic('missing unknown token');
        throw new Error(this.input.error(`Unknown word ${first.value}`, first.start, first.start + first.value.length));
    }
    raw(node: CssNode, prop: string, tokens: readonly Token[], customProperty: boolean = false): void {
        let value = '';
        let clean = true;
        for(let index = 0; index < tokens.length; index++) {
            const current = tokens[index] ?? panic('missing raw token');
            if(current.kind === 'space' && index === tokens.length - 1 && !customProperty) {
                clean = false;
            }
            else if(current.kind === 'comment') {
                const previous = index > 0 ? (tokens[index - 1] ?? panic('missing previous')).kind : 'empty';
                const next = index + 1 < tokens.length ? (tokens[index + 1] ?? panic('missing next')).kind : 'empty';
                if(previous !== 'empty' && previous !== 'space' && next !== 'empty' && next !== 'space') {
                    if(value.endsWith(',')) {
                        clean = false;
                    }
                    else {
                        value += current.value;
                    }
                }
                else {
                    clean = false;
                }
            }
            else {
                value += current.value;
            }
        }
        if(!clean) {
            const original = tokenString(tokens);
            const raw = new RawValue(original, value);
            if(this.scss) {
                let rewritten = '';
                for(const token of tokens) {
                    rewritten +=
                        token.kind === 'comment' && token.inline
                            ? `/*${token.value.slice(2).replace(/(\*\/|\/\*)/g, '*//*')}*/`
                            : token.value;
                }
                raw.raw = rewritten;
                if(original !== rewritten) {
                    raw.scss = original;
                }
            }
            node.rawValues.set(prop, raw);
        }
        node.set(prop, value);
    }
    colon(tokens: readonly Token[]): number {
        let brackets = 0;
        let previous: Token | undefined = undefined;
        for(let index = 0; index < tokens.length; index++) {
            const token = tokens[index] ?? panic('missing colon token');
            if(token.kind === '(') {
                brackets++;
            }
            if(token.kind === ')') {
                brackets--;
            }
            if(brackets === 0 && token.kind === ':') {
                if(previous === undefined) {
                    throw new Error(this.input.error('Double colon', token.start, token.start + token.value.length));
                }
                else if(previous.kind === 'word' && previous.value === 'progid') {
                    continue;
                }
                else {
                    return index;
                }
            }
            previous = token;
        }
        return -1;
    }
    missedSemicolon(tokens: readonly Token[]): void {
        const colon = this.colon(tokens);
        if(colon < 0) {
            return;
        }
        let found = 0;
        let token = new Token('', '', 0, 0);
        for(let index = colon - 1; index >= 0; index--) {
            token = tokens[index] ?? panic('missing preceding token');
            if(token.kind !== 'space') {
                found++;
                if(found === 2) {
                    break;
                }
            }
        }
        throw new Error(this.input.error('Missed semicolon', token.kind === 'word' ? token.end + 1 : token.start));
    }
    decl(tokens: Token[], customProperty: boolean): void {
        const node = this.node(this.init('decl', (tokens[0] ?? panic('missing declaration')).start));
        const last = tokens[tokens.length - 1] ?? panic('missing declaration end');
        if(last.kind === ';') {
            this.semicolon = true;
            tokens.pop();
        }
        let offset = last.endOrStart();
        if(offset <= 0) {
            offset = lastPosition(tokens);
        }
        this.finish(node, offset);
        let start = 0;
        while((tokens[start] ?? panic('missing property token')).kind !== 'word') {
            if(start === tokens.length - 1) {
                this.unknown([tokens[start] ?? panic('missing token')]);
            }
            start++;
        }
        node.rawStrings.set('before', node.raw('before') + tokenString(tokens, 0, start));
        node.start = this.input.position((tokens[start] ?? panic('missing property start')).start);
        const propStart = start;
        while(start < tokens.length) {
            const kind = (tokens[start] ?? panic('missing property part')).kind;
            if(kind === ':' || kind === 'space' || kind === 'comment') {
                break;
            }
            start++;
        }
        node.set('prop', tokenString(tokens, propStart, start));
        const betweenStart = start;
        while(start < tokens.length) {
            const token = tokens[start] ?? panic('missing colon part');
            start++;
            if(token.kind === ':') {
                break;
            }
            if(token.kind === 'word' && /\w/.test(token.value)) {
                this.unknown([token]);
            }
        }
        node.rawStrings.set('between', tokenString(tokens, betweenStart, start));
        const prop = node.get('prop');
        if(prop[0] === '_' || prop[0] === '*') {
            node.rawStrings.set('before', node.raw('before') + prop.slice(0, 1));
            node.set('prop', prop.slice(1));
        }
        const firstSpacesStart = start;
        while(start < tokens.length) {
            const kind = (tokens[start] ?? panic('missing value part')).kind;
            if(kind !== 'space' && kind !== 'comment') {
                break;
            }
            start++;
        }
        let firstSpaces = tokens.slice(firstSpacesStart, start);
        tokens = tokens.slice(start);
        for(let index = tokens.length - 1; index >= 0; index--) {
            const token = tokens[index] ?? panic('missing important token');
            const lowered = token.value.toLowerCase();
            if(lowered === '!important') {
                node.booleans.set('important', true);
                node.mark('important');
                let text = tokenString(tokens, index);
                tokens = tokens.slice(0, index);
                text = spacesEnd(tokens, false) + text;
                if(text !== ' !important') {
                    node.rawStrings.set('important', text);
                }
                break;
            }
            else if(lowered === 'important') {
                const cache = tokens.slice();
                let text = '';
                for(let j = index; j > 0; j--) {
                    const kind = (cache[j] ?? panic('missing important part')).kind;
                    if(text.trim().startsWith('!') && kind !== 'space') {
                        break;
                    }
                    text = (cache.pop() ?? panic('missing cached token')).value + text;
                }
                if(text.trim().startsWith('!')) {
                    node.booleans.set('important', true);
                    node.mark('important');
                    node.rawStrings.set('important', text);
                    tokens = cache;
                }
            }
            if(token.kind !== 'space' && token.kind !== 'comment') {
                break;
            }
        }
        let hasWord = false;
        for(const token of tokens) {
            if(token.kind !== 'space' && token.kind !== 'comment') {
                hasWord = true;
                break;
            }
        }
        if(hasWord) {
            node.rawStrings.set('between', node.raw('between') + tokenString(firstSpaces));
            firstSpaces = [];
        }
        this.raw(node, 'value', firstSpaces.concat(tokens), customProperty);
        if(node.get('value').includes(':') && !customProperty) {
            this.missedSemicolon(tokens);
        }
    }
    comment(token: Token): void {
        const node = this.node(this.init('comment', token.start));
        this.finish(node, token.endOrStart());
        const text = token.inline ? token.value.slice(2) : token.value.slice(2, -2);
        if(token.inline) {
            node.rawBooleans.set('inline', true);
        }
        if(text.trim() === '') {
            node.set('text', '');
            node.rawStrings.set('left', text);
            node.rawStrings.set('right', '');
        }
        else {
            const left = text.trimStart();
            const trimmed = left.trimEnd();
            node.set('text', token.inline ? trimmed.replace(/(\*\/|\/\*)/g, '*//*') : trimmed);
            if(token.inline) {
                node.rawStrings.set('text', trimmed);
            }
            node.rawStrings.set('left', text.slice(0, text.length - left.length));
            node.rawStrings.set('right', left.slice(trimmed.length));
        }
    }
    end(token: Token): void {
        const node = this.node(this.current);
        if(node.keys.includes('nodes') && node.children.length > 0) {
            node.rawBooleans.set('semicolon', this.semicolon);
        }
        this.semicolon = false;
        node.rawStrings.set('after', node.raw('after') + this.spaces);
        this.spaces = '';
        if(node.parent >= 0) {
            this.finish(node, token.start);
            this.current = node.parent;
        }
        else {
            throw new Error(this.input.error('Unexpected }', token.start, token.start + 1));
        }
    }
    nestedRule(tokens: Token[]): void {
        tokens.pop();
        const index = this.init('decl', (tokens[0] ?? panic('missing nested property')).start);
        const node = this.node(index);
        // NestedDeclaration creates isNested and nodes before Parser.init sets source.
        node.keys.pop();
        node.booleans.set('isNested', true);
        node.mark('isNested');
        node.mark('nodes');
        node.mark('source');
        for(let i = tokens.length - 1; i >= 0; i--) {
            const last = tokens[i] ?? panic('missing nested end');
            if(last.kind !== 'space') {
                this.finish(node, last.endOrStart());
                break;
            }
        }
        while(tokens.length > 0 && (tokens[0] ?? panic('missing nested token')).kind !== 'word') {
            node.rawStrings.set('before', node.raw('before') + (tokens[0] ?? panic('missing nested prefix')).value);
            tokens.splice(0, 1);
        }
        if(tokens.length === 0) {
            throw new Error(`{"notCssSyntaxError":"TypeError: Cannot read properties of undefined (reading '0')"}`);
        }
        node.start = this.input.position((tokens[0] ?? panic('missing nested start')).start);
        let prop = '';
        while(tokens.length > 0) {
            const token = tokens[0] ?? panic('missing nested property part');
            if(token.kind === ':' || token.kind === 'space' || token.kind === 'comment') {
                break;
            }
            prop += token.value;
            tokens.splice(0, 1);
        }
        node.set('prop', prop);
        let between = '';
        while(tokens.length > 0) {
            const token = tokens[0] ?? panic('missing nested colon');
            tokens.splice(0, 1);
            between += token.value;
            if(token.kind === ':') {
                break;
            }
        }
        if(prop[0] === '_' || prop[0] === '*') {
            node.rawStrings.set('before', node.raw('before') + prop.slice(0, 1));
            node.set('prop', prop.slice(1));
        }
        node.rawStrings.set('between', between + spacesStart(tokens));
        for(let i = tokens.length - 1; i > 0; i--) {
            const token = tokens[i] ?? panic('missing nested important');
            if(token.value === '!important') {
                node.booleans.set('important', true);
                node.mark('important');
                let text = tokenString(tokens, i);
                tokens = tokens.slice(0, i);
                text = spacesEnd(tokens, false) + text;
                if(text !== ' !important') {
                    node.rawStrings.set('important', text);
                }
                break;
            }
            else if(token.value === 'important') {
                const cache = tokens.slice();
                let text = '';
                for(let j = i; j > 0; j--) {
                    const kind = (cache[j] ?? panic('missing nested important part')).kind;
                    if(text.trim().startsWith('!') && kind !== 'space') {
                        break;
                    }
                    text = (cache.pop() ?? panic('missing nested cached token')).value + text;
                }
                if(text.trim().startsWith('!')) {
                    node.booleans.set('important', true);
                    node.mark('important');
                    node.rawStrings.set('important', text);
                    tokens = cache;
                }
            }
            if(token.kind !== 'space' && token.kind !== 'comment') {
                break;
            }
        }
        this.raw(node, 'value', tokens);
        if(node.get('value').includes(':')) {
            this.missedSemicolon(tokens);
        }
        this.current = index;
    }
    rule(tokens: Token[]): void {
        if(this.scss) {
            let colon = false;
            let brackets = 0;
            let value = '';
            for(const token of tokens) {
                if(colon) {
                    if(token.kind !== 'comment' && token.kind !== '{') {
                        value += token.value;
                    }
                }
                else if(token.kind === 'space' && token.value.includes('\n')) {
                    break;
                }
                else if(token.kind === '(') {
                    brackets++;
                }
                else if(token.kind === ')') {
                    brackets--;
                }
                else if(brackets === 0 && token.kind === ':') {
                    colon = true;
                }
            }
            if(colon && value.trim() !== '' && !/^[#:A-Za-z-]/.test(value)) {
                this.nestedRule(tokens);
                return;
            }
        }
        tokens.pop();
        const node = this.node(this.init('rule', (tokens[0] ?? panic('missing selector')).start));
        node.rawStrings.set('between', spacesEnd(tokens));
        this.raw(node, 'selector', tokens);
        this.current = this.nodes.length - 1;
    }
    emptyRule(token: Token): void {
        this.current = this.init('rule', token.start);
        const node = this.node(this.current);
        node.set('selector', '');
        node.rawStrings.set('between', '');
    }
    freeSemicolon(token: Token): void {
        this.spaces += token.value;
        const current = this.node(this.current);
        if(current.keys.includes('nodes') && current.children.length > 0) {
            const prev = this.node(current.children[current.children.length - 1] ?? panic('missing previous node'));
            if(prev.type === 'rule' && prev.raw('ownSemicolon') === '') {
                prev.rawStrings.set('ownSemicolon', this.spaces);
                this.spaces = '';
                this.finish(prev, token.start, prev.raw('ownSemicolon').length);
            }
        }
    }
    atrule(at: Token): void {
        if(this.scss) {
            let name = at.value;
            let previous = at;
            while(!this.tokenizer.endOfFile()) {
                const next = this.tokenizer.nextToken() ?? panic('missing joined at-rule token');
                if(next.kind === 'word' && next.start === previous.end + 1) {
                    name += next.value;
                    previous = next;
                }
                else {
                    this.tokenizer.back(next);
                    break;
                }
            }
            at = new Token('at-word', name, at.start, previous.end);
        }
        // AtRule's name precedes source in its own property order.
        const index = this.init('atrule', at.start);
        const node = this.node(index);
        node.keys.pop();
        node.set('name', at.value.slice(1));
        node.mark('source');
        if(node.get('name') === '') {
            throw new Error(this.input.error('At-rule without name', at.start, at.start + at.value.length));
        }
        const params: Token[] = [];
        const brackets: string[] = [];
        let last = false;
        let open = false;
        while(!this.tokenizer.endOfFile()) {
            const token = this.tokenizer.nextToken() ?? panic('missing at-rule token');
            const kind = token.kind;
            if(kind === '(' || kind === '[') {
                brackets.push(kind === '(' ? ')' : ']');
            }
            else if(kind === '{' && brackets.length > 0) {
                brackets.push('}');
            }
            else if(brackets.length > 0 && kind === brackets[brackets.length - 1]) {
                brackets.pop();
            }
            if(brackets.length === 0) {
                if(kind === ';') {
                    this.finish(node, token.start);
                    this.semicolon = true;
                    break;
                }
                else if(kind === '{') {
                    open = true;
                    break;
                }
                else if(kind === '}') {
                    let shift = params.length - 1;
                    while(shift >= 0 && (params[shift] ?? panic('missing param')).kind === 'space') {
                        shift--;
                    }
                    if(shift >= 0) {
                        this.finish(node, (params[shift] ?? panic('missing final param')).endOrStart());
                    }
                    this.end(token);
                    break;
                }
                else {
                    params.push(token);
                }
            }
            else {
                params.push(token);
            }
            if(this.tokenizer.endOfFile()) {
                last = true;
                break;
            }
        }
        node.rawStrings.set('between', spacesEnd(params));
        if(params.length > 0) {
            node.rawStrings.set('afterName', spacesStart(params));
            this.raw(node, 'params', params);
            if(last) {
                this.finish(node, (params[params.length - 1] ?? panic('missing params end')).endOrStart());
                this.spaces = node.raw('between');
                node.rawStrings.set('between', '');
            }
        }
        else {
            node.rawStrings.set('afterName', '');
            node.set('params', '');
        }
        if(open) {
            node.mark('nodes');
            this.current = index;
        }
    }
    other(start: Token): void {
        let end = false;
        let colon = false;
        let bracket: Token | undefined = undefined;
        const brackets: string[] = [];
        const custom = start.value.startsWith('--');
        const tokens: Token[] = [];
        let token: Token | undefined = start;
        while(token !== undefined) {
            const kind = token.kind;
            tokens.push(token);
            if(kind === '(' || kind === '[') {
                if(bracket === undefined) {
                    bracket = token;
                }
                brackets.push(kind === '(' ? ')' : ']');
            }
            else if(custom && colon && kind === '{') {
                if(bracket === undefined) {
                    bracket = token;
                }
                brackets.push('}');
            }
            else if(brackets.length === 0) {
                if(kind === ';') {
                    if(colon) {
                        this.decl(tokens, custom);
                        return;
                    }
                    break;
                }
                else if(kind === '{') {
                    this.rule(tokens);
                    return;
                }
                else if(kind === '}') {
                    this.tokenizer.back(tokens.pop() ?? panic('missing close token'));
                    end = true;
                    break;
                }
                else if(kind === ':') {
                    colon = true;
                }
            }
            else if(kind === brackets[brackets.length - 1]) {
                brackets.pop();
                if(brackets.length === 0) {
                    bracket = undefined;
                }
            }
            token = this.tokenizer.nextToken();
        }
        if(this.tokenizer.endOfFile()) {
            end = true;
        }
        if(brackets.length > 0) {
            const opened = bracket ?? panic('missing open bracket');
            throw new Error(this.input.error('Unclosed bracket', opened.start, opened.start + 1));
        }
        if(end && colon) {
            if(!custom) {
                while(tokens.length > 0) {
                    const kind = (tokens[tokens.length - 1] ?? panic('missing end token')).kind;
                    if(kind !== 'space' && kind !== 'comment') {
                        break;
                    }
                    this.tokenizer.back(tokens.pop() ?? panic('missing whitespace token'));
                }
            }
            this.decl(tokens, custom);
        }
        else {
            this.unknown(tokens);
        }
    }
    parse(): void {
        while(!this.tokenizer.endOfFile()) {
            const token = this.tokenizer.nextToken() ?? panic('missing parse token');
            switch(token.kind) {
                case 'space':
                    this.spaces += token.value;
                    break;
                case ';':
                    this.freeSemicolon(token);
                    break;
                case '}':
                    this.end(token);
                    break;
                case 'comment':
                    this.comment(token);
                    break;
                case 'at-word':
                    this.atrule(token);
                    break;
                case '{':
                    this.emptyRule(token);
                    break;
                default:
                    this.other(token);
            }
        }
        const current = this.node(this.current);
        if(current.parent >= 0) {
            throw new Error(this.input.error('Unclosed block', (current.start ?? panic('missing block start')).offset));
        }
        if(current.keys.includes('nodes') && current.children.length > 0) {
            current.rawBooleans.set('semicolon', this.semicolon);
        }
        current.rawStrings.set('after', current.raw('after') + this.spaces);
        this.finish(this.node(0), this.tokenizer.pos, 0);
    }
}
export class Parsed {
    readonly kind = 'Parsed';
    readonly parser: Parser;
    constructor(parser: Parser) {
        this.parser = parser;
        // Go's toEstree crosses from tokenizer units to public byte positions.
        for(const node of parser.nodes) {
            if(node.start !== undefined) {
                node.start.offset = parser.input.byteOffset(node.start.offset);
            }
            if(node.end !== undefined) {
                node.end.offset = parser.input.byteOffset(node.end.offset);
            }
            if(node.start !== undefined && node.end !== undefined) {
                node.range[0] = node.start.offset;
                node.range[1] = node.end.offset;
            }
        }
    }
}
export class Refused {
    readonly kind = 'Refused';
    readonly message: string;
    constructor(message: string) {
        this.message = message;
    }
}
export function parse(text: string, scss: boolean = false): Parsed | Refused {
    try {
        const parser = new Parser(new Input(text), scss);
        parser.parse();
        return new Parsed(parser);
    }
    catch(error) {
        if(error instanceof Error) {
            return new Refused(error.message);
        }
        throw error;
    }
}
