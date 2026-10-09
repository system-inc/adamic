// File-to-stdout driver for a selected Markdown leaf printer.
// main.ts <file> <mode>; --batch <escaped cases> is the audit/measurement protocol.
import { panic, programArguments, readTextFile, writeTextFile } from 'adamic';
import { formatLeaf } from './inline.ts';

function decode(text: string): string {
    const parts: string[] = [];
    let start = 0;
    for(let index = 0; index < text.length; index++) {
        if(text.charCodeAt(index) !== 92) continue;
        parts.push(text.slice(start, index));
        index++;
        const letter = text.slice(index, index + 1);
        parts.push(letter === 'n' ? '\n' : letter === 'r' ? '\r' : letter === 't' ? '\t' : letter);
        start = index + 1;
    }
    parts.push(text.slice(start));
    return parts.join('');
}
function encode(text: string): string {
    const parts: string[] = [];
    let start = 0;
    for(let index = 0; index < text.length; index++) {
        const code = text.charCodeAt(index);
        if(code !== 92 && code !== 10 && code !== 13 && code !== 9) continue;
        parts.push(text.slice(start, index));
        parts.push(code === 92 ? '\\\\' : code === 10 ? '\\n' : code === 13 ? '\\r' : '\\t');
        start = index + 1;
    }
    parts.push(text.slice(start));
    return parts.join('');
}
const args = programArguments();
const batch = args[0] === '--batch';
const path = args[batch ? 1 : 0] ?? panic('usage: main.ts <file> <mode> or --batch <cases>');
const read = readTextFile(path);
if(read.kind === 'Error') panic(read.message);
if(batch) {
    for(const line of read.text.split('\n')) {
        if(line === '') continue;
        console.log(encode(formatLeaf(line.slice(0, 1), decode(line.slice(1)))));
    }
}
else {
    // The prelude has no raw stdout function. /dev/stdout preserves terminal newlines on Linux.
    const written = writeTextFile('/dev/stdout', formatLeaf(args[1] ?? 'w', read.text));
    if(written.kind === 'Error') panic(written.message);
}
