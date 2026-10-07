import { panic, programArguments, readTextFile } from 'adamic';
import { Parser } from '../../../../../../typescript/parser/parser.ts';
import { countTree } from '../../../../../../typescript/parser/nodes.ts';

const path = programArguments()[0] ?? panic('missing source');
const source = readTextFile(path);
if(source.kind === 'Error') {
    panic(source.message);
}
const parser = new Parser(source.text, path);
console.log(`${countTree(parser.nodes, parser.file())}`);
