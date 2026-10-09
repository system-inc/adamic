// PostCSS 8.5.16 Input, with cohere's UTF-16 to byte boundary convention.
import { panic, utf8Length } from 'adamic';
import { quote } from '../selector/nodes.ts';

export class Position {
    offset: number;
    readonly line: number;
    readonly column: number;
    constructor(offset: number, line: number, column: number) {
        this.offset = offset;
        this.line = line;
        this.column = column;
    }
}
export class Input {
    readonly css: string;
    readonly byteOffsets: number[] = [];
    readonly lines: number[] = [0];
    constructor(text: string, stripBOM: boolean = true) {
        this.css = stripBOM && (text[0] === '\ufeff' || text[0] === '\ufffe') ? text.slice(1) : text;
        let bytes = 0;
        for(const character of this.css) {
            this.byteOffsets.push(bytes);
            bytes += utf8Length(character);
            if(character.length === 2) {
                this.byteOffsets.push(bytes);
            }
        }
        this.byteOffsets.push(bytes);
        for(let index = 0; index < this.css.length; index++) {
            if(this.css.charCodeAt(index) === 10) {
                this.lines.push(index + 1);
            }
        }
    }
    byteOffset(offset: number): number {
        if(offset < 0) {
            return offset;
        }
        if(offset > this.css.length) {
            return utf8Length(this.css) + offset - this.css.length;
        }
        return this.byteOffsets[offset] ?? panic('missing CSS byte offset');
    }
    slice(start: number, end: number): string {
        start = Math.min(Math.max(start, 0), this.css.length);
        end = Math.min(Math.max(end, 0), this.css.length);
        if(end <= start) {
            return '';
        }
        if(this.css.charCodeAt(start) >= 0xdc00 && this.css.charCodeAt(start) <= 0xdfff && this.css.charCodeAt(start - 1) >= 0xd800 && this.css.charCodeAt(start - 1) <= 0xdbff) {
            start++;
        }
        if(this.css.charCodeAt(end) >= 0xdc00 && this.css.charCodeAt(end) <= 0xdfff && this.css.charCodeAt(end - 1) >= 0xd800 && this.css.charCodeAt(end - 1) <= 0xdbff) {
            end++;
        }
        return this.css.slice(start, end);
    }
    position(offset: number): Position {
        let low = 0;
        let high = this.lines.length - 1;
        while(low < high) {
            const middle = Math.floor((low + high + 1) / 2);
            if(offset < (this.lines[middle] ?? panic('missing line'))) {
                high = middle - 1;
            }
            else {
                low = middle;
            }
        }
        return new Position(offset, low + 1, offset - (this.lines[low] ?? panic('missing line')) + 1);
    }
    error(reason: string, offset: number, end: number = -1): string {
        const start = this.position(offset);
        const fields: string[] = [`"column":${start.column}`];
        if(end >= 0) {
            const finish = this.position(end);
            fields.push(`"endColumn":${finish.column}`);
            fields.push(`"endLine":${finish.line}`);
            fields.push(`"endOffset":${this.byteOffset(end)}`);
        }
        fields.push(`"line":${start.line}`);
        fields.push('"name":"CssSyntaxError"');
        fields.push(`"offset":${this.byteOffset(offset)}`);
        fields.push(`"reason":${quote(reason)}`);
        return `{${fields.join(',')}}`;
    }
}
