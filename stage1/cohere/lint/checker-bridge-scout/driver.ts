// Measurement client of the existing scalar ABI; this does not replace Checker.ask.
import { panic, programArguments, readTextFile, tsgoProgram, tsgoInspect, tsgoRelease } from 'adamic';
import { Frames } from '../../typeaware/frames.ts';
function read(path: string): string {
    const result = readTextFile(path);
    if(result.kind === 'Error') { panic(result.message); }
    return result.text;
}
const args = programArguments();
const config = args[0] ?? panic('missing config');
const requests = read(args[1] ?? panic('missing requests'));
const repetitions = parseInt(args[2] ?? '1', 10);
if(!Number.isSafeInteger(repetitions) || repetitions < 1) { panic('invalid repetitions'); }
const program = tsgoProgram(config, []);
const useCache = args[3] === 'cache';
const cached: string[] = [];
let queries = 0;
for(let pass = 0; pass < repetitions; pass++) {
    const input = new Frames(requests);
    const count = input.natural();
    for(let index = 0; index < count; index++) {
        const file = input.field();
        const start = input.natural();
        const end = input.natural();
        const kind = input.field();
        const question = input.field();
        const expected = input.field();
        const value = useCache && pass > 0 ? cached[index] ?? panic('cache miss') : tsgoInspect(program, file, start, end, kind, question);
        if(useCache && pass === 0) { cached.push(value); }
        if(value !== expected) { panic('scout fact mismatch'); }
        if(pass === 0) { console.log(`${value.length}\n${value}`); }
        queries++;
    }
    input.end();
}
console.log(`queries ${queries}`);
tsgoRelease(program);
