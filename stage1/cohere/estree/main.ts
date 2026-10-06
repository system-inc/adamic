import { panic, programArguments, readTextFile } from 'adamic';
import { Parser } from '../../typescript/parser/parser.ts';
import { Converter } from './convert.ts';
import { collectComments } from './comments.ts';
import { Postprocessor, mergeComments } from './postprocess.ts';
import { dump, written } from './protocol.ts';
export function answer(path: string, text: string): string {
    if(!path.endsWith('.ts') && !path.endsWith('.mts') && !path.endsWith('.cts')) {
        return panic('ESTree source driver currently requires a TypeScript extension');
    }
    const convertedText = text.startsWith('#!') ? `//${text.slice(2)}` : text;
    const parser = new Parser(convertedText, path);
    const root = parser.file();
    if(parser.scanner.errors.length > 0) {
        return panic('ESTree scanner diagnostic');
    }
    const converter = new Converter(parser, convertedText);
    const program = converter.convert(root);
    const comments = mergeComments(converter.arena, collectComments(converter, root));
    const processor = new Postprocessor(converter.arena, text, comments, converter.offsets);
    const result = processor.program(program);
    const parts = [dump(converter.arena, result)];
    for(const comment of comments) {
        parts.push('comment\n');
        parts.push(dump(converter.arena, comment));
    }
    parts.push(`stripped ${written(processor.stripped.text())}\n`);
    return parts.join('');
}
function run(path: string): void {
    const file = readTextFile(path);
    if(file.kind === 'Error') {
        panic(file.message);
    }
    console.log(answer(path, file.text).slice(0, -1));
}
const args = programArguments();
const path = args[0] ?? panic('usage: main.ts <file.ts> | --manifest <file>');
if(path === '--manifest') {
    const manifest = readTextFile(args[1] ?? panic('missing manifest'));
    if(manifest.kind === 'Error') {
        panic(manifest.message);
    }
    for(const item of manifest.text.split('\n')) {
        if(item !== '') {
            run(item);
        }
    }
}
else {
    run(path);
}
