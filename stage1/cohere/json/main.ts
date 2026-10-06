// Format one file to stdout. --cases reads the escaped batch protocol used by the corpus test.
import { panic, programArguments, readTextFile } from 'adamic';
import { format } from './formatter.ts';
export function escaped(text: string): string {
    return text.split('\\').join('\\\\').split('\n').join('\\n').split('\r').join('\\r').split('\t').join('\\t');
}
function unescaped(text: string): string {
    const pieces = text.split('\\');
    const parts: string[] = [pieces[0] ?? ''];
    for(let index = 1; index < pieces.length; index++) {
        const piece = pieces[index] ?? '';
        if(piece === '') {
            parts.push('\\');
            index++;
            parts.push(pieces[index] ?? '');
            continue;
        }
        const char = piece.slice(0, 1);
        parts.push(char === 'n' ? '\n' : char === 'r' ? '\r' : char === 't' ? '\t' : char);
        parts.push(piece.slice(1));
    }
    return parts.join('');
}
const args = programArguments();
const batch = args[0] === '--cases';
const path = args[batch ? 1 : 0] ?? panic('usage: main.ts <file> | --cases <cases>');
const read = readTextFile(path);
if(read.kind === 'Error') {
    panic(read.message);
}
if(batch) {
    for(const line of read.text.split('\n')) {
        if(line === '') {
            continue;
        }
        const tab = line.indexOf('\t');
        const name = line.slice(0, tab);
        const text = unescaped(line.slice(tab + 1));
        try {
            console.log(`ok\t${escaped(format(name, text))}`);
        }
        catch(error) {
            console.log(`error\t${escaped(error instanceof Error ? error.message : '?')}`);
        }
    }
}
else {
    try {
        console.log(format(path, read.text).slice(0, -1));
    }
    catch(error) {
        panic(error instanceof Error ? error.message : '?');
    }
}
