import { panic, programArguments, tsgoProgram, tsgoRelease, tsgoTypeParts } from 'adamic';
const args = programArguments();
const config = args[0] ?? panic('missing config');
const path = args[1] ?? panic('missing path');
const count = Number.parseInt(args[2] ?? '10000', 10);
if(!Number.isInteger(count) || count < 2) {
    panic('query count must be at least two');
}
const program = tsgoProgram(config, [path]);
let units = 0;
for(let index = 0; index < count; index++) {
    units += tsgoTypeParts(program, path, 0, 2, 'PrefixUnaryExpression').length;
}
console.log(`${units}`);
tsgoRelease(program);
