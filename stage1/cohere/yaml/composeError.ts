// Composer diagnostics use UTF-16 positions.
export class ComposeError {
    start: number;
    end: number;
    code: string;
    message: string;
    constructor(start: number, end: number, code: string, message: string) {
        this.start = start;
        this.end = end;
        this.code = code;
        this.message = message;
    }
}
