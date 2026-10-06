import type { Repair } from './repair.ts';
import type { Suggestion } from './suggestion.ts';
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
    readonly repairs: Repair[] = [];
    readonly suggestions: Suggestion[] = [];
    constructor(rule: string, id: string, message: string, start: number, end: number, namespace = '@typescript-eslint/') {
        this.rule = `${namespace}${rule}`;
        this.id = id;
        this.message = message;
        this.start = start;
        this.end = end;
    }
    written(): string {
        let line = `${this.start}\t${this.end}\t${this.rule}\t${this.id}\t${written(this.message)}\t${this.repairs.length + (this.fixStart < 0 ? 0 : 1)}\t${this.suggestions.length}${this.fixStart < 0 ? '' : `\t${this.fixStart}\t${this.fixEnd}\t${written(this.replacement)}`}`;
        for(const fix of this.repairs) {
            line += `\t${fix.start}\t${fix.end}\t${written(fix.text)}`;
        }
        for(const suggestion of this.suggestions) {
            line += `\t${suggestion.id}\t${written(suggestion.message)}\t${suggestion.repairs.length}`;
            for(const fix of suggestion.repairs) {
                line += `\t${fix.start}\t${fix.end}\t${written(fix.text)}`;
            }
        }
        return line;
    }
}
