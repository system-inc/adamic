// Only yaml 2.9.0, with no port code or generated Adamic output.
import { createRequire } from 'node:module';
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
const [library, path] = process.argv.slice(2);
const require = createRequire(join(library, 'package.json'));
const version = require('yaml/package.json').version;
if (version !== '2.9.0') throw new Error(`expected yaml 2.9.0, got ${version}`);
const { Lexer } = require('yaml');
const unescape = text => text.replace(/\\([\\nrt])/g, (_, ch) => ch === 'n' ? '\n' : ch === 'r' ? '\r' : ch === 't' ? '\t' : ch);
let number = 0;
for (const line of readFileSync(path, 'utf8').split('\n')) {
    if (line === '') continue;
    const tab = line.indexOf('\t');
    const size = Number(line.slice(0, tab));
    const text = unescape(line.slice(tab + 1));
    const lexer = new Lexer();
    console.log(`case ${number++}`);
    function emit(source, incomplete) {
        for (const token of lexer.lex(source, incomplete)) {
            let written = '=';
            for (let i = 0; i < token.length; i++) written += token.charCodeAt(i).toString(16).padStart(4, '0');
            console.log(written);
        }
    }
    if (size === 0) emit(text, false);
    else {
        for (let start = 0; start < text.length; start += size) emit(text.slice(start, start + size), true);
        emit('', false);
    }
}
