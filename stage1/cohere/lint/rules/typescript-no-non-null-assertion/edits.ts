import { panic } from 'adamic';

// Suggestions may edit several ranges outside the diagnostic. Cache byte
// offsets once per file so reporting many assertions does not rescan prefixes.
export class Edits {
    readonly source: string;
    readonly offsets: number[] = [0];
    constructor(source: string) { this.source = source; }
    range(start: number, end: number, text: string): string {
        if(this.offsets.length === 1) {
            let bytes = 0;
            for(let position = 0; position < this.source.length; position++) {
                const code = this.source.codePointAt(position) ?? 0;
                if(code > 65535) { this.offsets.push(bytes); position++; }
                bytes += code < 128 ? 1 : code < 2048 ? 2 : code < 65536 ? 3 : 4;
                this.offsets.push(bytes);
            }
        }
        const first = this.offsets[start] ?? panic('suggestion starts outside source');
        const last = this.offsets[end] ?? panic('suggestion ends outside source');
        return `${first}:${last}:${text}`;
    }
}
