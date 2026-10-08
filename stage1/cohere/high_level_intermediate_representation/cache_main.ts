import { panic, programArguments, readTextFile } from 'adamic';
import { HIRFile, ForFunction } from './cache.ts';
import { dump } from './dump.ts';
import { SymbolSnapshot } from './symbol.ts';
function read(path: string): string { const value = readTextFile(path); if(value.kind === 'Error') { panic(value.message); } return value.text; }
const files = new Map<string, HIRFile>();
for(const line of read(programArguments()[0] ?? panic('missing cache manifest')).split('\n')) {
    if(line === '') { continue; }
    const fields = line.split('\t'); const sourcePath = fields[1] ?? panic('missing source'); const source = read(sourcePath);
    let file = files.get(sourcePath);
    if(file === undefined) { file = new HIRFile(source, new SymbolSnapshot(read(fields[2] ?? panic('missing symbols')))); files.set(sourcePath, file); }
    const index = file.at(Number.parseInt(fields[3] ?? '', 10), Number.parseInt(fields[4] ?? '', 10));
    const first = ForFunction(file, index) ?? panic('cache fixture declined');
    const before = dump(first);
    const second = ForFunction(file, index); const third = ForFunction(file, index);
    const unchecked = new HIRFile(source, undefined);
    console.log(`case\t${fields[0]}`);
    console.log(`identity ${first === second && second === third ? 1 : 0} unchecked ${ForFunction(unchecked, index) === undefined ? 1 : 0}`);
    console.log(dump(first).trimEnd());
    if(before !== dump(first)) { console.log('construction changed on cache hit'); }
}
