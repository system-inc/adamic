import { panic } from 'adamic';

function decimal(text: string, first: number, last: number, message: string): number {
    const negative = text.charCodeAt(first) === 45;
    const start = first + (negative ? 1 : 0);
    if(start >= last || (text.charCodeAt(start) === 48 && (negative || start + 1 < last))) {
        panic(message);
    }
    let value = 0;
    for(let at = start; at < last; at++) {
        const digit = text.charCodeAt(at) - 48;
        if(digit < 0 || digit > 9) {
            panic(message);
        }
        value = value * 10 + digit;
    }
    if(!Number.isSafeInteger(value)) {
        panic(message);
    }
    return negative ? -value : value;
}

export class Frames {
    readonly text: string;
    cursor = 0;
    constructor(text: string) {
        this.text = text;
    }
    next(): number {
        const newline = this.text.indexOf('\n', this.cursor);
        if(newline < this.cursor) {
            panic('invalid checker facts frame');
        }
        const length = decimal(this.text, this.cursor, newline, 'invalid checker facts length');
        if(length < 0 || newline + 1 + length > this.text.length) {
            panic('invalid checker facts length');
        }
        this.cursor = newline + 1 + length;
        return newline + 1;
    }
    field(): string {
        const first = this.next();
        return this.text.slice(first, this.cursor);
    }
    number(): number {
        const first = this.next();
        return decimal(this.text, first, this.cursor, 'invalid checker facts integer');
    }
    natural(): number {
        const value = this.number();
        if(value < 0) {
            panic('negative checker fact');
        }
        return value;
    }
    yes(): boolean {
        const value = this.number();
        if(value !== 0 && value !== 1) {
            panic('invalid checker facts boolean');
        }
        return value === 1;
    }
    ids(): number[] {
        const count = this.natural();
        const ids: number[] = [];
        for(let index = 0; index < count; index++) {
            ids.push(this.natural());
        }
        return ids;
    }
    end(): void {
        if(this.cursor !== this.text.length) {
            panic('trailing checker facts');
        }
    }
}
export function header(frames: Frames, expected: string): void {
    if(frames.number() !== 1 || frames.field() !== expected) {
        panic('unsupported checker facts schema');
    }
}
