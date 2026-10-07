import { panic, programArguments, tsgoProgram, tsgoTypeParts, tsgoRelease } from 'adamic';
const args = programArguments();
const config = args[0] ?? panic('missing config');
const path = args[1] ?? panic('missing path');
const program = tsgoProgram(config, [path]);
tsgoRelease(program);
console.log(tsgoTypeParts(program, path, 0, 2, 'PrefixUnaryExpression'));
