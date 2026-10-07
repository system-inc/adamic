import { panic, programArguments, readTextFile } from 'adamic';
import { decode, encode } from '../codec.ts';
import { decodeString } from '../decodeString.ts';
const args = programArguments();
const input = readTextFile(args[0] ?? panic('usage: decode_probe.ts <escaped sources>'));
if(input.kind === 'Error') panic(input.message);
const output: string[] = [];
for(const line of input.text.split('\n')) {
    if(line === '') continue;
    output.push(encode(decodeString(decode(line.slice(1)))));
}
console.log(output.join('\n'));
