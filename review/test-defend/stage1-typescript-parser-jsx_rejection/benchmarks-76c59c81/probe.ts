import { readTextFile, programArguments } from 'adamic';
import { Parser } from '/workspace/adamic/stage1/typescript/parser/parser.ts';
for(const path of programArguments()) {
    const read = readTextFile(path);
    if(read.kind === 'Error') { throw new Error(read.message); }
    const parser = new Parser(read.text, path);
    parser.file();
    console.log(JSON.stringify({path, roots: parser.roots.length, interfaces: parser.nodes.filter(node => node.kind === 'InterfaceDeclaration').length}));
}
