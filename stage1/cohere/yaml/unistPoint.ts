export class UnistPoint {
    line: number;
    column: number;
    offset: number;
    constructor(line: number, column: number, offset: number) {
        this.line = line;
        this.column = column;
        this.offset = offset;
    }
}
