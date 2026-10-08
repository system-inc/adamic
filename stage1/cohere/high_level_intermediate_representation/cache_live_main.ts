// Native rule entry witness: construct from resident compiler facts, never replayed graph data.
import { HIRFile, ForFunction } from './cache.ts';
import { panic, programArguments, readTextFile } from 'adamic';
import { Parser } from '../../typescript/parser/parser.ts';
import { Checker, openProgram, releaseProgram } from '../lint/checker.a';
import { readSymbols } from './symbol_live.ts';
import { dump } from './dump.ts';
function read(path: string): string { const value = readTextFile(path); if(value.kind === 'Error') { panic(value.message); } return value.text; }
const args = programArguments(); const rows = read(args[1] ?? panic('missing cache manifest')).split('\n').filter((line) => line !== '').map((line) => line.split('\t'));
const paths: string[] = []; for(const row of rows) { const path = row[1] ?? panic('missing source'); if(!paths.includes(path)) { paths.push(path); } }
const program = openProgram(args[0] ?? panic('missing compiler config'), paths); if(program.kind === 'Error') { panic(program.message); }
const files = new Map<string, HIRFile>();
for(const row of rows) {
    const path = row[1] ?? panic('missing source'); const source = read(path); let file = files.get(path);
    if(file === undefined) {
        const parser = new Parser(source, path); const root = parser.file();
        const checker = new Checker(program.value, path, parser, source, undefined, false); checker.root = root;
        file = new HIRFile(source, readSymbols(checker), path); files.set(path, file);
    }
    const index = file.at(Number.parseInt(row[3] ?? '', 10), Number.parseInt(row[4] ?? '', 10));
    const first = ForFunction(file, index) ?? panic('live cache fixture declined'); const before = dump(first);
    const second = ForFunction(file, index); const third = ForFunction(file, index);
    const unchecked = new HIRFile(source, undefined, path);
    console.log(`case\t${row[0]}`);
    console.log(`identity ${first === second && second === third ? 1 : 0} unchecked ${ForFunction(unchecked, index) === undefined ? 1 : 0}`);
    console.log(dump(first).trimEnd());
    if(before !== dump(first)) { console.log('construction changed on cache hit'); }
}
const released = releaseProgram(program.value); if(released.kind === 'Error') { panic(released.message); }
