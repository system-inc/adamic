import { utf8Length } from 'adamic';

// Immutable for one input. Interior byte positions retain the partial-code-point
// replacement semantics of Go's byte slices when rendered through JSON.
export class ByteView {
    readonly text: string;
    readonly length: number;
    readonly ascii: boolean;
    private readonly characters: number[] = [];
    private readonly partial: number[] = [];
    constructor(text: string) {
        this.text = text;
        this.length = utf8Length(text);
        this.ascii = this.length === text.length;
        if(!this.ascii) {
            let index = 0;
            for(const character of text) {
                const length = utf8Length(character);
                for(let byte = 0; byte < length; byte++) {
                    this.characters.push(index);
                    this.partial.push(byte);
                }
                index += character.length;
            }
            this.characters.push(index);
            this.partial.push(0);
        }
    }
    clamp(offset: number): number {
        return offset < 0 ? Math.max(0, this.length + offset) : Math.min(offset, this.length);
    }
    prefixLength(offset: number): number {
        const byte = this.clamp(offset);
        return this.ascii ? byte : (this.characters[byte] ?? 0) + (this.partial[byte] ?? 0);
    }
    slice(start: number, end: number = this.length): string {
        start = this.clamp(start);
        end = this.clamp(end);
        if(end <= start) return '';
        if(this.ascii) return this.text.slice(start, end);
        let leading = '';
        if((this.partial[start] ?? 0) > 0) {
            let next = start;
            while(next < end && (this.partial[next] ?? 0) > 0) next++;
            leading = replacement(next - start);
            start = next;
            if(start === end) return leading;
        }
        const trailing = replacement(this.partial[end] ?? 0);
        return leading + this.text.slice(this.characters[start] ?? 0, this.characters[end] ?? 0) + trailing;
    }
}

function replacement(count: number): string {
    let result = '';
    for(let i = 0; i < count; i++) result += '�';
    return result;
}
