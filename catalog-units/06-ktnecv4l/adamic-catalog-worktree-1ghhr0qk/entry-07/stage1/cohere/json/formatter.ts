// JSON paths through cohere's print_json.go, print_object.go, print_array.go and utility_text.go.
import { panic } from 'adamic';
import { Reader, numericValue } from './parser.ts';
import { Documents } from './doc.ts';
export function stringifyName(name: string): boolean {
    const pieces = name.split('/');
    const base = pieces[pieces.length - 1] ?? '';
    return base === 'package.json' || base === 'package-lock.json' || base === 'composer.json';
}
function numberText(raw: string): string {
    if(raw.length === 1) {
        return raw;
    }
    let value = raw.toLowerCase();
    const exponent = value.indexOf('e');
    let decimalPrefix = exponent > 0;
    for(let index = 0; index < exponent; index++) {
        const char = value.slice(index, index + 1);
        if((char < '0' || char > '9') && char !== '.' && !(index === 0 && (char === '+' || char === '-'))) {
            decimalPrefix = false;
        }
    }
    if(decimalPrefix) {
        let tail = value.slice(exponent + 1);
        const negative = tail.startsWith('-');
        if(tail.startsWith('-') || tail.startsWith('+')) {
            tail = tail.slice(1);
        }
        if(tail.charCodeAt(0) >= 48 && tail.charCodeAt(0) <= 57) {
            while(tail.length > 1 && tail.startsWith('0') && tail.charCodeAt(1) >= 48 && tail.charCodeAt(1) <= 57) {
                tail = tail.slice(1);
            }
            value = value.slice(0, exponent) + (tail === '0' ? '' : `e${negative ? '-' : ''}${tail}`);
        }
    }
    if(value.startsWith('.')) {
        value = `0${value}`;
    }
    const exp = value.indexOf('e');
    const split = exp < 0 || value.startsWith('0x') ? value.length : exp;
    let mantissa = value.slice(0, split);
    const dot = mantissa.indexOf('.');
    if(dot >= 0 && !mantissa.slice(dot + 1).includes('_')) {
        while(mantissa.endsWith('0') && mantissa.length > dot + 2) {
            mantissa = mantissa.slice(0, -1);
        }
    }
    if(mantissa.endsWith('.')) {
        mantissa = mantissa.slice(0, -1);
    }
    return mantissa + value.slice(split);
}
function stringText(raw: string): string {
    if(raw.startsWith('"')) {
        return raw;
    }
    const parts: string[] = ['"'];
    for(let index = 1; index < raw.length - 1; index++) {
        const char = raw.slice(index, index + 1);
        if(char === '\\' && index + 1 < raw.length - 1) {
            const next = raw.slice(index + 1, index + 2);
            if(next === "'") {
                parts.push(next);
                index++;
            }
            else if(next === '"' || next === '\\') {
                parts.push(char);
                parts.push(next);
                index++;
            }
            else {
                parts.push(char);
            }
        }
        else if(char === '"') {
            parts.push('\\"');
        }
        else {
            parts.push(char);
        }
    }
    parts.push('"');
    return parts.join('');
}
// cohere's decodeJavaScriptString for the json-stringify template-literal path.
function unicodeEnd(text: string, start: number): number {
    return text.slice(start, start + 1) === '{' ? text.indexOf('}', start) + 1 : start + 4;
}
function unicodeCode(text: string, start: number, end: number): number {
    return Number.parseInt(
        text.slice(start, start + 1) === '{' ? text.slice(start + 1, end - 1) : text.slice(start, end),
        16,
    );
}
function cookedTemplate(raw: string): string {
    const text = raw.slice(1, -1).split('\r\n').join('\n').split('\r').join('\n');
    const parts: string[] = [];
    for(let index = 0; index < text.length; index++) {
        const char = text.slice(index, index + 1);
        if(char !== '\\') {
            parts.push(char);
            continue;
        }
        index++;
        const escaped = text.slice(index, index + 1);
        if(escaped === 'n') {
            parts.push('\n');
        }
        else if(escaped === 'r') {
            parts.push('\r');
        }
        else if(escaped === 't') {
            parts.push('\t');
        }
        else if(escaped === 'b') {
            parts.push('\b');
        }
        else if(escaped === 'f') {
            parts.push('\f');
        }
        else if(escaped === 'v') {
            parts.push('\v');
        }
        else if(escaped === '0') {
            parts.push('\0');
        }
        else if(escaped === '\n' || escaped === '\u2028' || escaped === '\u2029') {
            continue;
        }
        else if(escaped === 'x') {
            parts.push(String.fromCharCode(Number.parseInt(text.slice(index + 1, index + 3), 16)));
            index += 2;
        }
        else if(escaped === 'u') {
            const end = unicodeEnd(text, index + 1);
            let code = unicodeCode(text, index + 1, end);
            index = end - 1;
            if(code >= 0xd800 && code <= 0xdbff && text.slice(end, end + 2) === '\\u') {
                const lowEnd = unicodeEnd(text, end + 2);
                const low = unicodeCode(text, end + 2, lowEnd);
                if(low >= 0xdc00 && low <= 0xdfff) {
                    code = (code - 0xd800) * 1024 + low - 0xdc00 + 0x10000;
                    index = lowEnd - 1;
                }
            }
            if(code >= 0xd800 && code <= 0xdfff) {
                code = 0xfffd;
            }
            if(code > 0xffff) {
                parts.push(
                    String.fromCharCode(
                        0xd800 + Math.floor((code - 0x10000) / 1024),
                        0xdc00 + ((code - 0x10000) % 1024),
                    ),
                );
            }
            else {
                parts.push(String.fromCharCode(code));
            }
        }
        else {
            parts.push(escaped);
        }
    }
    return parts.join('');
}
function jsonQuote(text: string): string {
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
        else if(char === '\b') {
            parts.push('\\b');
        }
        else if(char === '\f') {
            parts.push('\\f');
        }
        else if(code < 32) {
            parts.push('\\u');
            parts.push(code.toString(16).padStart(4, '0'));
        }
        else {
            parts.push(char);
        }
    }
    parts.push('"');
    return parts.join('');
}
class Printer {
    readonly reader: Reader;
    readonly stringify: boolean;
    readonly hasComments: boolean;
    readonly docs = new Documents();
    readonly leading: number[][] = [];
    readonly trailing: number[][] = [];
    readonly dangling: number[][] = [];
    constructor(reader: Reader, stringify: boolean) {
        this.reader = reader;
        this.stringify = stringify;
        this.hasComments = reader.comments.length > 0;
        if(this.hasComments) {
            for(let count = reader.nodes.length; count > 0; count--) {
                this.leading.push([]);
                this.trailing.push([]);
                this.dangling.push([]);
            }
        }
    }
    // Attach to the nearest property/element within the smallest enclosing container.
    attach(root: number): void {
        for(let commentIndex = 0; commentIndex < this.reader.comments.length; commentIndex++) {
            const comment = this.reader.comments[commentIndex] ?? panic('missing comment');
            let container = root;
            let span = this.reader.text.length + 1;
            for(let index = 0; index < this.reader.nodes.length; index++) {
                const node = this.reader.get(index);
                if(
                    (node.kind === 'ObjectExpression' || node.kind === 'ArrayExpression') &&
                    node.start < comment.start &&
                    node.end > comment.end &&
                    node.end - node.start < span
                ) {
                    container = index;
                    span = node.end - node.start;
                }
            }
            const node = this.reader.get(container);
            if(comment.end <= node.start) {
                (this.leading[container] ?? panic('missing leading')).push(commentIndex);
                continue;
            }
            if(comment.start >= node.end) {
                (this.trailing[container] ?? panic('missing trailing')).push(commentIndex);
                continue;
            }
            let previous = -1;
            let next = -1;
            for(const child of node.children) {
                const item = this.reader.get(child);
                if(item.end <= comment.start) {
                    previous = child;
                }
                if(next < 0 && item.start >= comment.end) {
                    next = child;
                }
            }
            const ownLine = this.reader.text
                .slice(previous < 0 ? node.start : this.reader.get(previous).end, comment.start)
                .includes('\n');
            if(previous >= 0 && !ownLine && comment.line) {
                (this.trailing[previous] ?? panic('missing trailing')).push(commentIndex);
            }
            else if(next >= 0) {
                (this.leading[next] ?? panic('missing leading')).push(commentIndex);
            }
            else if(previous >= 0) {
                (this.trailing[previous] ?? panic('missing trailing')).push(commentIndex);
            }
            else {
                (this.dangling[container] ?? panic('missing dangling')).push(commentIndex);
            }
        }
    }
    blankAfter(end: number): boolean {
        let index = end;
        while(index < this.reader.text.length && ' ,\t'.includes(this.reader.text.slice(index, index + 1))) {
            index++;
        }
        if(this.reader.text.slice(index, index + 2) === '\r\n') {
            index++;
        }
        if(this.reader.text.slice(index, index + 1) !== '\n') {
            return false;
        }
        index++;
        while(' \t\r'.includes(this.reader.text.slice(index, index + 1)) && index < this.reader.text.length) {
            index++;
        }
        return this.reader.text.slice(index, index + 1) === '\n';
    }
    print(index: number, key: boolean): number {
        const node = this.reader.get(index);
        const documents = this.docs;
        let result: number;
        if(node.kind === 'ObjectProperty') {
            const left = this.print(node.children[0] ?? panic('missing key'), true);
            const right = this.print(node.children[1] ?? panic('missing value'), false);
            result = documents.group(documents.concat([left, documents.text(': '), right]), false);
        }
        else if(node.kind === 'UnaryExpression') {
            result = documents.concat([
                documents.text(this.stringify && node.raw === '+' ? '' : node.raw),
                this.print(node.children[0] ?? panic('missing argument'), false),
            ]);
        }
        else if(node.kind === 'ArrayExpression' || node.kind === 'ObjectExpression') {
            const object = node.kind === 'ObjectExpression';
            const open = object ? '{' : '[';
            const close = object ? '}' : ']';
            const parts: number[] = [];
            let concise = !object && !this.stringify && node.children.length > 0;
            let sameNested = !object && node.children.length > 1;
            let previousKind = '';
            for(const child of node.children) {
                const item = this.reader.get(child);
                if(
                    item.kind !== 'NumericLiteral' &&
                    !(
                        item.kind === 'UnaryExpression' &&
                        this.reader.get(item.children[0] ?? panic('missing signed number')).kind === 'NumericLiteral'
                    )
                ) {
                    concise = false;
                }
                if(
                    (item.kind !== 'ObjectExpression' && item.kind !== 'ArrayExpression') ||
                    item.children.length <= 1 ||
                    (previousKind !== '' && item.kind !== previousKind)
                ) {
                    sameNested = false;
                }
                previousKind = item.kind;
                if(
                    this.hasComments &&
                    ((this.leading[child] ?? panic('missing comment list')).length > 0 ||
                        (this.trailing[child] ?? panic('missing comment list')).length > 0)
                ) {
                    concise = false;
                }
            }
            let broken = this.stringify || sameNested;
            if(
                object &&
                node.children.length > 0 &&
                this.reader.text
                    .slice(node.start, this.reader.get(node.children[0] ?? panic('missing first property')).start)
                    .includes('\n')
            ) {
                broken = true;
            }
            for(let childIndex = 0; childIndex < node.children.length; childIndex++) {
                const child = node.children[childIndex] ?? panic('missing container child');
                parts.push(this.print(child, false));
                if(childIndex + 1 < node.children.length) {
                    if(concise) {
                        const printed = parts.pop() ?? panic('missing numeric doc');
                        parts.push(documents.concat([printed, documents.text(',')]));
                    }
                    else {
                        parts.push(documents.text(','));
                    }
                    parts.push(documents.line(false, this.stringify));
                    if(!this.stringify && this.blankAfter(this.reader.get(child).end)) {
                        parts.push(documents.line(true, true));
                    }
                }
            }
            if(this.hasComments) {
                for(const commentIndex of this.dangling[index] ?? panic('missing comment list')) {
                    const comment = this.reader.comments[commentIndex] ?? panic('missing dangling comment');
                    parts.push(documents.text(comment.text));
                    if(comment.line || this.reader.text.slice(node.start, comment.start).includes('\n')) {
                        broken = true;
                    }
                }
            }
            if(node.children.length === 0 && (parts.length === 0 || (!broken && parts.length > 0))) {
                result = documents.concat([documents.text(open), documents.concat(parts), documents.text(close)]);
            }
            else {
                if(
                    !object &&
                    !this.stringify &&
                    node.children.length > 0 &&
                    this.reader.get(node.children[node.children.length - 1] ?? panic('missing last element')).kind ===
                        'Hole'
                ) {
                    parts.push(documents.text(','));
                }
                const spacing = documents.line(!object, this.stringify);
                const body = concise ? documents.fill(parts) : documents.concat(parts);
                result = documents.group(
                    documents.concat([
                        documents.text(open),
                        documents.indent(documents.concat([spacing, body])),
                        spacing,
                        documents.text(close),
                    ]),
                    broken,
                );
            }
        }
        else if(node.kind === 'StringLiteral') {
            result = documents.text(stringText(node.raw));
        }
        else if(node.kind === 'NumericLiteral') {
            const raw = this.stringify ? node.raw : numberText(node.raw);
            const canonical = key ? `${numericValue(node.raw)}` : '';
            const numericKey =
                key &&
                (this.stringify
                    ? canonical === node.raw
                    : canonical === raw &&
                      !raw.includes('e') &&
                      !raw.startsWith('0x') &&
                      !raw.startsWith('0o') &&
                      !raw.startsWith('0b') &&
                      !raw.includes('_'));
            result = documents.text(numericKey ? `"${this.stringify ? raw : numericValue(node.raw)}"` : raw);
        }
        else if(node.kind === 'Identifier') {
            result = documents.text(key ? `"${node.raw}"` : node.raw);
        }
        else if(node.kind === 'Hole') {
            result = documents.text(this.stringify ? 'null' : '');
        }
        else if(node.kind === 'TemplateLiteral') {
            result = documents.text(this.stringify ? jsonQuote(cookedTemplate(node.raw)) : node.raw);
        }
        else {
            result = documents.text(node.raw);
        }
        // No comment wrapper is needed when both lists are empty.
        if(
            !this.hasComments ||
            ((this.leading[index] ?? panic('missing comment list')).length === 0 &&
                (this.trailing[index] ?? panic('missing comment list')).length === 0)
        ) {
            return result;
        }
        const before: number[] = [];
        for(const commentIndex of this.leading[index] ?? panic('missing comment list')) {
            const comment = this.reader.comments[commentIndex] ?? panic('missing leading comment');
            before.push(documents.text(comment.text));
            before.push(
                documents.line(false, comment.line || this.reader.text.slice(comment.end, node.start).includes('\n')),
            );
            if(this.blankAfter(comment.end)) {
                before.push(documents.line(false, true));
            }
        }
        before.push(result);
        for(const commentIndex of this.trailing[index] ?? panic('missing comment list')) {
            const comment = this.reader.comments[commentIndex] ?? panic('missing trailing comment');
            const ownLine = this.reader.text.slice(node.end, comment.start).includes('\n');
            if(ownLine) {
                before.push(documents.line(false, true));
                before.push(documents.text(comment.text));
            }
            else if(comment.line) {
                before.push(documents.add('suffix', '', [documents.text(' '), documents.text(comment.text)], true));
            }
            else {
                before.push(documents.text(' '));
                before.push(documents.text(comment.text));
            }
        }
        return documents.concat(before);
    }
}
export function format(name: string, text: string): string {
    const stringify = stringifyName(name);
    const reader = new Reader(text);
    const root = reader.parse(stringify);
    const printer = new Printer(reader, stringify);
    printer.attach(root);
    const document = printer.print(root, false);
    return printer.docs.print(printer.docs.concat([document, printer.docs.line(false, true)]));
}
