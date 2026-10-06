import { panic } from 'adamic';
import type { Parser } from '../../typescript/parser/parser.ts';
import type { Scanner } from '../../typescript/scanner/scanner.ts';
import type { ParseNode } from '../../typescript/parser/nodes.ts';
import { Finding } from './finding.ts';
import { RepairEdit } from './repair_edit.ts';
import type { RepairSuggestion } from './repair_suggestion.ts';
import type { Settings } from './settings.ts';

// Borrow the source and syntax indexes; never own the linter importing the rules.
export class Batch5Context {
    readonly source: string;
    readonly parser: Parser;
    readonly scanner: Scanner;
    readonly parents: number[];
    readonly findings: Finding[];
    readonly selected: string;
    readonly settings: Settings;
    constructor(
        source: string,
        parser: Parser,
        scanner: Scanner,
        parents: number[],
        findings: Finding[],
        selected: string,
        settings: Settings,
    ) {
        this.source = source;
        this.parser = parser;
        this.scanner = scanner;
        this.parents = parents;
        this.findings = findings;
        this.selected = selected;
        this.settings = settings;
    }
    node(index: number): ParseNode {
        return this.parser.node(index);
    }
    start(index: number): number {
        this.scanner.pos = this.node(index).pos;
        this.scanner.scan();
        return this.scanner.start;
    }
    text(index: number): string {
        return this.source.slice(this.start(index), this.node(index).end);
    }
    ancestry(index: number, owner: number): void {
        this.parents[index] = owner;
        for(const child of this.node(index).children) {
            this.ancestry(child, index);
        }
    }
    parent(index: number): number {
        return this.parents[index] ?? -1;
    }
    enabled(rule: string): boolean {
        return this.selected === 'all' || this.selected === rule;
    }
    hasKind(index: number, kind: string): boolean {
        return this.node(index).children.some((child) => this.node(child).kind === kind);
    }
    unwrap(index: number): number {
        let result = index;
        while(this.node(result).kind === 'ParenthesizedExpression') {
            result = this.node(result).children[0] ?? panic('missing expression');
        }
        return result;
    }
    scanAt(position: number): RepairEdit {
        this.scanner.pos = position;
        this.scanner.scan();
        return new RepairEdit(this.scanner.start, this.scanner.pos, '');
    }
    report(
        index: number,
        rule: string,
        id: string,
        message: string,
        edits: RepairEdit[],
        suggestions: RepairSuggestion[],
        rangeEnd = -1,
    ): Finding {
        const edit = edits[0];
        const suggestion = suggestions[0];
        let first: RepairEdit | undefined;
        if(suggestion !== undefined) {
            first = suggestion.edits[0];
        }
        const finding = new Finding(
            rule,
            id,
            message,
            this.start(index),
            rangeEnd < 0 ? this.node(index).end : rangeEnd,
            edit === undefined ? (suggestion === undefined ? '' : 'suggestion') : 'fix',
            edit?.text ?? first?.text ?? '',
            suggestion?.message ?? '',
        );
        const primary = edit ?? first;
        if(primary !== undefined) {
            finding.editStart = primary.start;
            finding.editEnd = primary.end;
        }
        for(const extra of edits.slice(1)) {
            finding.extraFixes.push(extra);
        }
        for(const proposed of suggestions) {
            finding.suggestions.push(proposed);
        }
        this.findings.push(finding);
        return finding;
    }
}
