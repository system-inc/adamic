// One preorder traversal, with listeners generated from each rule directory.
import { panic } from 'adamic';
import { Parser } from '../../typescript/parser/parser.ts';
import type { Scanner } from '../../typescript/scanner/scanner.ts';
import type { Finding } from './finding.ts';
import { RuleContext } from './context.ts';
import type { Settings } from './settings.ts';
import { RuleSet } from './.generated/registry.ts';

export class Linter {
    readonly context: RuleContext;
    readonly findings: Finding[];
    readonly parents: number[] = [];
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
        this.context = new RuleContext(
            source,
            parser,
            scanner,
            selected,
            mode,
            nullPolicy,
            allowCatch,
            this.parents,
            settings,
        );
        this.findings = this.context.findings;
    }
    run(): void {
        const root = this.context.parser.file();
        this.context.parser.nodes.forEach(() => {
            this.parents.push(-1);
        });
        this.ancestry(root, -1);
        const rules = new RuleSet(this.context);
        rules.prepare(root);
        this.walk(root, -1, rules);
        rules.finish(root);
        this.findings.sort((left, right) => left.start - right.start);
    }
    ancestry(index: number, parent: number): void {
        this.parents[index] = parent;
        for(const child of this.context.node(index).children) {
            this.ancestry(child, index);
        }
    }
    walk(index: number, parent: number, rules: RuleSet): void {
        rules.visit(index, parent);
        for(const child of this.context.node(index).children) {
            this.walk(child, index, rules);
        }
    }
    fixed(): string {
        let result = this.context.source;
        let previous = this.context.source.length;
        for(let index = this.findings.length - 1; index >= 0; index--) {
            const finding = this.findings[index] ?? panic('missing finding');
            if(finding.repair !== 'fix') {
                continue;
            }
            if(finding.end > previous) {
                panic('overlapping fixes');
            }
            result = result.slice(0, finding.start) + finding.replacement + result.slice(finding.end);
            previous = finding.start;
        }
        // Selected rules propose disjoint edits; parse every rewrite before exposing it.
        if(result !== this.context.source) {
            const checked = new Parser(result, 'fixed source');
            checked.file();
        }
        return result;
    }
}
