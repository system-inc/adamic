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
    linter.run();
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
    shadow.run();
    const assignment = new Assignment(linter, volume, shadow);
    assignment.run();
    const voidRule = new VoidRule(linter, volume);
    voidRule.run();
    const returns = new Returns(linter, shadow);
    returns.run();
    const unbound = new Unbound(linter, shadow);
    unbound.run();
    volume.run();
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
