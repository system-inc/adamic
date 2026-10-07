import { panic, programArguments, readTextFile } from 'adamic';
import { encode } from '../codec.ts';
import { inputChunks } from '../inputChunks.ts';
import { TokenizerEvents, serializeTokenChunks, sliceTokenChunks } from '../tokenizerEvents.ts';
import { TokenArena, tokenPoint, type TokenPointInterface } from '../tokenArena.ts';
function pointText(point: TokenPointInterface): string {
    return `${point.line},${point.column},${point.offset},${point.index},${point.bufferIndex}`;
}
const input = readTextFile(programArguments()[0] ?? panic('usage: events_probe.ts <UTF-16 cases>'));
if(input.kind === 'Error') panic(input.message);
const output: string[] = [];
for(const line of input.text.split('\n')) {
    if(line === '') continue;
    const units: number[] = [];
    if(line !== '-') for(const value of line.split(',')) units.push(Number.parseInt(value, 10));
    const chunks = inputChunks(units);
    const arena = new TokenArena();
    const context = new TokenizerEvents(arena, 0, chunks, tokenPoint(1, 1, 0, 0, -1));
    context.defineSkip(tokenPoint(2, 3, 0, 0, 0));
    context.defineSkip(tokenPoint(3, 5, 0, 0, 0));
    context.construct = 7;
    context.enter('document', -1);
    const fields = arena.add('fields', context.now());
    arena.token(fields).spread = true;
    arena.token(fields).align.push('left');
    let data = false;
    while(context.point.index < chunks.length) {
        const code = context.prepare();
        if(code <= 0 && data) {
            context.exit('data');
            data = false;
        }
        if(code < 0) {
            const saved = context.store();
            context.construct = 9;
            context.enter('rollback', -1);
            context.consume(code);
            context.exit('rollback');
            context.restore(saved);
            context.enter('virtual', fields);
            context.consume(code);
            context.exit('virtual');
        }
        else if(code > 0) {
            if(!data) {
                context.enter('data', -1);
                data = true;
            }
            context.consume(code);
        }
        else {
            context.exit('document');
            context.consume(code);
        }
    }
    const rows: string[] = [];
    for(const event of context.events) {
        const token = arena.token(event.token);
        let text = '-';
        let expanded = '-';
        if(token.start.index !== token.end.index || (token.start.bufferIndex >= 0 && token.end.bufferIndex >= 0)) {
            const stream = sliceTokenChunks(chunks, token);
            text = serializeTokenChunks(stream, false);
            expanded = serializeTokenChunks(stream, true);
        }
        rows.push(
            `${event.enter}\t${token.type}\t${pointText(token.start)}\t${pointText(token.end)}\t${token.spread}\t${token.align.join(',')}\t${encode(text)}\t${encode(expanded)}`,
        );
    }
    rows.push(
        `now\t${pointText(context.now())}\t${context.previous}\t${context.construct}\t${context.consumed}\t${context.stack.length}`,
    );
    output.push(encode(rows.join('\n')));
}
console.log(output.join('\n'));
