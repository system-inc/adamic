// Format one file to stdout. --cases reads the escaped batch protocol used by the corpus test.
import { panic, programArguments, readTextFile } from 'adamic';
import { format } from './printer.ts';
import { defaults, type SettingsOptions } from './doc.ts';
export function escaped(text: string): string {
    let result = text;
    if(result.includes('\\')) {
        result = result.split('\\').join('\\\\');
    }
    if(result.includes('\n')) {
        result = result.split('\n').join('\\n');
    }
    if(result.includes('\r')) {
        result = result.split('\r').join('\\r');
    }
    if(result.includes('\t')) {
        result = result.split('\t').join('\\t');
    }
    return result;
}
function unescaped(text: string): string {
    if(!text.includes('\\')) {
        return text;
    }
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
const mode = args[2] ?? 'defaults';
const settings: SettingsOptions = {
    printWidth: mode === 'narrow' ? 80 : 120,
    tabWidth: mode === 'narrow' ? 2 : 4,
    useTabs: mode === 'tabs',
    bracketSpacing: mode !== 'tight',
};
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
        const text = unescaped(line.slice(1));
        const result = format(text, settings);
        console.log(result.kind === 'Ok' ? `ok\t${escaped(result.text)}` : `error\t${escaped(result.message)}`);
    }
}
else {
    const result = format(read.text, defaults);
    if(result.kind === 'Error') panic(result.message);
    console.log(result.text.slice(0, -1));
}
