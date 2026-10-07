// The dependency is recorded raw checker data; all decisions execute from .a source.
export * from '../../../../../oracle/adamic.mjs';
import {panic} from '../../../../../oracle/adamic.mjs';
import {readFileSync} from 'node:fs';
const fixtures = JSON.parse(readFileSync(process.env.ADAMIC_REACT_FACTS, 'utf8'));
let live = false;
export function tsgoProgram(config, files) {live = true;return 1;}
export function tsgoInspect(program, file, start, end, kind, question) {
    if(program !== 1 || !live) {panic('invalid or released checker handle');}
    if(kind !== 'SourceFile' || start !== 0 || end !== readFileSync(file).length) {panic('unexpected checker fixture coordinates');}
    return fixtures[file + '\0' + question] ?? panic('missing raw checker fixture');
}
export function tsgoRelease(program) {if(program !== 1 || !live) {panic('invalid or released checker handle');}live = false;}
