import { panic, programArguments, readTextFile, tsgoProgram, tsgoRelease } from 'adamic';
import { written } from '../../typescript/parser/nodes.ts';
import { Parser } from '../../typescript/parser/parser.ts';
import { Scanner } from '../../typescript/scanner/scanner.ts';
import { UnaryMinus, byteOffsets } from './unary_minus.ts';
import { Rules } from './rules.ts';
import { Reassign } from './reassign.ts';
import { Bindings } from './bindings.ts';
import { Before } from './before.ts';
import { Coercion } from './coercion.ts';
import { Caller } from './caller.ts';
import { Unused } from './unused.ts';
import { Flow } from './flow.ts';
import { PropertyAlias } from './property_alias.ts';
import { Casts } from './casts.ts';
import { Methods } from './methods.ts';

const args = programArguments();
const config = args[0] ?? panic('usage: coverage_suite.ts tsconfig manifest [--count]');
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
    const root = parser.file();
    for(let at = 0; at < parser.nodes.length; at++) {
        linter.parents.push(-1);
    }
    linter.links(root, -1);
    new Reassign(linter).run();
    new Before(new Bindings(linter)).run();
    new Coercion(linter).run();
    new Methods(linter).run();
    new Casts(linter).run();
    new PropertyAlias(new Bindings(linter)).run();
    new Flow(new Bindings(linter)).run();
    new Unused(new Bindings(linter)).run();
    new Caller(new Bindings(linter)).run();
    for(const finding of linter.findings) {
        finding.sortKey = finding.written();
    }
    linter.findings.sort((left, right) => left.sortKey < right.sortKey ? -1 : left.sortKey > right.sortKey ? 1 : 0);
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
