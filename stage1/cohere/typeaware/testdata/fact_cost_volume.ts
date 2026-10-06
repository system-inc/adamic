import { panic, programArguments, tsgoProgram, tsgoInspect, tsgoRelease } from 'adamic';
import { types } from '../facts.ts';
const args = programArguments();
const path = args[1] ?? panic('file');
const start = Number.parseInt(args[2] ?? panic('start'), 10);
const end = Number.parseInt(args[3] ?? panic('end'), 10);
const kind = args[4] ?? panic('kind');
let question = args[5] ?? panic('question');
const count = Number.parseInt(args[6] ?? panic('count'), 10);
if(!Number.isInteger(count) || count < 2) {
    panic('invalid count');
}
const program = tsgoProgram(args[0] ?? panic('config'), [path]);
if(question.startsWith('property-info:')) {
    const shape = types(tsgoInspect(program, path, start, end, kind, 'raw-shape'), 'raw-shape');
    question = `property-info\n${shape.root().id}\n${question.slice(14)}`;
}
let units = 0;
for(let index = 0; index < count; index++) {
    units += tsgoInspect(program, path, start, end, kind, question).length;
}
console.log(units.toString());
tsgoRelease(program);
