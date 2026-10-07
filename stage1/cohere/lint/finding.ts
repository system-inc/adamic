import type { Suggestion, SuggestionEdit } from './suggestions.a';

export class Finding {
    readonly suggestions: Suggestion[] = [];
    // The automatic edits after the first, which rides repair, replacement, editStart and editEnd. Cohere's
    // edit engine proposes each one on its own, so fixed() expands them into proposals of their own.
    readonly extraFixes: SuggestionEdit[] = [];
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
