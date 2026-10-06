export class Source {
    startLine: number | undefined;
    startColumn: number | undefined;
    endLine: number | undefined;
    endColumn: number | undefined;
    constructor(
        startLine: number | undefined,
        startColumn: number | undefined,
        endLine: number | undefined,
        endColumn: number | undefined,
    ) {
        this.startLine = startLine;
        this.startColumn = startColumn;
        this.endLine = endLine;
        this.endColumn = endColumn;
    }
}
