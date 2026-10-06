import { panic } from 'adamic';
import { Parser } from '../../typescript/parser/parser.ts';
import { Converter } from './convert.ts';
import { collectComments } from './comments.ts';
import { Postprocessor, mergeComments } from './postprocess.ts';
import { dump, written } from './protocol.ts';
export function answer(path: string, text: string): string {
    if(text.includes('\ufffd')) {
        return panic('ESTree cannot recover original UTF-8 bytes from replacement-decoded input');
    }
    const file = path.toLowerCase();
    if(file.endsWith('.tsx') || file.endsWith('.jsx')) {
        return panic('ESTree requires the unported JSX parser grammar for this extension');
    }
    const babel = file.endsWith('.js') || file.endsWith('.mjs') || file.endsWith('.cjs');
    if(!babel && !file.endsWith('.ts') && !file.endsWith('.mts') && !file.endsWith('.cts') && !file.endsWith('.a')) {
        return panic('unrecognized ESTree source extension');
    }
    const convertedText = text.startsWith('#!') ? `//${text.slice(2)}` : text;
    const parser = new Parser(convertedText, path);
    const root = parser.file();
    if(parser.scanner.errors.length > 0) {
        return panic('ESTree scanner diagnostic');
    }
    const converter = new Converter(parser, convertedText);
    converter.babel = babel;
    const comments = mergeComments(converter.arena, collectComments(converter, root));
    if(babel) {
        for(const id of comments) {
            const comment = converter.arena.node(id);
            if(
                comment.type === 'Block' &&
                comment.string('value').startsWith('*') &&
                /@(?:type|satisfies)\b/.test(comment.string('value'))
            ) {
                converter.typeCastEnds.push(comment.end);
            }
        }
    }
    const program = converter.convert(root);
    const processor = new Postprocessor(converter.arena, text, comments, converter.offsets);
    const result = processor.program(program);
    if(babel) {
        const stack = [result];
        while(stack.length > 0) {
            const id = stack.pop() ?? -1;
            const node = converter.arena.node(id);
            if(node.type.startsWith('TS')) {
                return panic(
                    `${path}: ${node.type} at byte ${node.start} is TypeScript syntax, which the babel parser does not read`,
                );
            }
            const children = converter.arena.children(id);
            for(let index = children.length - 1; index >= 0; index--) {
                stack.push(children[index] ?? -1);
            }
        }
    }
    const parts = [dump(converter.arena, result)];
    for(const comment of comments) {
        parts.push('comment\n');
        parts.push(dump(converter.arena, comment));
    }
    parts.push(`stripped ${written(processor.stripped.text())}\n`);
    return parts.join('');
}
