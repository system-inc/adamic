// Scout transport only. It exports the unchanged parser, not a flat-tree fallback.
import { panic, programArguments, readTextFile } from 'adamic';
import { Parser } from '../parser/parser.ts';
import { written, countTree } from '../parser/nodes.ts';

const args = programArguments();
const manifest = readTextFile(args[0] ?? panic('missing manifest'));
if(manifest.kind === 'Error') panic(manifest.message);
const repeats = Number(args[1] ?? '0');
let total = 0;
let caseNumber = 0;
for(const path of manifest.text.split('\n')) {
    if(path === '') continue;
    const source = readTextFile(path);
    if(source.kind === 'Error') panic(source.message);
    const parser = new Parser(source.text, path);
    const root = parser.file();
    if(args.includes('--walk')) {
        for(let pass = 0; pass < repeats; pass++) total += countTree(parser.nodes, root);
    }
    else {
        console.log(`case ${caseNumber} ${root} ${parser.roots.join(',')}`);
        for(const node of parser.nodes) {
            console.log(`${node.kind}\t${node.pos}\t${node.end}\t${node.optional ? 32 : 0}\t${node.literalFlags}\t${node.list}\t${node.trailing ? 1 : 0}\t${node.multiLine ? 1 : 0}\t${node.operator}\t${written(node.text)}\t${written(node.raw)}\t${node.semantic}\t${node.children.join(',')}`);
        }
    }
    caseNumber++;
}
if(args.includes('--walk')) console.log(`${total}`);
