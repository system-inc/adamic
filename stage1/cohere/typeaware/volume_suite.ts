import { panic, programArguments, readTextFile, tsgoProgram, tsgoRelease } from 'adamic';
import { written } from '../../typescript/parser/nodes.ts';
import { Parser } from '../../typescript/parser/parser.ts';
import { Scanner } from '../../typescript/scanner/scanner.ts';
import { UnaryMinus, byteOffsets } from './unary_minus.ts';
import { Frames, header } from './frames.ts';
import { Rules } from './rules.ts';
import { Returns } from './returns.ts';
import { Unbound } from './unbound.ts';
import { VoidRule } from './void.ts';
import { Assignment } from './assignment.ts';
import { Shadow } from './shadow.ts';
import { Volume } from './volume.ts';

const args = programArguments();
const config = args[0] ?? panic('usage: volume_suite.ts tsconfig manifest [--count]');
const manifest = readTextFile(args[1] ?? panic('missing manifest'));
if(manifest.kind === 'Error') {
    panic(manifest.message);
}
const paths = manifest.text.split('\n').filter((path) => path !== '');
// --rules names a manifest of rules, as findings name them: only the walks that own one run, and only their findings
// print, so one file's work splits by rule family (volume_agreement_corpus_shards_test.go). Without it every walk runs.
const rulesAt = args.indexOf('--rules');
let selected: string[] = [];
if(rulesAt >= 0) {
    const listed = readTextFile(args[rulesAt + 1] ?? panic('missing rules manifest'));
    if(listed.kind === 'Error') {
        panic(listed.message);
    }
    selected = listed.text.split('\n').filter((rule) => rule !== '');
}
const rulesWalk = [
    '@typescript-eslint/no-unsafe-unary-minus',
    '@typescript-eslint/related-getter-setter-pairs',
    '@typescript-eslint/no-unsafe-declaration-merging',
    '@typescript-eslint/no-unsafe-argument',
    '@typescript-eslint/restrict-plus-operands',
    '@typescript-eslint/no-unnecessary-boolean-literal-compare',
];
const volumeWalk = [
    '@typescript-eslint/prefer-nullish-coalescing',
    '@typescript-eslint/switch-exhaustiveness-check',
    '@typescript-eslint/no-unsafe-enum-comparison',
    '@typescript-eslint/no-unsafe-type-assertion',
    '@typescript-eslint/no-unsafe-member-access',
];
const shadowWalk = ['@typescript-eslint/no-shadow'];
const assignmentWalk = ['@typescript-eslint/no-unsafe-assignment'];
const voidWalk = ['@typescript-eslint/no-confusing-void-expression'];
const returnsWalk = ['@typescript-eslint/consistent-return'];
const unboundWalk = ['@typescript-eslint/unbound-method'];
const walks = [rulesWalk, volumeWalk, shadowWalk, assignmentWalk, voidWalk, returnsWalk, unboundWalk];
for(const rule of selected) {
    if(!walks.some((walk) => walk.includes(rule))) {
        panic(`unknown rule in --rules: ${rule}`);
    }
}
function runs(walk: string[]): boolean {
    return rulesAt < 0 || walk.some((rule) => selected.includes(rule));
}
const program = tsgoProgram(config, paths);
let findings = 0;
let optionsChecked = false;
for(const path of paths) {
    const source = readTextFile(path);
    if(source.kind === 'Error') {
        panic(source.message);
    }
    const parser = new Parser(source.text, path);
    const scanner = new Scanner(source.text);
    const offsets = byteOffsets(source.text);
    const unary = new UnaryMinus(program, path, parser, scanner, offsets);
    const linter = new Rules(program, path, parser, scanner, offsets, unary);
    linter.run(runs(rulesWalk));
    if(!optionsChecked) {
        const options = new Frames(linter.ask(parser.nodes.length - 1, 'strict-this'));
        header(options, 'strict-this');
        const strictThis = options.yes();
        options.end();
        if(!strictThis) {
            panic('volume suite requires noImplicitThis; implicit-this messages are not yet ported');
        }
        optionsChecked = true;
    }
    const volume = new Volume(linter);
    const shadow = new Shadow(linter);
    if(runs(shadowWalk)) {
        shadow.run();
    }
    const assignment = new Assignment(linter, volume, shadow);
    if(runs(assignmentWalk)) {
        assignment.run();
    }
    const voidRule = new VoidRule(linter, volume);
    if(runs(voidWalk)) {
        voidRule.run();
    }
    const returns = new Returns(linter, shadow);
    if(runs(returnsWalk)) {
        returns.run();
    }
    const unbound = new Unbound(linter, shadow);
    if(runs(unboundWalk)) {
        unbound.run();
    }
    volume.run(runs(volumeWalk));
    const printed = linter.findings.filter((finding) => rulesAt < 0 || selected.includes(finding.rule));
    findings += printed.length;
    if(!args.includes('--count')) {
        console.log(`file\t${written(path)}`);
        for(const finding of printed) {
            console.log(finding.written());
        }
    }
}
console.log(`findings ${findings}`);
tsgoRelease(program);
