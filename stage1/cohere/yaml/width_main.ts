// Exhaustive width oracle driver: comma-separated Unicode code points per line.
import { panic, programArguments, readTextFile } from 'adamic';
import { stringWidth } from './width.ts';
const read = readTextFile(programArguments()[0] ?? panic('usage: width_main.ts <cases>'));
if(read.kind === 'Error') panic(read.message);
for(const line of read.text.split('\n')) {
    if(line === '') continue;
    const parts: string[] = [];
    for(const point of line.split(',')) parts.push(String.fromCodePoint(Number(point)));
    console.log(String(stringWidth(parts.join(''))));
}
