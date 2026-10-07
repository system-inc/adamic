import { panic, programArguments, readTextFile } from 'adamic';
import { inputChunks, serializeChunks } from '../inputChunks.ts';
const args = programArguments();
const input = readTextFile(args[0] ?? panic('usage: chunks_probe.ts <UTF-16 cases>'));
if(input.kind === 'Error') panic(input.message);
const output: string[] = [];
for(const line of input.text.split('\n')) {
    if(line === '') continue;
    const units: number[] = [];
    if(line !== '-') {
        for(const field of line.split(',')) {
            const unit = Number.parseInt(field, 10);
            if(!Number.isInteger(unit) || unit < 0 || unit > 65535) panic('UTF-16 unit');
            units.push(unit);
        }
    }
    output.push(serializeChunks(inputChunks(units)));
}
console.log(output.join('\n'));
