import { panic, programArguments, readTextFile } from 'adamic';
import { AstArena } from '../astArena.ts';
import { AstPreprocessor } from '../astPreprocess.ts';
import { astInteger, readAstNode, serializeAst } from '../astProtocol.ts';
import { decode, encode } from '../codec.ts';
const args = programArguments();
const input = readTextFile(args[0] ?? panic('usage: ast_probe.ts <unprocessed AST fixtures>'));
if(input.kind === 'Error') panic(input.message);
let arena = new AstArena();
let source = '';
let tabWidth = 4;
const output: string[] = [];
for(const line of input.text.split('\n')) {
    const fields = line.split('\t');
    const kind = fields[0] ?? '';
    if(kind === 'A') {
        arena = new AstArena();
        source = decode(fields[1] ?? '');
        tabWidth = astInteger(fields[2] ?? '');
    }
    else if(kind === 'N') readAstNode(arena, fields);
    else if(kind === 'F') {
        const processor = new AstPreprocessor(arena, source, tabWidth);
        const root = processor.run(0);
        output.push(encode(processor.error === '' ? serializeAst(arena, root) : `E\t${processor.error}`));
    }
    else if(line !== '') panic('unknown AST fixture');
}
console.log(output.join('\n'));
