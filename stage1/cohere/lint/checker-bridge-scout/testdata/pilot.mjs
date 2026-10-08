// Node exercises the unchanged production port's Checker and rule, not the ABI client.
import { PilotPlanner, validatePlan, planMutant } from '../planning.ts';
import { readFileSync } from 'node:fs';
import { panic } from 'adamic';
import { Parser } from '../../../../typescript/parser/parser.ts';
import { Scanner } from '../../../../typescript/scanner/scanner.ts';
import { Checker, FileQuestion } from '../../checker.a';
import { RuleContext } from '../../context.ts';
import { Settings } from '../../settings.ts';
import { Rule } from '../../rules/no-unnecessary-boolean-literal-compare/rule.a';
const path = process.argv[2];
const text = readFileSync(path, 'utf8');
const parser = new Parser(text, path);
const root = parser.file();
const parents = Array(parser.nodes.length).fill(-1);
function ancestry(index) {
    for(const child of parser.node(index).children) { parents[child] = index; ancestry(child); }
}
ancestry(root);
const checker = new Checker(0, path, parser, text, readFileSync(process.argv[3], 'utf8'), true);
checker.root = root;
checker.enter('@typescript-eslint/no-unnecessary-boolean-literal-compare', ['ReadsCompilerOptions']);
if(process.argv[4]) { panic(checker.askFile(new FileQuestion(process.argv[4], 'options')).reason); }
const context = new RuleContext(text, parser, new Scanner(text), '@typescript-eslint/no-unnecessary-boolean-literal-compare', '', '', false, parents, new Settings(), root, checker);
const plan = planMutant(new PilotPlanner(context).plan(root), process.argv[6] ?? '');
const rule = new Rule(context);
rule.prepare(root);
function walk(index) {
    const node = parser.node(index);
    if(node.kind === 'BinaryExpression') { rule.visit(node, index); }
    for(const child of node.children) { walk(child); }
}
walk(root);
checker.finish();
if(checker.refusals.length !== 0) { throw new Error(JSON.stringify(checker.refusals)); }
validatePlan(plan, checker);
if(readFileSync(process.argv[3], 'utf8') !== checker.transcript()) { throw new Error('ask plan differs from port'); }
function byte(position) { return Buffer.byteLength(text.slice(0, position)); }
function frame(value) { return `${value.length}\n${value}`; }
console.log(`findings ${context.findings.length}`);
for(const f of context.findings) {
    if(f.repair !== 'fix' || f.extraFixes.length !== 0 || f.suggestions.length !== 0) { throw new Error('unexpected pilot repair shape'); }
    console.log([String(byte(f.start)),String(byte(f.end)),f.rule,f.id,f.message,String(byte(f.editStart)),String(byte(f.editEnd)),f.replacement + (process.argv[5] ?? '')].map(frame).join(""));
}
