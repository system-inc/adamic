import { runProcess as execute, programArguments, panic } from 'adamic';
import { quoted } from '../config/json.ts';
const arguments_ = programArguments();
const result = execute('/bin/sh', ['-c', arguments_[0] ?? panic('missing script')], arguments_[1] ?? '', [
    'ADAMIC_PARENT_VALUE',
    'ADAMIC_CHILD_VALUE=one',
    'ADAMIC_CHILD_VALUE=two',
]);
console.log(`${quoted(result.output)}\t${result.exitCode}\t${result.signal}\t${quoted(result.error)}`);
