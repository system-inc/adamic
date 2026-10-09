// The lint harness: parse, one preorder walk that hands each node to the generated rule registry, the
// stable finding sort, and the converging fixer. Every rule lives in its own directory under rules/.
import { RuleContext } from './context.ts';
import { createRuleSet, type RuleSet } from './.generated/registry.ts';
import { panic, utf8Length } from 'adamic';
import { Parser } from '../../typescript/parser/parser.ts';
import type { ParseNode } from '../../typescript/parser/nodes.ts';
import type { Scanner } from '../../typescript/scanner/scanner.ts';

import { Finding } from './finding.ts';
import { Scanner as SourceScanner } from '../../typescript/scanner/scanner.ts';
import type { Settings } from './settings.ts';

// passBudget and passBudgetMessage are cohere's edit engine's DefaultMaxPasses and ReasonPassesReached.
const passBudget = 10;
const passBudgetMessage = 'the pass budget was exhausted before the file converged';

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
    // The rules still proposing fixes on the pass that exhausted the budget, in the order they first
    // propose, as cohere's edit engine names them. Empty when the fixes converge.
    readonly unconverged: string[] = [];
    // junkRows appends a copy of every row to the node table after parsing, attached to nothing. Stage 1
    // reads the table only by following links from the root, and the flat copy of typescript-go's tree
    // (#k4fm1vf) relies on it: its tables hold rows no link reaches. TestNodeTableIsLinkOnly sets this
    // and requires the same findings, so a rule that walks the table by row fails there first.
    junkRows = false;
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
        if(this.junkRows) {
            const attached = this.parser.nodes.length;
            for(let index = 0; index < attached; index++) {
                this.parser.nodes.push(this.parser.node(index));
            }
        }
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
        let lastProposals: Finding[] = [];
        for(let pass = 0; pass < passBudget; pass++) {
            const proposals: Finding[] = [];
            for(const finding of findings) {
                if(finding.repair === 'fix') {
                    proposals.push(finding);
                    for(const extra of finding.extraFixes) {
                        const fix = new Finding(
                            finding.rule,
                            finding.id,
                            finding.message,
                            finding.start,
                            finding.end,
                            'fix',
                            extra.text,
                            '',
                        );
                        fix.editStart = extra.start;
                        fix.editEnd = extra.end;
                        proposals.push(fix);
                    }
                }
            }
            lastProposals = proposals;
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
            // The applied fixes are sorted and disjoint, so the new text is built in one pass from the pieces
            // between them. Splicing each fix into the whole text instead copied the file once per fix, which
            // on checker.ts (3.1 MB, 708 fixes) was most of the run's instructions.
            const pieces: string[] = [];
            let copied = 0;
            for(const finding of applied) {
                pieces.push(current.slice(copied, finding.editStart));
                pieces.push(finding.replacement);
                copied = finding.editEnd;
            }
            pieces.push(current.slice(copied));
            const result = pieces.join('');
            // The parser takes its JSX and JavaScript modes from the path, so each pass reparses under the
            // file's own path: a .tsx file's fixed source is still TSX.
            const parser = new Parser(result, this.parser.path);
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
            next.junkRows = this.junkRows;
            next.run();
            current = result;
            findings = next.findings;
        }
        // The budget ran out with fixes still landing. Cohere's edit engine discards the whole run and
        // leaves the file as it was found, because a fixpoint it did not reach is not a result it may
        // write, and it names the rules still proposing on the last pass, since a rule whose fix does not
        // silence its own finding is the defect, not the file.
        this.rejected.push(`rejected fix-engine 0 0  ${passBudgetMessage} (${passBudget} passes)`);
        for(const finding of lastProposals) {
            if(!this.unconverged.includes(finding.rule)) {
                this.unconverged.push(finding.rule);
            }
        }
        return this.source;
    }
}
