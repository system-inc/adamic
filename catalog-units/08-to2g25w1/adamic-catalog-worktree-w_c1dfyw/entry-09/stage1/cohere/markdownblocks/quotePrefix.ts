// Cohere micromark/block_quote.go start and continuation recognition.
// The caller owns token/event emission; this primitive reports prefix consumption.
import { panic } from 'adamic';
import type { ChunkInterface } from './preprocess.ts';
export interface QuotePrefixInterface {
    readonly matched: boolean;
    readonly consumed: number;
    readonly offset: number;
    readonly column: number;
    readonly open: boolean;
}
function markdownSpace(code: number): boolean {
    return code === 32 || code === -2 || code === -1;
}
export function quotePrefix(
    chunks: readonly ChunkInterface[],
    continuation: boolean,
    open: boolean,
    codeIndentedDisabled: boolean,
): QuotePrefixInterface {
    const codes: number[] = [];
    // Only the first line can be consumed by a quote prefix. Preserve tab virtual codes.
    for(const chunk of chunks) {
        if(chunk.isText) {
            for(let index = 0; index < chunk.text.length; index++) codes.push(chunk.text.charCodeAt(index));
        }
        else {
            codes.push(chunk.code);
            if(chunk.code === 0 || chunk.code < -2) break;
        }
    }
    let index = 0;
    if(continuation) {
        while(markdownSpace(codes[index] ?? panic('missing prefix code'))) {
            if(!codeIndentedDisabled && index >= 3) break;
            index++;
        }
    }
    if(codes[index] !== 62) return { matched: false, consumed: 0, offset: 0, column: 1, open };
    index++;
    if(markdownSpace(codes[index] ?? panic('missing code after quote'))) index++;
    let offset = 0;
    for(let current = 0; current < index; current++) {
        if(codes[current] !== -1) offset++;
    }
    return { matched: true, consumed: index, offset, column: offset + 1, open: true };
}
