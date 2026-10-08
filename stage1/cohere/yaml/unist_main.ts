// Compare the printer tree, source spans, parents, comments, and parse failures.
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
import { UnistContext } from './unistContext.ts';
function byteCanonical(text: string, offset: number): number {
    const code = text.charCodeAt(offset);
    const previous = text.charCodeAt(offset - 1);
    return code >= 0xdc00 && code <= 0xdfff && previous >= 0xd800 && previous <= 0xdbff ? offset - 1 : offset;
}
class UnistOutline {
    parts: string[] = [];
    context: UnistContext;
    constructor(context: UnistContext) {
        this.context = context;
    }
    point(index: number, end: boolean): string {
        const point = end ? this.context.end(index) : this.context.start(index);
        return `${point.line},${point.column},${byteCanonical(this.context.text, point.offset)}`;
    }
    list(items: readonly number[]): void {
        this.parts.push('[');
        for(const item of items) {
            this.outline(item);
            this.parts.push(',');
        }
        this.parts.push(']');
    }
    outline(index: number): void {
        if(index < 0) {
            this.parts.push('-');
            return;
        }
        const node = this.context.node(index);
        const parent =
            node.parent < 0
                ? '-'
                : `${this.context.node(node.parent).type},${this.point(node.parent, false)},${this.point(node.parent, true)}`;
        this.parts.push(
            `${node.type}|${this.point(index, false)}|${this.point(index, true)}|${parent}|${hex(node.value)}|${node.chomping}|${node.indent}|${node.directivesEndMarker ? 1 : 0}|${node.documentEndMarker ? 1 : 0}|${hex(node.name)}|${node.parameters.map(hex).join(',')}|`,
        );
        this.list(node.children);
        this.parts.push('|');
        this.outline(node.tag);
        this.parts.push('|');
        this.outline(node.anchor);
        this.parts.push('|');
        this.list(node.middleComments);
        this.parts.push('|');
        this.list(node.leadingComments);
        this.parts.push('|');
        this.outline(node.trailingComment);
        this.parts.push('|');
        this.list(node.endComments);
        this.parts.push('|');
        this.outline(node.indicatorComment);
        this.parts.push('|');
        this.list(node.comments);
    }
}
const path = programArguments()[0] ?? panic('usage: compose_main.ts <cases>');
const read = readTextFile(path);
if(read.kind === 'Error') panic(read.message);
let number = 0;
for(const line of read.text.split('\n')) {
    if(line === '') continue;
    const tab = line.indexOf('\t');
    const text = unescaped(line.slice(tab + 1));
    const parser = new CSTParser();
    console.log(`case ${number}`);
    number++;
    parser.parse(text);
    const composer = new Composer(parser);
    composer.compose(text.length);
    const context = new UnistContext(text, parser, composer);
    let syntax = false;
    for(const document of composer.documents) {
        if(document.diagnostics.errors.length === 0) continue;
        const error = document.diagnostics.errors[0];
        if(error === undefined) continue;
        const start = context.point(context.offset(error.start));
        const end = context.point(context.offset(error.end));
        console.log(
            `syntax|${error.code}|${hex(error.message)}|${start.line},${start.column},${byteCanonical(text, start.offset)}|${end.line},${end.column},${byteCanonical(text, end.offset)}`,
        );
        syntax = true;
        break;
    }
    if(syntax) continue;
    const root = context.parse();
    if(context.errorName !== '') console.log(`throw|${context.errorName}|${hex(context.errorMessage)}`);
    else {
        const writer = new UnistOutline(context);
        writer.outline(root);
        console.log(writer.parts.join(''));
    }
}
