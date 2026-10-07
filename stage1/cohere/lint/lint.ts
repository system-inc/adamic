// The lint harness: parse, one preorder walk that hands each node to the generated rule registry, the
// stable finding sort, and the converging fixer. Every rule lives in its own directory under rules/.
import { RuleContext } from './context.ts';
import { createRuleSet, type RuleSet } from './.generated/registry.ts';
import { panic, utf8Length } from 'adamic';
import { Parser } from '../../typescript/parser/parser.ts';
import type { ParseNode } from '../../typescript/parser/nodes.ts';
import type { Scanner } from '../../typescript/scanner/scanner.ts';

import type { Finding } from './finding.ts';
import { Scanner as SourceScanner } from '../../typescript/scanner/scanner.ts';
import type { Settings } from './settings.ts';

// compareFindings orders by position alone. The sort is stable, so findings at one position keep the order
// they were collected in: walk order, and descriptor order within one node's visit.
function compareFindings(left: Finding, right: Finding): number {
    return left.start - right.start;
}
function compareEdits(left: Finding, right: Finding): number {
    const position = left.editStart - right.editStart;
    if(position !== 0) {
        return position;
    }
    const end = left.editEnd - right.editEnd;
    if(end !== 0) {
        return end;
    }
    return left.rule < right.rule ? -1 : left.rule > right.rule ? 1 : 0;
}

export class Linter {
    readonly source: string;
    readonly parser: Parser;
    readonly scanner: Scanner;
    readonly findings: Finding[] = [];
    readonly rejected: string[] = [];
    parents: number[] = [];
    root = -1;
    readonly selected: string;
    readonly mode: string;
    readonly nullPolicy: string;
    readonly settings: Settings;
    readonly allowCatch: boolean;
    constructor(
        source: string,
        parser: Parser,
        scanner: Scanner,
        selected: string,
        mode: string,
        nullPolicy: string,
        allowCatch: boolean,
        settings: Settings,
    ) {
        this.settings = settings;
        this.source = source;
        this.parser = parser;
        this.scanner = scanner;
        this.selected = selected === '' ? 'all' : selected;
        this.mode = mode === '' ? 'Always' : mode;
        this.nullPolicy = this.mode === 'Always' ? (nullPolicy === '' ? 'Always' : nullPolicy) : 'Ignore';
        this.allowCatch = allowCatch;
    }
    run(): void {
        this.root = this.parser.file();
        this.parents = this.parser.nodes.map(() => -1);
        this.ancestry(this.root, -1);
        const context = new RuleContext(
            this.source,
            this.parser,
            this.scanner,
            this.selected,
            this.mode,
            this.nullPolicy,
            this.allowCatch,
            this.parents,
            this.settings,
            this.root,
        );
        const rules = createRuleSet(context);
        rules.prepare(this.root);
        this.walk(this.root, -1, rules);
        rules.finish(this.root);
        for(const finding of context.findings) {
            this.findings.push(finding);
        }
        this.findings.sort(compareFindings);
    }
    ancestry(index: number, parent: number): void {
        this.parents[index] = parent;
        for(const child of this.node(index).children) {
            this.ancestry(child, index);
        }
    }
    node(index: number): ParseNode {
        return this.parser.node(index);
    }
    walk(index: number, parent: number, rules: RuleSet): void {
        rules.visit(index, parent);
        for(const child of this.node(index).children) {
            this.walk(child, index, rules);
        }
    }
    fixed(): string {
        let current = this.source;
        let findings = this.findings;
        for(let pass = 0; pass < 10; pass++) {
            const proposals = findings.filter((finding) => finding.repair === 'fix');
            proposals.sort(compareEdits);
            const applied: Finding[] = [];
            let previous = -1;
            let winner = '';
            for(const finding of proposals) {
                if(
                    finding.editStart < previous ||
                    (finding.editStart === previous && finding.editStart === finding.editEnd)
                ) {
                    this.rejected.push(
                        `rejected ${finding.rule} ${utf8Length(current.slice(0, finding.editStart))} ${utf8Length(current.slice(0, finding.editEnd))} ${winner} overlaps another fix`,
                    );
                    continue;
                }
                if(finding.editStart < 0 || finding.editEnd < finding.editStart || finding.editEnd > current.length) {
                    panic('invalid fix range');
                }
                if(current.slice(finding.editStart, finding.editEnd) === finding.replacement) {
                    panic('nonprogressing fix');
                }
                applied.push(finding);
                previous = finding.editEnd;
                winner = finding.rule;
            }
            if(applied.length === 0) {
                return current;
            }
            let result = current;
            for(let index = applied.length - 1; index >= 0; index--) {
                const finding = applied[index] ?? panic('missing fix');
                result = result.slice(0, finding.editStart) + finding.replacement + result.slice(finding.editEnd);
            }
            const parser = new Parser(result, 'fixed source');
            const scanner = new SourceScanner(result);
            const next = new Linter(
                result,
                parser,
                scanner,
                this.selected,
                this.mode,
                this.nullPolicy,
                this.allowCatch,
                this.settings,
            );
            next.run();
            current = result;
            findings = next.findings;
        }
        panic('fix pass budget exhausted');
    }
}
