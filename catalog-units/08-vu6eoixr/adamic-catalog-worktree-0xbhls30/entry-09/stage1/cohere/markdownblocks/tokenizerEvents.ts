// micromark/create-tokenizer: numeric token identities and owned event frames.
import { panic } from 'adamic';
import type { InputChunkInterface } from './inputChunks.ts';
import {
    tokenPoint,
    copyTokenPoint,
    tokenEvent,
    type TokenArena,
    type TokenPointInterface,
    type TokenEventInterface,
    type MarkdownTokenInterface,
} from './tokenArena.ts';
export interface TokenStoreInterface {
    readonly from: number;
    readonly point: TokenPointInterface;
    readonly previous: number;
    readonly construct: number;
    readonly stack: readonly number[];
}
export class TokenizerEvents {
    readonly arena: TokenArena;
    readonly context: number;
    readonly chunks: readonly InputChunkInterface[];
    readonly events: TokenEventInterface[] = [];
    readonly stack: number[] = [];
    readonly columns = new Map<number, number>();
    point = tokenPoint(1, 1, 0, 0, -1);
    previous = 0;
    construct = -1;
    expected = 0;
    consumed = true;
    constructor(arena: TokenArena, context: number, chunks: readonly InputChunkInterface[], from: TokenPointInterface) {
        this.arena = arena;
        this.context = context;
        this.chunks = chunks;
        this.point.line = from.line === 0 ? 1 : from.line;
        this.point.column = from.column === 0 ? 1 : from.column;
        this.point.offset = from.offset;
    }
    now(): TokenPointInterface {
        return copyTokenPoint(this.point);
    }
    prepare(): number {
        const chunk = this.chunks[this.point.index] ?? panic('micromark input exhausted');
        if(chunk.isText && this.point.bufferIndex < 0) this.point.bufferIndex = 0;
        this.expected = chunk.isText
            ? (chunk.text[this.point.bufferIndex] ?? panic('micromark buffer exhausted'))
            : chunk.code;
        this.consumed = false;
        return this.expected;
    }
    consume(code: number): void {
        if(code !== this.expected) panic('micromark: expected given code to equal expected code');
        if(code === -5 || code === -4 || code === -3) {
            this.point.line++;
            this.point.column = 1;
            this.point.offset += code === -3 ? 2 : 1;
            this.accountSkip();
        }
        else if(code !== -1) {
            this.point.column++;
            this.point.offset++;
        }
        if(this.point.bufferIndex < 0) this.point.index++;
        else {
            this.point.bufferIndex++;
            if(
                this.point.bufferIndex === (this.chunks[this.point.index] ?? panic('micromark text chunk')).text.length
            ) {
                this.point.bufferIndex = -1;
                this.point.index++;
            }
        }
        this.previous = code;
        this.consumed = true;
    }
    enter(type: string, fields: number): number {
        const id = fields < 0 ? this.arena.add(type, this.point) : fields;
        this.arena.token(id).type = type;
        this.arena.token(id).start = this.now();
        this.events.push(tokenEvent(true, id, this.context));
        this.stack.push(id);
        return id;
    }
    exit(type: string): number {
        const id = this.stack.pop() ?? panic('micromark: cannot close w/o open tokens');
        const token = this.arena.token(id);
        token.end = this.now();
        if(type !== token.type) panic('micromark: expected exit token to match current token');
        this.events.push(tokenEvent(false, id, this.context));
        return id;
    }
    defineSkip(point: TokenPointInterface): void {
        this.columns.set(point.line, point.column);
        this.accountSkip();
    }
    accountSkip(): void {
        const column = this.columns.get(this.point.line);
        if(column !== undefined && this.point.column < 2) {
            this.point.column = column;
            this.point.offset += column - 1;
        }
    }
    store(): TokenStoreInterface {
        return {
            from: this.events.length,
            point: this.now(),
            previous: this.previous,
            construct: this.construct,
            stack: this.stack.slice(),
        };
    }
    restore(info: TokenStoreInterface): void {
        this.point = copyTokenPoint(info.point);
        this.previous = info.previous;
        this.construct = info.construct;
        while(this.events.length > info.from) this.events.pop();
        while(this.stack.length > 0) this.stack.pop();
        for(const id of info.stack) this.stack.push(id);
        this.accountSkip();
    }
}
export function sliceTokenChunks(
    chunks: readonly InputChunkInterface[],
    token: MarkdownTokenInterface,
): InputChunkInterface[] {
    if(token.start.index === token.end.index) {
        const chunk = chunks[token.start.index] ?? panic('token slice index');
        if(!chunk.isText || token.start.bufferIndex < 0 || token.end.bufferIndex < 0)
            panic('token slice is not a text buffer');
        return [{ text: chunk.text.slice(token.start.bufferIndex, token.end.bufferIndex), code: 0, isText: true }];
    }
    let view = chunks.slice(token.start.index, token.end.index);
    if(token.start.bufferIndex > -1) {
        const head = view[0] ?? panic('token slice head');
        if(head.isText) view[0] = { text: head.text.slice(token.start.bufferIndex), code: 0, isText: true };
        else view = view.slice(1);
    }
    if(token.end.bufferIndex > 0) {
        const tail = chunks[token.end.index] ?? panic('token slice tail');
        view.push({ text: tail.text.slice(0, token.end.bufferIndex), code: 0, isText: true });
    }
    return view;
}
export function serializeTokenChunks(chunks: readonly InputChunkInterface[], expandTabs: boolean): string {
    const units: number[] = [];
    let atTab = false;
    for(const chunk of chunks) {
        if(chunk.isText) {
            for(const unit of chunk.text) units.push(unit);
        }
        else if(chunk.code === -5) units.push(13);
        else if(chunk.code === -4) units.push(10);
        else if(chunk.code === -3) {
            units.push(13);
            units.push(10);
        }
        else if(chunk.code === -2) units.push(expandTabs ? 32 : 9);
        else if(chunk.code === -1) {
            if(!expandTabs && atTab) continue;
            units.push(32);
        }
        else units.push(chunk.code & 65535);
        atTab = !chunk.isText && chunk.code === -2;
    }
    const output: string[] = [];
    for(let index = 0; index < units.length; index++) {
        const unit = units[index] ?? panic('serialized unit');
        if(unit >= 55296 && unit <= 56319) {
            const next = units[index + 1] ?? -1;
            if(next >= 56320 && next <= 57343) {
                output.push(String.fromCodePoint((unit - 55296) * 1024 + next - 56320 + 65536));
                index++;
            }
            else output.push('�');
        }
        else output.push(unit >= 56320 && unit <= 57343 ? '�' : String.fromCodePoint(unit));
    }
    return output.join('');
}
