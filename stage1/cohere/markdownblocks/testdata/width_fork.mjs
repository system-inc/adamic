// Observe the fork's private StringWidth through its real group-fit decisions.
import fs from 'node:fs';
import {createRequire} from 'node:module';
const prettier = createRequire(import.meta.url)(process.argv[2] + '/standalone.js');
if (prettier.version !== '3.9.6') throw new Error('expected pinned fork 3.9.6');
const {group, line} = prettier.doc.builders;
const decode = text => text.replace(/\\(.)/g, (_, c) => c === 'n' ? '\n' : c === 'r' ? '\r' : c === 't' ? '\t' : c);
const output = [];
for (const source of fs.readFileSync(process.argv[3], 'utf8').split('\n')) {
    if (source === '') continue;
    const text = decode(source.slice(1));
    let low = 3, high = text.length * 2 + 3;
    while (low < high) {
        const middle = Math.floor((low + high) / 2);
        const rendered = prettier.doc.printer.printDocToString(group(['@', text, line, '@']), {printWidth: middle, tabWidth: 4}).formatted;
        if (rendered.endsWith(' @')) high = middle;
        else low = middle + 1;
    }
    output.push(String(low - 3));
}
process.stdout.write(output.join('\n') + '\n');
