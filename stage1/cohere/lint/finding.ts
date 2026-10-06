import type { RepairEdit } from './repair_edit.ts';
import type { RepairSuggestion } from './repair_suggestion.ts';
export class Finding {
    readonly extraFixes: RepairEdit[] = [];
    readonly suggestions: RepairSuggestion[] = [];
    editStart: number;
    editEnd: number;
    readonly rule: string;
    readonly id: string;
    readonly message: string;
    readonly start: number;
    readonly end: number;
    readonly repair: string;
    readonly replacement: string;
    readonly suggestion: string;
    constructor(
        rule: string,
        id: string,
        message: string,
        start: number,
        end: number,
        repair: string,
        replacement: string,
        suggestion: string,
    ) {
        this.editStart = start;
        this.editEnd = end;
        this.rule = rule;
        this.id = id;
        this.message = message;
        this.start = start;
        this.end = end;
        this.repair = repair;
        this.replacement = replacement;
        this.suggestion = suggestion;
    }
}
