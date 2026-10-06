// CST comparison protocol includes absent fields, null keys, UTF-16 offsets and line starts.
import { panic, programArguments, readTextFile } from 'adamic';
import { CSTParser } from './cstParser.ts';
function unescaped(text: string): string {
    const parts: string[] = [];
    let start = 0;
    for(let index = 0; index < text.length; index++) {
        if(text.slice(index, index + 1) !== '\\') continue;
        parts.push(text.slice(start, index));
        index++;
        const char = text.slice(index, index + 1);
        parts.push(char === 'n' ? '\n' : char === 'r' ? '\r' : char === 't' ? '\t' : char);
        start = index + 1;
    }
    parts.push(text.slice(start));
    return parts.join('');
}
function hex(text: string): string {
    const parts: string[] = [];
    for(let index = 0; index < text.length; index++) parts.push(text.charCodeAt(index).toString(16).padStart(4, '0'));
    return parts.join('');
}
class Outline {
    parts: string[] = [];
    parser: CSTParser;
    constructor(parser: CSTParser) {
        this.parser = parser;
    }
    list(children: readonly number[]): void {
        this.parts.push('[');
        for(const child of children) {
            this.outline(child);
            this.parts.push(',');
        }
        this.parts.push(']');
    }
    outline(index: number): void {
        if(index < 0) {
            this.parts.push('-');
            return;
        }
        const token = this.parser.get(index);
        this.parts.push(
            `${token.type}|${token.offset}|${token.indentPresent ? String(token.indent) : '-'}|${hex(token.source)}|${hex(token.message)}|`,
        );
        this.list(token.start);
        this.parts.push('|');
        this.outline(token.flowStart);
        this.parts.push('|');
        this.outline(token.value);
        this.parts.push('|');
        if(token.endPresent) this.list(token.end);
        else this.parts.push('-');
        this.parts.push('|');
        this.list(token.props);
        this.parts.push('|[');
        for(const item of token.items) {
            this.list(item.start);
            this.parts.push(`|${item.keyPresent ? 1 : 0}|`);
            this.outline(item.key);
            this.parts.push('|');
            if(item.sepPresent) this.list(item.sep);
            else this.parts.push('-');
            this.parts.push('|');
            this.outline(item.value);
            this.parts.push(`|${item.explicitKey ? 1 : 0},`);
        }
        this.parts.push(']');
    }
}
const path = programArguments()[0] ?? panic('usage: cst_main.ts <cases>');
const read = readTextFile(path);
if(read.kind === 'Error') panic(read.message);
let number = 0;
for(const line of read.text.split('\n')) {
    if(line === '') continue;
    const tab = line.indexOf('\t');
    const size = Number(line.slice(0, tab));
    const text = unescaped(line.slice(tab + 1));
    const parser = new CSTParser();
    console.log(`case ${number}`);
    number++;
    if(size === 0) parser.parse(text);
    else {
        for(let start = 0; start < text.length; start += size) parser.parse(text.slice(start, start + size), true);
        parser.parse('', false);
    }
    for(const root of parser.roots) {
        const writer = new Outline(parser);
        writer.outline(root);
        console.log(writer.parts.join(''));
    }
    const lines: string[] = [];
    for(const start of parser.lineStarts) lines.push(String(start));
    console.log(`lines ${lines.join(',')}`);
}
