import { panic, programArguments, readTextFile, tsgoProgram, tsgoRelease } from 'adamic';
import { written } from '../../../typescript/parser/nodes.ts';
import { Parser } from '../../../typescript/parser/parser.ts';
import { Scanner } from '../../../typescript/scanner/scanner.ts';
import { UnaryMinus, byteOffsets } from '../unary_minus.ts';
import { Rules } from '../rules.ts';

const args = programArguments();
const config = args[0] ?? panic('usage: main.ts tsconfig manifest [--count]');
const manifest = readTextFile(args[1] ?? panic('missing manifest'));
if(manifest.kind === 'Error') {
    panic(manifest.message);
}
const paths = manifest.text.split('\n').filter((path) => path !== '');
const program = tsgoProgram(config, paths);
// The full root set stays unchanged; --files chooses only the checks to run.
const filesFlag = args.indexOf('--files');
let runPaths = paths;
if(filesFlag >= 0) {
    const selected = readTextFile(args[filesFlag + 1] ?? panic('missing selected manifest'));
    if(selected.kind === 'Error') { panic(selected.message); }
    runPaths = selected.text.split('\n').filter((path) => path !== '');
    for(const path of runPaths) {
        if(!paths.includes(path)) { panic('selected file outside program roots'); }
    }
}
let findings = 0;
for(const path of runPaths) {
    const source = readTextFile(path);
    if(source.kind === 'Error') {
        panic(source.message);
    }
    const parser = new Parser(source.text, path);
    const scanner = new Scanner(source.text);
    const offsets = byteOffsets(source.text);
    const unary = new UnaryMinus(program, path, parser, scanner, offsets);
    const linter = new Rules(program, path, parser, scanner, offsets, unary);
    linter.run();
    findings += linter.findings.length;
    if(!args.includes('--count')) {
        console.log(`file\t${written(path)}`);
        for(const finding of linter.findings) {
            console.log(finding.written());
        }
    }
}
console.log(`findings ${findings}`);
tsgoRelease(program);
