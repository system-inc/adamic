// Same line transport as GraphQL; canonical own-property JSON for comparisons.
import { panic, programArguments, readTextFile } from 'adamic';
import { compose } from './compose.ts';
import { renderTree } from './tree.ts';

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
const modes: boolean[] = [];
for(const line of read.text.split('\n')) {
    if(line !== '') {
        texts.push(unescape(line.slice(2)));
        modes.push(line[1] === 'S');
    }
}
let count = 0;
let parsed = 0;
let nodes = 0;
for(let round = 0; round < (args[2] === 'repeat' ? 10 : 1); round++) {
    for(let index = 0; index < texts.length; index++) {
        const css = texts[index] ?? panic('missing stylesheet');
        const result = compose(css, modes[index] ?? false);
        if(args[1] !== 'count') {
            console.log(`case ${count}`);
            console.log(
                result.kind === 'Refused' ? `error ${result.message}` : renderTree(result.tree, result.tree.root),
            );
        }
        count++;
        if(result.kind === 'Composed') {
            parsed++;
            nodes += result.tree.at(result.tree.root).list('nodes').length;
        }
    }
}
if(args[1] === 'count') {
    console.log(`${parsed} of ${count} stylesheets parsed, ${nodes} nodes`);
}
