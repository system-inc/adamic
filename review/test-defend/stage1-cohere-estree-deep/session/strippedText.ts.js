import { panic, utf8Length } from 'adamic';
export function unitOffsets(text) {
    const offsets = [
        0
    ];
    let bytes = 0;
    for(let index = 0; index < text.length; index++){
        const code = text.codePointAt(index) ?? panic('missing code point');
        if (code > 65535) {
            offsets.push(bytes);
            index++;
        }
        bytes += code < 128 ? 1 : code < 2048 ? 2 : code < 65536 ? 3 : 4;
        offsets.push(bytes);
    }
    if (bytes !== utf8Length(text)) {
        panic('ESTree byte mapping differs');
    }
    return offsets;
}
export function byteUnit(offsets, byte) {
    let left = 0;
    let right = offsets.length;
    while(left < right){
        const middle = Math.floor((left + right) / 2);
        if ((offsets[middle] ?? -1) < byte) {
            left = middle + 1;
        } else {
            right = middle;
        }
    }
    if (offsets[left] === byte) {
        return left;
    }
    return panic('ESTree position inside UTF-8 character');
}
export class StrippedText {
    source;
    arena;
    comments;
    offsets;
    made = false;
    stripped = '';
    strippedOffsets = [];
    constructor(source, arena, comments, offsets){
        this.source = source;
        this.arena = arena;
        this.comments = comments;
        this.offsets = offsets;
    }
    text() {
        if (!this.made) {
            const parts = [];
            let previous = 0;
            for (const id of this.comments){
                const comment = this.arena.node(id);
                const start = byteUnit(this.offsets, comment.start);
                const end = byteUnit(this.offsets, comment.end);
                parts.push(this.source.slice(previous, start));
                for (const line of this.source.slice(start, end).split('\n')){
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
export function trimEndGo(text) {
    let end = text.length;
    while(end > 0){
        const unit = text.charCodeAt(end - 1);
        if (unit !== 32 && (unit < 9 || unit > 13)) {
            break;
        }
        end--;
    }
    return text.slice(0, end);
}
