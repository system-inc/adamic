import { panic } from 'adamic';
import type { Parser } from '../../../typescript/parser/parser.ts';
import type { Scanner } from '../../../typescript/scanner/scanner.ts';
import type { ParseNode } from '../../../typescript/parser/nodes.ts';
import type { Finding } from '../finding.ts';
import type { Settings } from '../settings.ts';
import { RepairEdit } from '../repair_edit.ts';
import type { CommentRange } from './comment_ranges.ts';

// Rule modules borrow this acyclic context, never the linter that imports them.
export class RuleContext {
    readonly source: string;
    readonly parser: Parser;
    readonly scanner: Scanner;
    readonly parents: number[];
    readonly findings: Finding[];
    readonly selected: string;
    readonly settings: Settings;
    comments: CommentRange[] = [];
    readonly importedNames = new Set<string>();
    readonly declaredNames = new Set<string>();
    bindingsReady = false;
    argsBound = false;
    argsNonRest = false;
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
    recordBinding(text: string, imported: boolean, rest: boolean): void {
        if(imported) {
            this.importedNames.add(text);
        }
        else {
            this.declaredNames.add(text);
        }
        if(text === 'args') {
            this.argsBound = true;
            if(!rest) {
                this.argsNonRest = true;
            }
        }
    }
    completeBindings(): void {
        for(const text of this.declaredNames) {
            this.importedNames.delete(text);
        }
        this.bindingsReady = true;
    }
    record(finding: Finding): void {
        this.findings.push(finding);
    }
    setParent(index: number, owner: number): void {
        this.parents[index] = owner;
    }
    scanAt(position: number): RepairEdit {
        this.scanner.pos = position;
        this.scanner.scan();
        return new RepairEdit(this.scanner.start, this.scanner.pos, '');
    }
    node(index: number): ParseNode {
        return this.parser.node(index);
    }
    enabled(rule: string): boolean {
        return this.selected === 'all' || this.selected === rule;
    }
    start(index: number): number {
        this.scanner.pos = this.node(index).pos;
        this.scanner.scan();
        return this.scanner.start;
    }
    hasKind(index: number, kind: string): boolean {
        for(const child of this.node(index).children) {
            if(this.node(child).kind === kind) {
                return true;
            }
        }
        return false;
    }
    unwrap(index: number): number {
        let cursor = index;
        while(this.node(cursor).kind === 'ParenthesizedExpression') {
            cursor = this.node(cursor).children[0] ?? panic('parenthesis without operand');
        }
        return cursor;
    }
    text(index: number): string {
        return this.source.slice(this.start(index), this.node(index).end);
    }
}
