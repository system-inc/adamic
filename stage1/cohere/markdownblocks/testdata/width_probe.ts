import { panic, programArguments, readTextFile } from 'adamic';
import { decode } from '../codec.ts';
import { stringWidth } from '../width.ts';
const args = programArguments();
const input = readTextFile(args[0] ?? panic('usage: width_probe.ts <cases>'));
if(input.kind === 'Error') panic(input.message);
for(const line of input.text.split('\n')) {
    if(line === '') continue;
    console.log(`${stringWidth(decode(line.slice(1)))}`);
}
