import { referenceError } from './sourceReferences.ts';
import { syntaxError } from './sourceValidation.ts';
import { externalModule } from './sourceModules.ts';
import { panic } from 'adamic';
import { Parser } from './sourceParser.ts';
import { Converter } from './convert.ts';
import { collectComments } from './comments.ts';
import { Postprocessor, mergeComments } from './postprocess.ts';
import { dump, written } from './protocol.ts';
export function answer(path: string, text: string): string {
    if(text.includes('\ufffd')) {
        return panic('ESTree cannot recover original UTF-8 bytes from replacement-decoded input');
    }
    const file = path.toLowerCase();
    const babel = file.endsWith('.js') || file.endsWith('.mjs') || file.endsWith('.cjs') || file.endsWith('.jsx');
    if(
        !babel &&
        !file.endsWith('.ts') &&
        !file.endsWith('.mts') &&
        !file.endsWith('.cts') &&
        !file.endsWith('.a') &&
        !file.endsWith('.tsx')
    ) {
        return panic('unrecognized ESTree source extension');
    }
    const convertedText = text.startsWith('#!') ? `//${text.slice(2)}` : text;
    let parser = new Parser(convertedText, path);
    parser.jsx = babel || file.endsWith('.tsx');
    let root = parser.file();
    if(!['.d.ts', '.d.mts', '.d.cts'].some((suffix) => file.endsWith(suffix)) && externalModule(parser.nodes, root)) {
        parser = new Parser(convertedText, path);
        parser.jsx = babel || file.endsWith('.tsx');
        parser.awaitContext = true;
        root = parser.file();
    }
    if(parser.diagnostics.length > 0) {
        const diagnostic = parser.diagnostics[0] ?? panic('missing diagnostic');
        return panic(`ESTree parser diagnostic ${diagnostic.code} at ${diagnostic.start}: ${diagnostic.message}`);
    }
    if(parser.scanner.errors.length > 0) {
        return panic('ESTree scanner diagnostic');
    }
    const syntax = syntaxError(parser.nodes);
    if(syntax !== '') {
        return panic(`ESTree parser: ${syntax}`);
    }
    const converter = new Converter(parser, convertedText);
    converter.babel = babel;
    for(let id = 0; id < parser.nodes.length; id++) {
        if(parser.nodes[id]?.kind === 'Decorator' && (converter.parents[id] ?? -1) < 0) {
            return panic('ESTree unattached decorator');
        }
    }
    const comments = mergeComments(converter.arena, collectComments(converter, root));
    const reference = referenceError(convertedText, converter.arena, comments);
    if(reference !== '') {
        return panic(`ESTree parser: ${reference}`);
    }
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
