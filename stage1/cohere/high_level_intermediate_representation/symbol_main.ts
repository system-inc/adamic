// Compiler-fact witness driver, separate from the construction coverage driver.
import { panic, programArguments, readTextFile } from 'adamic';
import { Parser } from '../../typescript/parser/parser.ts';
import { written } from '../../typescript/parser/nodes.ts';
import { Checker, openProgram, releaseProgram } from '../lint/checker.a';
import { SymbolFacts, readSymbol } from './symbol.ts';
function read(path: string): string {
    const result = readTextFile(path);
    if(result.kind === 'Error') { panic(result.message); }
    return result.text;
}
function printFacts(facts: SymbolFacts): void {
    console.log(`symbol ${facts.identity} ${written(facts.name)} ${facts.declarations.length}`);
    for(const declaration of facts.declarations) {
        console.log(`${declaration.kind} ${declaration.sameSource} ${declaration.start}:${declaration.end} ${written(declaration.name)} ${written(declaration.property)} ${written(declaration.module)}`);
    }
}
const args = programArguments();
const mode = args[0] ?? panic('missing mode');
if(mode === '--replay') {
    for(const path of read(args[1] ?? panic('missing facts manifest')).split('\n')) {
        if(path !== '') { printFacts(new SymbolFacts(read(path))); }
    }
}
else if(mode === '--live') {
    const config = args[1] ?? panic('missing config');
    const file = args[2] ?? panic('missing source');
    const source = read(file);
    const parser = new Parser(source, file); parser.file();
    const program = openProgram(config, [file]);
    if(program.kind === 'Error') { panic(program.message); }
    const checker = new Checker(program.value, file, parser, source, undefined, false);
    for(const line of read(args[3] ?? panic('missing query manifest')).split('\n')) {
        if(line === '') { continue; }
        const span = line.split('\t');
        let selected = -1;
        for(let index = 0; index < parser.nodes.length; index++) {
            const node = parser.node(index);
            if(node.kind === 'Identifier' && `${checker.offsets[node.pos]}` === span[0] && `${checker.offsets[node.end]}` === span[1]) { selected = index; break; }
        }
        if(selected < 0) { panic(`port parser has no exact identifier ${line}`); }
        printFacts(readSymbol(checker, selected));
    }
    const released = releaseProgram(program.value);
    if(released.kind === 'Error') { panic(released.message); }
}
else { panic('unknown symbol witness mode'); }
