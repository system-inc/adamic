import { panic, programArguments, readTextFile } from 'adamic';
import { lowerSource } from './lower.ts';
import { dump } from './dump.ts';
const input = readTextFile(programArguments()[0] ?? panic('missing corpus path'));
if(input.kind === 'Error') { panic(input.message); }
for(const source of input.text.split('\n')) {
    if(source === '') { continue; }
    const fn = lowerSource(source) ?? panic(`outside straight-line subset: ${source}`);
    console.log(dump(fn).trimEnd());
}
