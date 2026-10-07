import { panic, programArguments, readTextFile } from 'adamic';
import { whitespaceFrame } from '../whitespaceCodec.ts';
import { printWhitespace } from '../whitespace.ts';
import { DocumentArena } from '../document.ts';
const args = programArguments();
const input = readTextFile(args[0] ?? panic('usage: whitespace_probe.ts <native fixture stream>'));
if(input.kind === 'Error') panic(input.message);
const output: string[] = [];
for(const line of input.text.split('\n')) {
    if(!line.startsWith('S\t')) continue;
    const frame = whitespaceFrame(line.split('\t'));
    const observed: string[] = [];
    for(const mode of ['preserve', 'always', 'never']) {
        for(let link = 0; link < 2; link++) {
            const arena = new DocumentArena();
            const id = printWhitespace(arena, {
                value: frame.value,
                proseWrap: mode,
                link: link === 1,
                previous: frame.previous,
                next: frame.next,
                afterNext: frame.afterNext,
                ancestorKinds: frame.ancestorKinds,
                ancestorSetext: frame.ancestorSetext,
                samples: frame.samples,
            });
            const node = arena.node(id);
            let classification = '';
            if(node.kind === 't') classification = `T${node.text}`;
            else if(node.kind === 'a') classification = 'H';
            else if(node.kind === 'h') classification = (node.flag & 1) !== 0 ? 'S' : 'L';
            else panic('unknown whitespace document');
            observed.push(classification);
        }
    }
    output.push(observed.join(','));
}
console.log(output.join('\n'));
