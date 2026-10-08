import { panic, programArguments, readTextFile, writeTextFile } from 'adamic';
import { format } from './format.ts';
function escaped(text: string): string {
    return text.split('\\').join('\\\\').split('\n').join('\\n').split('\r').join('\\r').split('\t').join('\\t');
}
function unescaped(text: string): string {
    const parts: string[] = [];
    let start = 0;
    for(let index = 0; index < text.length; index++) {
        if(text.charCodeAt(index) !== 92) continue;
        parts.push(text.slice(start, index));
        index++;
        const character = text.slice(index, index + 1);
        parts.push(character === 'n' ? '\n' : character === 'r' ? '\r' : character === 't' ? '\t' : character);
        start = index + 1;
    }
    parts.push(text.slice(start));
    return parts.join('');
}
const args = programArguments();
const batch = args[0] === '--cases';
const path = args[batch ? 1 : 0] ?? panic('usage: main.ts <file> | --cases <cases>');
const read = readTextFile(path);
if(read.kind === 'Error') panic(read.message);
if(batch) {
    for(const line of read.text.split('\n')) {
        if(line === '') continue;
        const tab = line.indexOf('\t');
        const result = format(unescaped(line.slice(tab + 1)));
        console.log(
            `${result.kind === 'Ok' ? 'ok' : 'error'}\t${escaped(result.kind === 'Ok' ? result.text : result.message)}`,
        );
    }
}
else {
    const result = format(read.text);
    if(result.kind === 'Error') panic(result.message);
    const written = writeTextFile('/dev/stdout', result.text);
    if(written.kind === 'Error') panic(written.message);
}
