// Calls the unchanged production rule through Checker.ask/askFile; no rule predicate here.
import { panic, programArguments, readTextFile, writeTextFile } from 'adamic';
import { Parser } from '../../../typescript/parser/parser.ts';
import { Scanner } from '../../../typescript/scanner/scanner.ts';
import { Checker, FileQuestion, openProgram, releaseProgram } from '../checker.a';
import { RuleContext } from '../context.ts';
import { Settings } from '../settings.ts';
import { PilotPlanner, validatePlan, planMutant } from './planning.ts';
import { Rule } from '../rules/no-unnecessary-boolean-literal-compare/rule.a';
function read(path: string): string {
    const value = readTextFile(path);
    if(value.kind === 'Error') { panic(value.message); }
    return value.text;
}
function frame(value: string): string { return `${value.length}\n${value}`; }
function ancestry(parser: Parser, parents: number[], index: number): void {
    for(const child of parser.node(index).children) { parents[child] = index; ancestry(parser, parents, child); }
}
function walk(rule: Rule, parser: Parser, index: number): void {
    const node = parser.node(index);
    if(node.kind === 'BinaryExpression') { rule.visit(node, index); }
    for(const child of node.children) { walk(rule, parser, child); }
}
const args = programArguments();
const config = args[0] ?? panic('missing config');
const sourceArgument = args[1] ?? panic('missing source');
const isManifest = sourceArgument === '--manifest';
const pathList = isManifest ? read(args[2] ?? panic('missing manifest')).trim().split('\n') : [sourceArgument];
const repetitions = parseInt(args[isManifest ? 3 : 2] ?? '1', 10);
const recordPath = args[isManifest ? 4 : 3] ?? '';
const forbiddenRead = isManifest ? '' : args[4] ?? '';
const repairMutant = isManifest ? '' : args[5] ?? '';
if(!Number.isSafeInteger(repetitions) || repetitions < 1) { panic('invalid repetitions'); }
const opened = openProgram(config, []);
if(opened.kind === 'Error') { panic(opened.message); }
const program = opened.value;
for(const path of pathList) {
    if(isManifest) { console.log(`file ${path}`); }
const source = read(path);
const parser = new Parser(source, path);
const root = parser.file();
const parents: number[] = [];
for(let index = 0; index < parser.nodes.length; index++) { parents.push(-1); }
ancestry(parser, parents, root);
for(let pass = 0; pass < repetitions; pass++) {
    const checker = new Checker(program, path, parser, source, undefined, pass === 0);
    checker.root = root;
    checker.enter('@typescript-eslint/no-unnecessary-boolean-literal-compare', ['ReadsCompilerOptions']);
    if(forbiddenRead !== '') {
        const refusal = checker.askFile(new FileQuestion(forbiddenRead, 'options'));
        releaseProgram(program);
        panic(refusal.reason);
    }
    const context = new RuleContext(source, parser, new Scanner(source), '@typescript-eslint/no-unnecessary-boolean-literal-compare', '', '', false, parents, new Settings(), root, checker);
    const plan = planMutant(new PilotPlanner(context).plan(root), isManifest ? '' : args[6] ?? '');
    const rule = new Rule(context);
    rule.prepare(root);
    walk(rule, parser, root);
    checker.finish();
    if(checker.refusals.length !== 0) { panic('pilot checker refusal'); }
    if(pass === 0) { validatePlan(plan, checker); }
    if(pass === 0) {
        if(recordPath !== '') {
            const saved = writeTextFile(isManifest ? `${recordPath}.${pathList.indexOf(path)}` : recordPath, checker.transcript());
            if(saved.kind === 'Error') { panic(saved.message); }
        }
        console.log(`findings ${context.findings.length}`);
        for(const finding of context.findings) {
            if(finding.repair !== 'fix' || finding.extraFixes.length !== 0 || finding.suggestions.length !== 0) { panic('unexpected pilot repair shape'); }
            const start = checker.offsets[finding.start] ?? panic('bad finding start');
            const end = checker.offsets[finding.end] ?? panic('bad finding end');
            const editStart = checker.offsets[finding.editStart] ?? panic('bad edit start');
            const editEnd = checker.offsets[finding.editEnd] ?? panic('bad edit end');
            console.log(frame(`${start}`) + frame(`${end}`) + frame(finding.rule) + frame(finding.id) + frame(finding.message) + frame(`${editStart}`) + frame(`${editEnd}`) + frame(finding.replacement + repairMutant));
        }
    }
}
}
const released = releaseProgram(program);
if(released.kind === 'Error') { panic(released.message); }
