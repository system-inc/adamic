export class LayoutCommand {
    document: number;
    indentation: number;
    mode: number;
    offset = 0;
    constructor(document: number, indentation: number, mode: number) {
        this.document = document;
        this.indentation = indentation;
        this.mode = mode;
    }
}
