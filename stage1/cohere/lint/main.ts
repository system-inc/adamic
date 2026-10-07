import { panic, programArguments, readTextFile, writeTextFile } from 'adamic';
import { written } from '../../typescript/parser/nodes.ts';
import { Parser } from '../../typescript/parser/parser.ts';
import { Scanner } from '../../typescript/scanner/scanner.ts';
import { Linter } from './lint.ts';
import { Settings } from './settings.ts';
import { Checker, openProgram, releaseProgram, type ProgramResult } from './checker.a';
import { hash } from './checker_hash.a';

// Test only: see Linter.junkRows.
const junkRows = programArguments().includes('--junk-rows');

function run(row: string, countOnly: boolean, program = 0, replayPrefix = '', recordPrefix = '', caseNumber = 0): number {
    const fields = row.split('\t');
    const path = fields[0] ?? panic('missing path');
    const source = readTextFile(path);
    if(source.kind === 'Error') {
        panic(source.message);
    }
    const parser = new Parser(source.text, path);
    const scanner = new Scanner(source.text);
    let replay: string | undefined = undefined;
    let replayError = '';
    const rowHeader = `${path}\n${hash(source.text)}\n`;
    if(replayPrefix !== '') {
        const transcript = readTextFile(`${replayPrefix}.${caseNumber}`);
        if(transcript.kind === 'Error') { replay = ''; replayError = transcript.message; }
        else if(!transcript.text.startsWith(rowHeader)) { replay = ''; replayError = 'transcript row differs'; }
        else { replay = transcript.text.slice(rowHeader.length); }
    }
    const checker = program === 0 && replayPrefix === '' ? undefined : new Checker(program, path, parser, source.text, replay);
    const settings = new Settings();
    settings.load(fields[5] ?? '');
    const linter = new Linter(
        source.text,
        parser,
        scanner,
        fields[1] ?? 'all',
        settings.read('mode', fields[2] ?? ''),
        settings.read('null', fields[3] ?? ''),
        settings.read('allowemptycatch', fields[4] === 'true' ? 'true' : 'false') === 'true',
        settings,
        checker,
    );
    linter.junkRows = junkRows;
    linter.run();
    if(checker !== undefined) {
        if(replayError !== '') { checker.refuse(0, 0, replayError); }
        if(recordPrefix !== '') { save(`${recordPrefix}.${caseNumber}`, rowHeader + checker.transcript()); }
        for(const refusal of checker.refusals) {
            console.log(`refused ${refusal.rule} ${path} ${refusal.start} ${refusal.end} ${written(refusal.reason)}`);
        }
    }
    if(!countOnly && program === 0 && replayPrefix === '') { for(const skipped of linter.skipped) { console.log(skipped); } }
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
        let repair = finding.repair;
        let replacement = finding.replacement;
        let description = finding.suggestion;
        let editStart = offsets[finding.editStart] ?? 0;
        let editEnd = offsets[finding.editEnd] ?? 0;
        let complete = false;
        if(finding.suggestions.length > 0) {
            const suggestion = finding.suggestions[0] ?? panic('missing suggestion');
            const edit = suggestion.edits[0];
            if(
                finding.suggestions.length === 1 &&
                suggestion.edits.length === 1 &&
                edit !== undefined &&
                edit.start === finding.start &&
                edit.end === finding.end
            ) {
                repair = 'suggestion';
                replacement = edit.text;
                description = suggestion.message;
                editStart = offsets[edit.start] ?? panic('suggestion outside source');
                editEnd = offsets[edit.end] ?? panic('suggestion end outside source');
            }
            else if(
                finding.suggestions.length === 1 &&
                suggestion.edits.length > 0 &&
                suggestion.edits.every((value) => !value.text.includes('|') && !value.text.includes(':'))
            ) {
                repair = `suggestion-edits:${suggestion.id}`;
                replacement = suggestion.edits
                    .map(
                        (value) =>
                            `${offsets[value.start] ?? panic('suggestion outside source')}:${offsets[value.end] ?? panic('suggestion end outside source')}:${value.text}`,
                    )
                    .join('|');
                description = suggestion.message;
            }
            else {
                repair = 'suggestions';
                replacement = '';
                description = '';
                complete = true;
            }
        }
        console.log(
            `range ${start} ${end} ${finding.id} ${repair}\t${written(replacement)}\t${written(description)}\t${editStart} ${editEnd}`,
        );
        for(const extra of finding.extraFixes) {
            console.log(
                `fix-edit\t${offsets[extra.start] ?? panic('fix outside source')} ${offsets[extra.end] ?? panic('fix end outside source')}\t${written(extra.text)}`,
            );
        }
        if(complete) {
            for(const suggestion of finding.suggestions) {
                console.log(
                    `suggestion\t${written(suggestion.id)}\t${written(suggestion.message)}\t${suggestion.edits.length}`,
                );
                for(const edit of suggestion.edits) {
                    console.log(
                        `suggestion-edit\t${offsets[edit.start] ?? panic('suggestion outside source')} ${offsets[edit.end] ?? panic('suggestion end outside source')}\t${written(edit.text)}`,
                    );
                }
            }
        }
    }
    if(fields[6] === 'recovery') {
        console.log('recovery findings only');
    }
    else {
        const fixed = linter.fixed();
        for(const rejection of linter.rejected) {
            console.log(rejection);
        }
        if(linter.unconverged.length > 0) {
            console.log(`unconverged\t${linter.unconverged.join(',')}`);
        }
        console.log(`fixed\t${written(fixed)}`);
    }
    return linter.findings.length;
}

