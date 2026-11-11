import { panic, programArguments, tsgoProgram, tsgoInspect, tsgoRelease } from 'adamic';
const args = programArguments();
const config = args[0] ?? panic('config');
const file = args[1] ?? panic('file');
const program = tsgoProgram(config, [file]);
console.log(tsgoInspect(program, file, 0, 2, 'PrefixUnaryExpression', 'raw-shape'));
tsgoRelease(program);
console.log(tsgoInspect(program, file, 0, 2, 'PrefixUnaryExpression', 'name\n1'));
