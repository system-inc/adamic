// Same line transport as GraphQL; canonical own-property JSON for comparisons.
import { panic, programArguments, readTextFile } from 'adamic';
import { parse } from './parser.ts';
import { render } from './nodes.ts';

// gap 5: panic is a statement in this helper, not a conditional expression.
function unescapeLetter(letter: string | undefined): string {
    // gap 6: case undefined is refused; narrow before the switch.
    if(letter === undefined) {
        panic('missing case escape');
    }
    switch(letter) {
        case 'n':
            return '\n';
        case 'r':
            return '\r';
        case 't':
            return '\t';
        case '\\':
            return '\\';
        default:
            panic('unknown case escape');
    }
}

function unescape(line: string): string {
    const parts: string[] = [];
    let start = 0;
    for(let index = 0; index < line.length; index++) {
        if(line[index] !== '\\') {
            continue;
        }
        // gap 1: push one value at a time.
        parts.push(line.slice(start, index));
        index++;
        const letter = line[index];
        parts.push(unescapeLetter(letter));
        start = index + 1;
    }
    parts.push(line.slice(start));
    return parts.join('');
}
const args = programArguments();
const read = readTextFile(args[0] ?? panic('usage: main.ts <cases file> [count]'));
if(read.kind === 'Error') {
    panic(read.message);
}
const texts: string[] = [];
for(const line of read.text.split('\n')) {
    if(line !== '') {
        texts.push(unescape(line.slice(1)));
    }
}
let count = 0;
let parsed = 0;
let nodes = 0;
for(let round = 0; round < (args[2] === 'repeat' ? 10 : 1); round++) {
    for(const css of texts) {
        const result = parse(css);
        if(args[1] !== 'count') {
            console.log(`case ${count}`);
            console.log(result.kind === 'Refused' ? `error ${result.message}` : render(result.nodes, 0, css));
        }
        count++;
        if(result.kind === 'Parsed') {
            parsed++;
            nodes += (result.nodes[0] ?? panic('missing root')).children.length;
        }
    }
}
if(args[1] === 'count') {
    console.log(`${parsed} of ${count} selectors parsed, ${nodes} nodes`);
}
