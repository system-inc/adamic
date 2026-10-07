// Immutable comment value, with UTF-8 byte ranges and columns.
export class Comment {
    readonly start: number;
    readonly end: number;
    readonly isBlock: boolean;
    readonly text: string;
    readonly startLine: number;
    readonly startColumn: number;
    readonly endLine: number;
    constructor(
        start: number,
        end: number,
        isBlock: boolean,
        text: string,
        startLine: number,
        startColumn: number,
        endLine: number,
    ) {
        this.start = start;
        this.end = end;
        this.isBlock = isBlock;
        this.text = text;
        this.startLine = startLine;
        this.startColumn = startColumn;
        this.endLine = endLine;
    }
}
