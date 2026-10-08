// Compare formatter-facing composed trees, document directives, and diagnostics.
import { panic, programArguments, readTextFile } from 'adamic';
import { CSTParser } from './cstParser.ts';
import { Composer } from './composer.ts';
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
class ComposedOutline {
    composer: Composer;
    parts: string[] = [];
    constructor(composer: Composer) {
        this.composer = composer;
    }
    token(index: number): string {
        if(index < 0) return '-';
        const token = this.composer.parser.get(index);
        return `${token.type},${token.offset},${hex(token.source)}`;
    }
    outline(index: number): void {
        if(index < 0) {
            this.parts.push('-');
            return;
        }
        const node = this.composer.node(index);
        this.parts.push(
            `${node.kind}|${node.className}|${node.range.length === 0 ? '-' : node.range.join(',')}|${this.token(node.sourceToken)}|${node.sourceItem >= 0 ? 1 : 0}|${node.anchorPresent ? 1 : 0}|${hex(node.anchor)}|${hex(node.tag)}|${hex(node.comment)}|${hex(node.commentBefore)}|${node.spaceBefore ? 1 : 0}|${node.flow ? 1 : 0}|${node.sourcePresent ? 1 : 0}|${hex(node.source)}|${node.type}|${node.format}|${node.minFractionDigits}|[`,
        );
        for(const item of node.items) {
            this.outline(item);
            this.parts.push(',');
        }
        this.parts.push(']|');
        this.outline(node.key);
        this.parts.push('|');
        this.outline(node.value);
    }
}
const path = programArguments()[0] ?? panic('usage: compose_main.ts <cases>');
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
    const composer = new Composer(parser);
    composer.compose(text.length);
    for(const document of composer.documents) {
        const writer = new ComposedOutline(composer);
        writer.outline(document.contents);
        const tags: string[] = [];
        for(const [handle, prefix] of document.directives.tags) tags.push(`${hex(handle)}=${hex(prefix)}`);
        tags.sort((first, second) => (first < second ? -1 : first > second ? 1 : 0));
        const errors: string[] = [];
        for(const error of document.diagnostics.errors)
            errors.push(`${error.start},${error.end},${error.code},${hex(error.message)}`);
        const warnings: string[] = [];
        for(const warning of document.warnings)
            warnings.push(`${warning.start},${warning.end},${warning.code},${hex(warning.message)}`);
        console.log(
            `${document.directives.version}|${document.directives.explicit ? 1 : 0}|${document.directives.docStart ? 1 : 0}|${document.directives.docEnd ? 1 : 0}|${tags.join(';')}|${document.range.join(',')}|${hex(document.commentBefore)}|${hex(document.comment)}|${writer.parts.join('')}|errors:${errors.join(';')}|warnings:${warnings.join(';')}`,
        );
    }
}
