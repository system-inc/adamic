import { panic, programArguments, readTextFile } from 'adamic';
import { Parser } from '../../../../typescript/parser/parser.ts';
import { written } from '../../../../typescript/parser/nodes.ts';
import { Scanner } from '../../../../typescript/scanner/scanner.ts';
import { RuleContext } from '../../context.ts';
import { Settings } from '../../settings.ts';
import type { Finding } from '../../finding.ts';
import { create, type Rule } from './rule.ts';
import type { Suggestion } from './repairs.ts';

class Result {
    readonly finding: Finding;
    readonly suggestions: Suggestion[];
    constructor(finding: Finding, suggestions: Suggestion[]) {
        this.finding = finding;
        this.suggestions = suggestions;
    }
}
function walk(context: RuleContext, rule: Rule, results: Result[], index: number): void {
    const before = context.findings.length;
    rule.visit(index, context.parents[index] ?? -1);
    for(let cursor = before; cursor < context.findings.length; cursor++) {
        results.push(new Result(context.findings[cursor] ?? panic('missing finding'), rule.suggestions(index)));
    }
    for(const child of context.node(index).children) { walk(context, rule, results, child); }
}
function run(row: string, countOnly: boolean): number {
    const fields = row.split('\t');
    const path = fields[0] ?? panic('missing path');
    const input = readTextFile(path);
    if(input.kind === 'Error') { panic(input.message); }
    const parser = new Parser(input.text, fields[1] ?? path);
    const root = parser.file();
    const parents: number[] = [];
    for(const node of parser.nodes) { parents.push(-1); }
    for(let index = 0; index < parser.nodes.length; index++) {
        for(const child of parser.node(index).children) { parents[child] = index; }
    }
    const settings = new Settings();
    settings.load(fields[2] ?? '');
    const context = new RuleContext(input.text, parser, new Scanner(input.text), '', '', '', false, parents, settings);
    const rule = create(context);
    const results: Result[] = [];
    walk(context, rule, results, root);
    if(countOnly) { return results.length; }
    results.sort((left, right) => left.finding.start - right.finding.start);
    const offsets: number[] = [0];
    let bytes = 0;
    for(let index = 0; index < input.text.length; index++) {
        const code = input.text.codePointAt(index) ?? 0;
        if(code > 65535) { offsets.push(bytes); index++; }
        bytes += code < 128 ? 1 : code < 2048 ? 2 : code < 65536 ? 3 : 4;
        offsets.push(bytes);
    }
    for(const result of results) {
        const finding = result.finding;
        const start = offsets[finding.start] ?? panic('finding outside source');
        const end = offsets[finding.end] ?? panic('finding outside source');
        console.log(`finding\t${start}\t${end}\t${finding.id}\t${written(finding.message)}`);
        for(const suggestion of result.suggestions) {
            console.log(`suggestion\t${suggestion.id}\t${written(suggestion.message)}`);
            for(const edit of suggestion.edits) {
                console.log(`edit\t${offsets[edit.start] ?? panic('edit outside source')}\t${offsets[edit.end] ?? panic('edit outside source')}\t${written(edit.text)}`);
            }
        }
    }
    console.log(`fixed\t${written(input.text)}`);
    return results.length;
}
const args = programArguments();
const manifest = readTextFile(args[0] ?? panic('missing manifest'));
if(manifest.kind === 'Error') { panic(manifest.message); }
const countOnly = args.includes('--count');
let count = 0;
let caseNumber = 0;
for(const row of manifest.text.split('\n')) {
    if(row !== '') {
        if(!countOnly) { console.log(`case ${caseNumber}`); }
        count += run(row, countOnly);
        caseNumber++;
    }
}
if(countOnly) { console.log(`${count}`); }
