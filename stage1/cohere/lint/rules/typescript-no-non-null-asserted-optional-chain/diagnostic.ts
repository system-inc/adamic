// Complete diagnostics for this unit's validator. The shared Finding model
// cannot carry multiple suggestion edits; keep those edits explicitly here.
import { panic } from 'adamic';
import type { RuleContext } from '../../context.ts';
export class Edit {
    readonly start: number;
    readonly end: number;
    readonly text: string;
    constructor(start: number, end: number, text: string) {
        this.start = start;
        this.end = end;
        this.text = text;
    }
}
export class Suggestion {
    readonly id: string;
    readonly message: string;
    readonly edits: Edit[];
    constructor(id: string, message: string, edits: Edit[]) {
        this.id = id;
        this.message = message;
        this.edits = edits;
    }
}
export class Diagnostic {
    readonly rule: string;
    readonly id: string;
    readonly message: string;
    readonly start: number;
    readonly end: number;
    readonly suggestions: Suggestion[];
    constructor(rule: string, id: string, message: string, start: number, end: number, suggestions: Suggestion[]) {
        this.rule = rule;
        this.id = id;
        this.message = message;
        this.start = start;
        this.end = end;
        this.suggestions = suggestions;
    }
}
// Bounded to the current file; the driver empties this array before every run.
export const diagnostics: Diagnostic[] = [];
let completeSuggestions = false;
export function enableCompleteSuggestions(): void {
    completeSuggestions = true;
}
export function record(context: RuleContext, index: number, rule: string, id: string, message: string, suggestions: Suggestion[]): void {
    if(suggestions.length > 0 && !completeSuggestions) {
        panic('NotYet: shared Finding model cannot serialize full suggestion edits');
    }
    context.report(index, rule, id, message, '', '', '');
    diagnostics.push(new Diagnostic(rule, id, message, context.start(index), context.node(index).end, suggestions));
}