function save(path: string, text: string): void {
    const result = writeTextFile(path, text);
    if(result.kind === 'Error') { panic(result.message); }
}
const args = programArguments();
const first = args[0] ?? panic('usage: main.ts <file> or --manifest <file> [--count]');
if(first === '--manifest') {
    const manifest = readTextFile(args[1] ?? panic('missing manifest'));
    if(manifest.kind === 'Error') {
        panic(manifest.message);
    }
    const countOnly = args.includes('--count');
    // `--shard <index>/<count>` runs every count-th row starting at index, numbering cases as the whole
    // manifest does, so the shards' outputs merge by case number into exactly the single-process output
    // (stage1/cohere/lint/shards). Rows are dealt out in turn rather than in blocks, which spreads a
    // directory of large files across the shards.
    let shardIndex = 0;
    let shardCount = 1;
    const shardFlag = args.indexOf('--shard');
    if(shardFlag >= 0) {
        const shard = (args[shardFlag + 1] ?? panic('missing shard')).split('/');
        shardIndex = Number.parseInt(shard[0] ?? panic('missing shard index'), 10);
        shardCount = Number.parseInt(shard[1] ?? panic('missing shard count'), 10);
        if(!(shardCount >= 1 && shardIndex >= 0 && shardIndex < shardCount)) {
            panic('shard must be <index>/<count> with 0 <= index < count');
        }
    }
    const rows = manifest.text.split('\n').filter((row) => row !== '');
    const declaration = rows[0] ?? '';
    const config = declaration.startsWith('program ') ? declaration.slice(8) : '';
    if(rows.slice(1).some((row) => row.startsWith('program '))) { panic('program must precede every row'); }
    const recordAt = args.indexOf('--record');
    const replayAt = args.indexOf('--replay');
    const recordPrefix = recordAt < 0 ? '' : args[recordAt + 1] ?? panic('missing recording prefix');
    const replayPrefix = replayAt < 0 ? '' : args[replayAt + 1] ?? panic('missing replay prefix');
    if(recordPrefix !== '' && replayPrefix !== '') { panic('record and replay are exclusive'); }
    if(config === '' && (recordPrefix !== '' || replayPrefix !== '')) { panic('transcript requires a program'); }
    let header = '';
    if(config !== '') {
        const contents = readTextFile(config);
        if(contents.kind === 'Error') { panic(contents.message); }
        header = `checker transcript 1\nprogram ${config}\nsha256 ${hash(contents.text)}\n`;
    }
    if(recordPrefix !== '') { save(`${recordPrefix}.header`, header); }
    let headerError = '';
    if(replayPrefix !== '') {
        const recorded = readTextFile(`${replayPrefix}.header`);
        if(recorded.kind === 'Error') { headerError = recorded.message; }
        else if(recorded.text !== header) { headerError = 'transcript program differs'; }
    }
    const opened: ProgramResult = config === '' || replayPrefix !== '' ? { kind: 'Ok', value: 0 } : openProgram(config, []);
    const program = opened.kind === 'Ok' ? opened.value : 0;
    if(opened.kind === 'Error') { headerError = opened.message; }
    let count = 0;
    let caseNumber = 0;
    for(const row of rows) {
        if(row === '' || row.startsWith('program ')) {
            continue;
        }
        if(caseNumber % shardCount === shardIndex) {
            if(!countOnly) {
                console.log(`case ${caseNumber}`);
            }
            if(headerError !== '') { console.log(`refused ${row.split('\t')[1] ?? 'all'} ${row.split('\t')[0] ?? ''} 0 0 ${written(headerError)}`); }
        else { count += run(row, countOnly, program, replayPrefix, recordPrefix, caseNumber); }
        }
        caseNumber++;
    }
    if(program !== 0) {
        const released = releaseProgram(program);
        if(released.kind === 'Error') { console.log(`refused checker ${config} 0 0 ${written(released.message)}`); }
    }
    if(countOnly) {
        console.log(`${count}`);
    }
}
else {
    run(first, false);
}
