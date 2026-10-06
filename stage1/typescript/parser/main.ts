import { panic, programArguments, readTextFile, utf8Length } from 'adamic';
import { Parser } from './parser.ts';
import { countTree, printTree } from './nodes.ts';

function run(path: string, countOnly: boolean, whole: boolean, docTypes: boolean): number {
    const source = readTextFile(path);
    if(source.kind === 'Error') {
        panic(source.message);
    }
    const parser = new Parser(source.text, path);
    const types: number[] = docTypes ? parser.docTypes() : [];
    const file = docTypes ? -1 : parser.file();
    const offsets: number[] = [0];
    if(!countOnly) {
        let bytes = 0;
        for(let index = 0; index < source.text.length; index++) {
            const code = source.text.codePointAt(index) ?? 0;
            if(code > 65535) {
                offsets.push(bytes);
                index++;
            }
            bytes += code < 128 ? 1 : code < 2048 ? 2 : code < 65536 ? 3 : 4;
            offsets.push(bytes);
        }
        if(bytes !== utf8Length(source.text)) {
            panic('source byte mapping differs');
        }
    }
    let count = 0;
    for(const root of docTypes ? types : whole ? [file] : parser.roots) {
        if(!countOnly) {
            console.log(docTypes ? 'type' : whole ? 'file' : 'expression');
        }
        count += countOnly
            ? countTree(parser.nodes, root)
            : printTree(parser.nodes, root, offsets, 0, whole || docTypes);
    }
    return count;
}
const args = programArguments();
const first = args[0] ?? panic('usage: main.ts <file> or --manifest <file> [--count]');
if(first === '--manifest') {
    const manifest = readTextFile(args[1] ?? panic('missing manifest'));
    if(manifest.kind === 'Error') {
        panic(manifest.message);
    }
    const countOnly = args.includes('--count');
    let count = 0;
    let caseNumber = 0;
    for(const path of manifest.text.split('\n')) {
        if(path === '') {
            continue;
        }
        if(!countOnly) {
            console.log(`case ${caseNumber}`);
        }
        count += run(path, countOnly, args.includes('--whole'), args.includes('--doc-types'));
        caseNumber++;
    }
    if(countOnly) {
        console.log(`${count}`);
    }
}
else {
    run(first, false, args.includes('--whole'), args.includes('--doc-types'));
}
