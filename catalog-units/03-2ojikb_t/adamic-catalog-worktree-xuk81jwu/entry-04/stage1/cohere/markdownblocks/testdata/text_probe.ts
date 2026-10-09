import { panic, programArguments, readTextFile } from 'adamic';
import { decode, encode } from '../codec.ts';
import { splitText } from '../splitText.ts';
const args = programArguments();
const input = readTextFile(args[0] ?? panic('usage: text_probe.ts <cases>'));
if(input.kind === 'Error') panic(input.message);
const output: string[] = [];
for(const line of input.text.split('\n')) {
    if(line === '') continue;
    const fields: string[] = [];
    for(const token of splitText(decode(line.slice(1))))
        fields.push(
            `${token.type}\t${token.kind}\t${token.cj ? 1 : 0},${token.leading ? 1 : 0},${token.trailing ? 1 : 0}\t${encode(token.value)}`,
        );
    output.push(encode(fields.join('\n')));
}
console.log(output.join('\n'));
