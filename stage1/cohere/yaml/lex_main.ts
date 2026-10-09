// Lexical comparison driver. Tokens are UTF-16 hexadecimal units, including lone surrogates.
import { panic, programArguments, readTextFile } from 'adamic';
import { Lexer } from './lexer.ts';
function unescaped(text: string): string {
    const parts: string[] = [];
    let start = 0;
    for(let index = 0; index < text.length; index++) {
        if(text.slice(index, index + 1) !== '\\') continue;
        const prefix = text.slice(start, index);
        index++;
        const char = text.slice(index, index + 1);
        parts.push(prefix, char === 'n' ? '\n' : char === 'r' ? '\r' : char === 't' ? '\t' : char);
        start = index + 1;
    }
    parts.push(text.slice(start));
    return parts.join('');
}
function tokenText(token: string): string {
    const parts: string[] = ['='];
    for(let index = 0; index < token.length; index++) {
        parts.push(token.charCodeAt(index).toString(16).padStart(4, '0'));
    }
    return parts.join('');
}
const path = programArguments()[0] ?? panic('usage: lex_main.ts <cases>');
const read = readTextFile(path);
if(read.kind === 'Error') panic(read.message);
let number = 0;
for(const line of read.text.split('\n')) {
    if(line === '') continue;
    const tab = line.indexOf('\t');
    const chunkSize = Number(line.slice(0, tab));
    const text = unescaped(line.slice(tab + 1));
    const lexer = new Lexer();
    console.log(`case ${number}`);
    number++;
    if(chunkSize === 0) {
        for(const token of lexer.lex(text)) console.log(tokenText(token));
    }
    else {
        for(let start = 0; start < text.length; start += chunkSize) {
            for(const token of lexer.lex(text.slice(start, start + chunkSize), true)) console.log(tokenText(token));
        }
        for(const token of lexer.lex('', false)) console.log(tokenText(token));
    }
}
