import { panic, programArguments, tsgoProgram, tsgoRelease, tsgoInspect } from 'adamic';
const args = programArguments();
const program = tsgoProgram(args[0] ?? panic('config'), [args[1] ?? panic('file')]);
tsgoRelease(program);
console.log(tsgoInspect(program, args[1] ?? panic('file'), 0, 2, 'PrefixUnaryExpression', 'type'));
