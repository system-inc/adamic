// Comment blanking preserves Go byte offsets even for non-ASCII comment text.
import { panic, utf8Length } from 'adamic';
import type { Arena } from './arena.ts';
export function unitOffsets(text: string): number[] {
    const offsets = [0];
    let bytes = 0;
    for(let index = 0; index < text.length; index++) {
        const code = text.codePointAt(index) ?? panic('missing code point');
        if(code > 65535) {
            offsets.push(bytes);
            index++;
        }
        bytes += code < 128 ? 1 : code < 2048 ? 2 : code < 65536 ? 3 : 4;
        offsets.push(bytes);
    }
    if(bytes !== utf8Length(text)) {
        panic('ESTree byte mapping differs');
    }
    return offsets;
}
export function byteUnit(offsets: readonly number[], byte: number): number {
    let left = 0;
    let right = offsets.length;
    while(left < right) {
        const middle = Math.floor((left + right) / 2);
        if((offsets[middle] ?? -1) < byte) {
            left = middle + 1;
        }
        else {
            right = middle;
        }
    }
    if(offsets[left] === byte) {
        return left;
    }
    return panic('ESTree position inside UTF-8 character');
}
export class StrippedText {
    readonly source: string;
    readonly arena: Arena;
    readonly comments: readonly number[];
    readonly offsets: readonly number[];
    made = false;
    stripped = '';
    strippedOffsets: number[] = [];
    constructor(source: string, arena: Arena, comments: readonly number[], offsets: readonly number[]) {
        this.source = source;
        this.arena = arena;
        this.comments = comments;
        this.offsets = offsets;
    }
    text(): string {
        if(!this.made) {
            const parts: string[] = [];
            let previous = 0;
            for(const id of this.comments) {
                const comment = this.arena.node(id);
                const start = byteUnit(this.offsets, comment.start);
                const end = byteUnit(this.offsets, comment.end);
                parts.push(this.source.slice(previous, start));
                for(const line of this.source.slice(start, end).split('\n')) {
                    parts.push(' '.repeat(utf8Length(line)));
                    parts.push('\n');
                }
                parts.pop();
                previous = end;
            }
            parts.push(this.source.slice(previous));
            this.stripped = parts.join('');
            this.strippedOffsets = unitOffsets(this.stripped);
            this.made = true;
        }
        return this.stripped;
    }
}
// The pinned Go lastWhitespaceSize returns after its one-byte attempt. A final UTF-8
// continuation byte therefore stops trimming, including NBSP and U+2028.
export function trimEndGo(text: string): string {
    let end = text.length;
    while(end > 0) {
        const unit = text.charCodeAt(end - 1);
        if(unit !== 32 && (unit < 9 || unit > 13)) {
            break;
        }
        end--;
    }
    return text.slice(0, end);
}
