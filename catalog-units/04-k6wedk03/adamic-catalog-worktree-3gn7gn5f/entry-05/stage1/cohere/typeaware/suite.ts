import { panic, programArguments, readTextFile, tsgoProgram, tsgoRelease } from 'adamic';
import { written } from '../../typescript/parser/nodes.ts';
import { Parser } from '../../typescript/parser/parser.ts';
import { Scanner } from '../../typescript/scanner/scanner.ts';
import { UnaryMinus, byteOffsets } from './unary_minus.ts';
import { Rules } from './rules.ts';

const args = programArguments();
const config = args[0] ?? panic('usage: main.ts tsconfig manifest [--count]');
const manifest = readTextFile(args[1] ?? panic('missing manifest'));
if(manifest.kind === 'Error') {
    panic(manifest.message);
}
const paths = manifest.text.split('\n').filter((path) => path !== '');
const program = tsgoProgram(config, paths);
let findings = 0;
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
