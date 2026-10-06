// Resolve property runs under the same contexts as Go's YAML composer.
import { panic, programArguments, readTextFile } from 'adamic';
import { CSTParser } from './cstParser.ts';
import { PropsResolver } from './propsResolver.ts';
import { ScalarResolution } from './scalarResolution.ts';
function unescaped(text: string): string {
    const parts: string[] = [];
    let start = 0;
    for(let index = 0; index < text.length; index++) {
        if(text.slice(index, index + 1) !== '\\') continue;
        parts.push(text.slice(start, index));
        index++;
        const character = text.slice(index, index + 1);
        parts.push(character === 'n' ? '\n' : character === 'r' ? '\r' : character === 't' ? '\t' : character);
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
class PropsOutline {
    parser: CSTParser;
    constructor(parser: CSTParser) {
        this.parser = parser;
    }
    token(index: number): string {
        if(index < 0) return '-';
        const token = this.parser.get(index);
        return `${token.type},${token.offset},${hex(token.source)}`;
    }
    resolve(
        tokens: readonly number[],
        flow: string,
        indicator: string,
        next: number,
        offset: number,
        indent: number,
        newline: boolean,
    ): void {
        const diagnostics = new ScalarResolution();
        const resolver = new PropsResolver(this.parser, diagnostics);
        const props = resolver.resolve(tokens, flow, indicator, next, offset, indent, newline);
        const errors: string[] = [];
        for(const error of diagnostics.errors)
            errors.push(`${error.start},${error.end},${error.code},${hex(error.message)}`);
        const warnings: string[] = [];
        for(const warning of props.warnings)
            warnings.push(`${warning.start},${warning.end},${warning.code},${hex(warning.message)}`);
        console.log(
            `${this.token(props.comma)}|${this.token(props.found)}|${props.spaceBefore ? 1 : 0}|${hex(props.comment)}|${props.hasNewline ? 1 : 0}|${this.token(props.anchor)}|${this.token(props.tag)}|${this.token(props.newlineAfterProp)}|${props.end}|${props.start}|${errors.join(';')}|${warnings.join(';')}`,
        );
    }
    visit(index: number): void {
        if(index < 0) return;
        const token = this.parser.get(index);
        if(token.type === 'document')
            this.resolve(
                token.start,
                '',
                'doc-start',
                token.value >= 0 ? token.value : (token.end[0] ?? -1),
                token.offset,
                0,
                true,
            );
        else {
            const flow =
                token.type === 'flow-collection'
                    ? this.parser.get(token.flowStart).source === '{'
                        ? 'flow map'
                        : 'flow sequence'
                    : '';
            for(const item of token.items) {
                const next = item.key >= 0 ? item.key : (item.sep[0] ?? -1);
                this.resolve(
                    item.start,
                    flow,
                    token.type === 'block-seq' ? 'seq-item-ind' : 'explicit-key-ind',
                    token.type === 'block-seq' ? item.value : next,
                    token.offset,
                    token.indent,
                    flow === '',
                );
                if(item.sepPresent)
                    this.resolve(item.sep, flow, 'map-value-ind', item.value, token.offset, token.indent, false);
            }
        }
        this.visit(token.value);
        for(const item of token.items) {
            this.visit(item.key);
            this.visit(item.value);
        }
    }
}
const path = programArguments()[0] ?? panic('usage: props_main.ts <cases>');
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
    const writer = new PropsOutline(parser);
    for(const root of parser.roots) writer.visit(root);
}
