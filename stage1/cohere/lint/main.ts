import { panic, programArguments, readTextFile } from 'adamic';
import { written } from '../../typescript/parser/nodes.ts';
import { Parser } from '../../typescript/parser/parser.ts';
import { Scanner } from '../../typescript/scanner/scanner.ts';
import { Linter } from './lint.ts';
import { Settings } from './settings.ts';

function run(row: string, countOnly: boolean): number {
    const fields = row.split('\t');
    const path = fields[0] ?? panic('missing path');
    const source = readTextFile(path);
    if(source.kind === 'Error') {
        panic(source.message);
    }
    const parser = new Parser(source.text, path);
    const scanner = new Scanner(source.text);
    const settings = new Settings();
    settings.load(fields[5] ?? '');
    const linter = new Linter(
        source.text,
        parser,
        scanner,
        fields[1] ?? 'all',
        fields[2] ?? '',
        fields[3] ?? '',
        fields[4] === 'true',
        settings,
    );
    linter.run();
    if(countOnly) {
        return linter.findings.length;
    }
    const offsets: number[] = [0];
    const lines: number[] = [0];
    let bytes = 0;
    for(let index = 0; index < source.text.length; index++) {
        const code = source.text.codePointAt(index) ?? 0;
        if(code === 10) {
            lines.push(index + 1);
        }
        if(code > 65535) {
            offsets.push(bytes);
            index++;
        }
        bytes += code < 128 ? 1 : code < 2048 ? 2 : code < 65536 ? 3 : 4;
        offsets.push(bytes);
    }
    for(const finding of linter.findings) {
        const start = offsets[finding.start] ?? panic('finding outside source');
        const end = offsets[finding.end] ?? panic('finding end outside source');
        let left = 0;
        let right = lines.length;
        while(left < right) {
            const middle = Math.floor((left + right) / 2);
            if((lines[middle] ?? 0) <= finding.start) {
                left = middle + 1;
            }
            else {
                right = middle;
            }
        }
        const line = left;
        const column = start - (offsets[lines[line - 1] ?? 0] ?? 0) + 1;
        console.log(`${path}:${line}:${column}\n  ${finding.rule}  ${finding.message}\n`);
        console.log(
            `range ${start} ${end} ${finding.id} ${finding.repair}\t${written(finding.replacement)}\t${written(finding.suggestion)}\t${offsets[finding.editStart] ?? 0} ${offsets[finding.editEnd] ?? 0}`,
        );
    }
    if(fields[6] === 'recovery') {
        console.log('recovery findings only');
    }
    else {
        const fixed = linter.fixed();
        for(const rejection of linter.rejected) {
            console.log(rejection);
        }
        console.log(`fixed\t${written(fixed)}`);
    }
    return linter.findings.length;
}

const args = programArguments();
const first = args[0] ?? panic('usage: main.ts <file> or --manifest <file> [--count]');
if(first === '--manifest') {
    const manifest = readTextFile(args[1] ?? panic('missing manifest'));
    if(manifest.kind === 'Error') {
        panic(manifest.message);
    }
    const countOnly = args.includes('--count');
    let count = 0;
    let caseNumber = 0;
    for(const row of manifest.text.split('\n')) {
        if(row === '') {
            continue;
        }
        if(!countOnly) {
            console.log(`case ${caseNumber}`);
        }
        count += run(row, countOnly);
        caseNumber++;
    }
    if(countOnly) {
        console.log(`${count}`);
    }
}
else {
    run(first, false);
}
