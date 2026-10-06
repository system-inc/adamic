import { written } from '../../typescript/parser/nodes.ts';
export class Diagnostic {
    readonly rule: string;
    readonly id: string;
    readonly message: string;
    readonly start: number;
    readonly end: number;
    fixStart = -1;
    fixEnd = -1;
    replacement = '';
    sortKey = '';
    constructor(rule: string, id: string, message: string, start: number, end: number) {
        this.rule = `@typescript-eslint/${rule}`;
        this.id = id;
        this.message = message;
        this.start = start;
        this.end = end;
    }
    written(): string {
        return `${this.start}\t${this.end}\t${this.rule}\t${this.id}\t${written(this.message)}\t${this.fixStart < 0 ? 0 : 1}\t0${this.fixStart < 0 ? '' : `\t${this.fixStart}\t${this.fixEnd}\t${written(this.replacement)}`}`;
    }
}
