// Native parser primitives on source text, not Go AST fixtures.
import { panic, programArguments, readTextFile } from 'adamic';
import { decode, encode } from '../codec.ts';
import { preprocess } from '../preprocess.ts';
import { quotePrefix } from '../quotePrefix.ts';
const args = programArguments();
const input = readTextFile(args[0] ?? panic('usage: prefix_probe.ts <source cases>'));
if(input.kind === 'Error') panic(input.message);
for(const line of input.text.split('\n')) {
    if(line === '') continue;
    const chunks = preprocess(decode(line.slice(1)));
    const fields: string[] = [];
    for(const chunk of chunks) fields.push(chunk.isText ? `t:${encode(chunk.text)}` : `c:${chunk.code}`);
    let result = encode(fields.join('\t'));
    for(let mode = 0; mode < 8; mode++) {
        const prefix = quotePrefix(chunks, (mode & 1) !== 0, (mode & 2) !== 0, (mode & 4) !== 0);
        result += `\t${prefix.matched ? 1 : 0},${prefix.consumed},${prefix.offset},${prefix.column},${prefix.open ? 1 : 0}`;
    }
    console.log(result);
}
