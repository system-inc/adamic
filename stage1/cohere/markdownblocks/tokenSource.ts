// One owned, immutable tokenizer chunk stream.
import type { InputChunkInterface } from './inputChunks.ts';
export class TokenSource {
    readonly chunks: readonly InputChunkInterface[];
    constructor(chunks: readonly InputChunkInterface[]) {
        this.chunks = chunks;
    }
}
