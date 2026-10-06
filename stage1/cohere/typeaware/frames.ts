import { panic } from 'adamic';

export class Frames {
    readonly text: string;
    cursor = 0;
    constructor(text: string) {
        this.text = text;
    }
    field(): string {
        const newline = this.text.indexOf('\n', this.cursor);
        if(newline < this.cursor) {
            panic('invalid checker facts frame');
        }
        const digits = this.text.slice(this.cursor, newline);
        const length = Number.parseInt(digits, 10);
        if(
            !Number.isInteger(length) ||
            length < 0 ||
            length.toString() !== digits ||
            newline + 1 + length > this.text.length
        ) {
            panic('invalid checker facts length');
        }
        this.cursor = newline + 1 + length;
        return this.text.slice(newline + 1, this.cursor);
    }
    number(): number {
        const digits = this.field();
        const value = Number.parseInt(digits, 10);
        if(!Number.isSafeInteger(value) || value.toString() !== digits) {
            panic('invalid checker facts integer');
        }
        return value;
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
