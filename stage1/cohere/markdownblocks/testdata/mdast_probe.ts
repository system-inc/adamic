import { panic, programArguments, readTextFile } from 'adamic';
import { encode, decode } from '../codec.ts';
import { MdastCompiler } from '../mdastCompile.ts';
import { TokenSource } from '../tokenSource.ts';
import { parseFrontMatter } from '../parseFrontMatter.ts';
import type { InputChunkInterface } from '../inputChunks.ts';
import {
    TokenArena,
    tokenPoint,
    tokenEvent,
    type TokenPointInterface,
    type TokenEventInterface,
} from '../tokenArena.ts';
function number(value: string | undefined): number {
    return Number.parseInt(value ?? panic('event number'), 10);
}
function point(value: string | undefined): TokenPointInterface {
    const fields = (value ?? panic('event point')).split(',');
    return tokenPoint(number(fields[0]), number(fields[1]), number(fields[2]), number(fields[3]), number(fields[4]));
}
const args = programArguments();
const input = readTextFile(args[0] ?? panic('usage: mdast_probe.ts <events>'));
if(input.kind === 'Error') panic(input.message);
let tokens = new TokenArena();
let sources: readonly TokenSource[] = [];
let events: TokenEventInterface[] = [];
const output: string[] = [];
let source = '';
for(const line of input.text.split('\n')) {
    const fields = line.split('\t');
    if(fields[0] === 'S') source = decode(fields[1] ?? panic('mdast source'));
    else if(fields[0] === 'C') {
        const chunks: InputChunkInterface[] = [];
        for(const field of (fields[1] ?? panic('chunks')).split(';')) {
            if(field === '') continue;
            const text: number[] = [];
            if(field.startsWith('T') && field.length > 1)
                for(const unit of field.slice(1).split(',')) text.push(number(unit));
            chunks.push({
                text,
                code: field.startsWith('C') ? number(field.slice(1)) : 0,
                isText: field.startsWith('T'),
            });
        }
        sources = [...sources, new TokenSource(chunks)];
    }
    else if(fields[0] === 'T') {
        const id = tokens.add(fields[1] ?? panic('token type'), point(fields[2]));
        tokens.token(id).end = point(fields[3]);
        tokens.token(id).spread = fields[4] === 'true';
        if(fields[5] !== '') tokens.token(id).align = (fields[5] ?? panic('align')).split(',');
    }
    else if(fields[0] === 'E') events.push(tokenEvent(fields[1] === 'true', number(fields[2]), number(fields[3])));
    else if(fields[0] === 'F') {
        const arena = new MdastCompiler(tokens, sources, events).compile();
        const front = parseFrontMatter(source);
        arena.attachFrontMatter(front);
        output.push(encode(`content\t${encode(front.content)}\n${arena.canonical(0)}`));
        tokens = new TokenArena();
        sources = [];
        events = [];
    }
    else if(line !== '') panic('event protocol');
}
console.log(output.join('\n'));
