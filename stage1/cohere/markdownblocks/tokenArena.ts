// Numeric micromark token storage and leaf point/event records.
import { panic } from 'adamic';
export interface TokenPointInterface {
    line: number;
    column: number;
    offset: number;
    index: number;
    bufferIndex: number;
}
export function tokenPoint(
    line: number,
    column: number,
    offset: number,
    index: number,
    bufferIndex: number,
): TokenPointInterface {
    return { line, column, offset, index, bufferIndex };
}
export function copyTokenPoint(point: TokenPointInterface): TokenPointInterface {
    return tokenPoint(point.line, point.column, point.offset, point.index, point.bufferIndex);
}
export interface MarkdownTokenInterface {
    type: string;
    start: TokenPointInterface;
    end: TokenPointInterface;
    spread: boolean;
    align: string[];
}
export interface TokenEventInterface {
    readonly enter: boolean;
    readonly token: number;
    readonly context: number;
}
export function tokenEvent(enter: boolean, token: number, context: number): TokenEventInterface {
    return { enter, token, context };
}
export class TokenArena {
    readonly tokens: MarkdownTokenInterface[] = [];
    token(id: number): MarkdownTokenInterface {
        return this.tokens[id] ?? panic('missing micromark token');
    }
    add(type: string, start: TokenPointInterface): number {
        const id = this.tokens.length;
        this.tokens.push({
            type,
            start: copyTokenPoint(start),
            end: tokenPoint(0, 0, 0, 0, 0),
            spread: false,
            align: [],
        });
        return id;
    }
}
