import { utf8Length } from 'adamic';

export interface Location {
    readonly line: number;
    readonly column: number;
    readonly byte: number;
}

export class Locations {
    readonly source: string;
    readonly starts: number[] = [0];
    constructor(source: string) {
        this.source = source;
        for (let index = 0; index < source.length; index++) {
            const code = source.charCodeAt(index);
            if (code === 13) {
                if (source.charCodeAt(index + 1) === 10) { index++; }
                this.starts.push(index + 1);
            } else if (code === 10 || code === 8232 || code === 8233) { this.starts.push(index + 1); }
        }
    }
    at(position: number): Location {
        let low = 0;
        let high = this.starts.length;
        while (low + 1 < high) {
            const middle = Math.floor((low + high) / 2);
            if ((this.starts[middle] ?? 0) <= position) { low = middle; } else { high = middle; }
        }
        return { line: low + 1, column: utf8Length(this.source.slice(this.starts[low] ?? 0, position)) + 1, byte: utf8Length(this.source.slice(0, position)) };
    }
}
