// cohere/internal/format/estree/parse_json.go. Raw literals are retained for both printers.
import { panic } from 'adamic';
import { letters, continuations } from './identifierTables.ts';
export interface JsonNodeInterface {
    kind: string;
    start: number;
    end: number;
    raw: string;
    children: number[];
}
export interface CommentInterface {
    start: number;
    end: number;
    text: string;
    line: boolean;
}
export function quoted(text: string): string {
    const parts: string[] = ['"'];
    for(let index = 0; index < text.length; index++) {
        const char = text.slice(index, index + 1);
        const code = text.charCodeAt(index);
        if(char === '"' || char === '\\') {
            parts.push('\\');
            parts.push(char);
        }
        else if(char === '\n') {
            parts.push('\\n');
        }
        else if(char === '\r') {
            parts.push('\\r');
        }
        else if(char === '\t') {
            parts.push('\\t');
        }
        else if(code < 32) {
            parts.push('\\x');
            parts.push(code.toString(16).padStart(2, '0'));
        }
        else {
            parts.push(char);
        }
    }
    parts.push('"');
    return parts.join('');
}
function whitespace(code: number): boolean {
    return (
        (code >= 9 && code <= 13) ||
        code === 32 ||
        code === 0xa0 ||
        code === 0x1680 ||
        (code >= 0x2000 && code <= 0x200a) ||
        code === 0x2028 ||
        code === 0x2029 ||
        code === 0x202f ||
        code === 0x205f ||
        code === 0x3000 ||
        code === 0xfeff
    );
}
function digit(code: number): number {
    return code >= 48 && code <= 57
        ? code - 48
        : code >= 65 && code <= 70
          ? code - 55
          : code >= 97 && code <= 102
            ? code - 87
            : 100;
}
function inRanges(code: number, ranges: readonly (readonly number[])[]): boolean {
    for(const range of ranges) {
        const start = range[0] ?? panic('missing character range start');
        const end = range[1] ?? panic('missing character range end');
        const stride = range[2] ?? panic('missing character range stride');
        if(code >= start && code <= end && (code - start) % stride === 0) {
            return true;
        }
    }
    return false;
}
function nameStart(code: number): boolean {
    return (
        code === 36 ||
        code === 95 ||
        (code >= 65 && code <= 90) ||
        (code >= 97 && code <= 122) ||
        (code >= 128 && inRanges(code, letters))
    );
}
export function numericValue(raw: string): number {
    const text = raw.includes('_') ? raw.split('_').join('') : raw;
    const prefix = text.slice(0, 2).toLowerCase();
    if(prefix === '0x' || prefix === '0o' || prefix === '0b') {
        return Number.parseInt(text.slice(2), prefix === '0x' ? 16 : prefix === '0o' ? 8 : 2);
    }
    return Number.parseFloat(text);
}
// Go cohere's separatorsBetweenDigits: inspect neighbours in the whole text.
function separatorsBetweenDigits(text: string, start: number, end: number, radix: number): boolean {
    const forbidden = radix === 16 ? '.X_x' : '.BEO_beo';
    for(let index = start; index < end; index++) {
        if(text.charCodeAt(index) !== 95) {
            continue;
        }
        if(index + 1 >= text.length || digit(text.charCodeAt(index + 1)) >= radix ||
            forbidden.includes(text.slice(index + 1, index + 2))) {
            return false;
        }
        if(index > 0 && forbidden.includes(text.slice(index - 1, index))) {
            return false;
        }
    }
    return true;
}
export class Reader {
    position = 0;
    readonly nodes: JsonNodeInterface[] = [];
    readonly comments: CommentInterface[] = [];
    readonly text: string;
    constructor(text: string) {
        this.text = text;
    }
    get(index: number): JsonNodeInterface {
        return this.nodes[index] ?? panic(`missing node ${index}`);
    }
    add(kind: string, start: number, raw: string, children: number[]): number {
        const index = this.nodes.length;
        this.nodes.push({ kind, start, end: this.position, raw, children });
        return index;
    }
    message(detail: string): string {
        return `json: line ${this.text.slice(0, this.position).split('\n').length}: ${detail}`;
    }
    skip(): void {
        while(this.position < this.text.length) {
            const start = this.position;
            const code = this.text.charCodeAt(start);
            if(code === 47 && this.text.charCodeAt(start + 1) === 47) {
                this.position += 2;
                while(
                    this.position < this.text.length &&
                    !'\n\r\u2028\u2029'.includes(this.text.slice(this.position, this.position + 1))
                ) {
                    this.position++;
                }
                this.comments.push({
                    start,
                    end: this.position,
                    text: this.text.slice(start, this.position),
                    line: true,
                });
            }
            else if(code === 47 && this.text.charCodeAt(start + 1) === 42) {
                const end = this.text.indexOf('*/', start + 2);
                if(end < 0) {
                    throw new Error(this.message('unterminated comment'));
                }
                this.position = end + 2;
                this.comments.push({
                    start,
                    end: this.position,
                    text: this.text.slice(start, this.position),
                    line: false,
                });
            }
            else if(whitespace(code)) {
                this.position++;
            }
            else {
                return;
            }
        }
    }
    peek(): number {
        this.skip();
        return this.text.charCodeAt(this.position);
    }
    expect(char: string): void {
        if(this.peek() !== char.charCodeAt(0)) {
            throw new Error(this.message(`expected '${char}'`));
        }
        this.position++;
    }
    name(): string {
        const start = this.position;
        while(this.position < this.text.length) {
            const code = this.text.codePointAt(this.position) ?? panic('missing identifier point');
            const isStart = nameStart(code);
            const continuation =
                !isStart &&
                this.position > start &&
                (digit(code) < 10 ||
                    code === 0x200c ||
                    code === 0x200d ||
                    (code >= 128 && inRanges(code, continuations)));
            if(!isStart && !continuation) {
                break;
            }
            this.position += code > 0xffff ? 2 : 1;
        }
        return this.text.slice(start, this.position);
    }
    unexpected(): string {
        return this.message(`unexpected ${quoted(this.text.slice(this.position, this.position + 10))}`);
    }
    number(): number {
        const start = this.position;
        let position = start;
        let base = 10;
        const prefix = this.text.slice(position, position + 2).toLowerCase();
        if(prefix === '0x' || prefix === '0o' || prefix === '0b') {
            base = prefix === '0x' ? 16 : prefix === '0o' ? 8 : 2;
            position += 2;
            while(digit(this.text.charCodeAt(position)) < base || this.text.charCodeAt(position) === 95) {
                position++;
            }
        }
        else {
            while(digit(this.text.charCodeAt(position)) < 10 || this.text.charCodeAt(position) === 95) {
                position++;
            }
            if(this.text.slice(position, position + 1) === '.') {
                position++;
                while(digit(this.text.charCodeAt(position)) < 10 || this.text.charCodeAt(position) === 95) {
                    position++;
                }
            }
            if('eE'.includes(this.text.slice(position, position + 1)) && position < this.text.length) {
                position++;
                if('+-'.includes(this.text.slice(position, position + 1)) && position < this.text.length) {
                    position++;
                }
                while(digit(this.text.charCodeAt(position)) < 10 || this.text.charCodeAt(position) === 95) {
                    position++;
                }
            }
        }
        if(!separatorsBetweenDigits(this.text, base === 10 ? start : start + 2, position, base)) {
            throw new Error(this.message('A numeric separator is only allowed between two digits.'));
        }
        const raw = this.text.slice(start, position);
        const literal = raw.includes('_') ? raw.split('_').join('') : raw;
        if(base === 10 && literal.length > 1 && literal.startsWith('0') && digit(literal.charCodeAt(1)) < 10) {
            throw new Error(this.message(`legacy octal literal ${quoted(literal)}`));
        }
        const exponent = literal.toLowerCase().indexOf('e');
        const badExponent =
            base === 10 &&
            exponent >= 0 &&
            (literal.endsWith('e') || literal.endsWith('E') || literal.endsWith('+') || literal.endsWith('-'));
        if(Number.isNaN(numericValue(literal)) || badExponent) {
            throw new Error(this.message(`invalid number ${quoted(literal)}`));
        }
        if(nameStart(this.text.charCodeAt(position)) || digit(this.text.charCodeAt(position)) < 10) {
            throw new Error(this.message(`invalid number ${quoted(this.text.slice(start, position + 1))}`));
        }
        this.position = position;
        return this.add('NumericLiteral', start, raw, []);
    }
    string(): number {
        const start = this.position;
        const quote = this.text.slice(start, start + 1);
        let position = start + 1;
        while(position < this.text.length && this.text.charCodeAt(position) !== quote.charCodeAt(0)) {
            const code = this.text.charCodeAt(position);
            if(code === 92) {
                position++;
            }
            else if(quote !== '`' && (code === 10 || code === 13)) {
                throw new Error(this.message('unterminated string'));
            }
            else if(quote === '`' && this.text.slice(position, position + 2) === '${') {
                throw new Error(this.message("'TemplateLiteral' with expression is not allowed in JSON"));
            }
            position++;
        }
        if(position >= this.text.length) {
            throw new Error(this.message(quote === '`' ? 'unterminated template' : 'unterminated string'));
        }
        this.position = position + 1;
        const raw = this.text.slice(start, this.position);
        // Validate the escapes even when printing preserves the raw spelling.
        for(let index = 1; index < raw.length - 1; index++) {
            if(raw.charCodeAt(index) !== 92) {
                continue;
            }
            index++;
            const char = raw.slice(index, index + 1);
            if((char >= '1' && char <= '9') || (char === '0' && digit(raw.charCodeAt(index + 1)) < 10)) {
                throw new Error(this.message('an octal escape in a string'));
            }
            if(char === 'x' || char === 'u') {
                let length = char === 'x' ? 2 : 4;
                let digitsStart = index + 1;
                if(char === 'u' && raw.slice(index + 1, index + 2) === '{') {
                    const closing = raw.indexOf('}', index + 2);
                    if(closing < 0) {
                        throw new Error(this.message('an incomplete \\u{} escape'));
                    }
                    digitsStart++;
                    length = closing - digitsStart;
                    if(length === 0 || Number.parseInt(raw.slice(digitsStart, closing), 16) > 0x10ffff) {
                        throw new Error(this.message('an invalid \\u{} escape'));
                    }
                }
                if(digitsStart + length > raw.length - 1) {
                    throw new Error(this.message(`an incomplete \\${char} escape`));
                }
                for(let offset = 0; offset < length; offset++) {
                    if(digit(raw.charCodeAt(digitsStart + offset)) >= 16) {
                        throw new Error(this.message(`an invalid \\${char} escape`));
                    }
                }
                index = digitsStart + length - 1;
                if(raw.slice(digitsStart - 1, digitsStart) === '{') {
                    index++;
                }
            }
        }
        return this.add(quote === '`' ? 'TemplateLiteral' : 'StringLiteral', start, raw, []);
    }
    // Mutual recursion is one method, as in the GraphQL port, around stage 0's forward-method gap.
    value(key: boolean): number {
        const code = this.peek();
        const start = this.position;
        if(code === 123 || code === 91) {
            const object = code === 123;
            this.position++;
            const children: number[] = [];
            const closer = object ? '}' : ']';
            const closingCode = object ? 125 : 93;
            while(this.peek() !== closingCode) {
                if(!object && this.peek() === 44) {
                    this.position++;
                    children.push(this.add('Hole', this.position - 1, '', []));
                    continue;
                }
                if(object) {
                    const propertyStart = this.position;
                    if(this.peek() === 91) {
                        throw new Error(this.message('Computed key is not allowed in JSON'));
                    }
                    const propertyKey = this.value(true);
                    if(this.peek() !== 58) {
                        throw new Error(this.message('Shorthand property is not allowed in JSON'));
                    }
                    this.position++;
                    const propertyValue = this.value(false);
                    children.push(this.add('ObjectProperty', propertyStart, '', [propertyKey, propertyValue]));
                }
                else {
                    children.push(this.value(false));
                }
                if(this.peek() === 44) {
                    this.position++;
                    continue;
                }
                this.expect(closer);
                return this.add(object ? 'ObjectExpression' : 'ArrayExpression', start, '', children);
            }
            this.position++;
            return this.add(object ? 'ObjectExpression' : 'ArrayExpression', start, '', children);
        }
        if(code === 34 || code === 39 || code === 96) {
            return this.string();
        }
        if(!key && (code === 43 || code === 45)) {
            const char = code === 43 ? '+' : '-';
            this.position++;
            const argument = this.value(false);
            const node = this.get(argument);
            if(
                node.kind !== 'NumericLiteral' &&
                !(node.kind === 'Identifier' && (node.raw === 'Infinity' || node.raw === 'NaN'))
            ) {
                throw new Error(this.message(`Operator '${char}' before '${node.kind}' is not allowed in JSON`));
            }
            return this.add('UnaryExpression', start, char, [argument]);
        }
        if(code === 46 || (code >= 48 && code <= 57)) {
            return this.number();
        }
        const name = this.name();
        if(name === '') {
            throw new Error(this.unexpected());
        }
        if(key) {
            return this.add('Identifier', start, name, []);
        }
        if(name === 'null' || name === 'true' || name === 'false') {
            return this.add(name === 'null' ? 'NullLiteral' : 'BooleanLiteral', start, name, []);
        }
        if(name === 'NaN' || name === 'Infinity' || name === 'undefined') {
            return this.add('Identifier', start, name, []);
        }
        throw new Error(this.message(`Identifier '${name}' is not allowed in JSON`));
    }
    parse(stringify: boolean): number {
        this.skip();
        if(this.position >= this.text.length) {
            throw new Error('json: empty input');
        }
        const root = this.value(false);
        this.skip();
        if(this.position < this.text.length) {
            throw new Error(this.unexpected());
        }
        if(stringify && this.comments.length > 0) {
            throw new Error('json: Comment is not allowed in JSON');
        }
        return root;
    }
}
