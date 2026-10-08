import { parallelMap, panic, programArguments, readTextFile, writeTextFile } from 'adamic';
import { written } from '../../typescript/parser/nodes.ts';
import { Parser } from '../../typescript/parser/parser.ts';
import { Scanner } from '../../typescript/scanner/scanner.ts';
import { Linter } from './lint.ts';
import { Settings } from './settings.ts';
import { Checker, openProgram, releaseProgram, type ProgramResult } from './checker.a';
import { hash } from './checker_hash.a';

// Test only: see Linter.junkRows.
const junkRows = programArguments().includes('--junk-rows');

interface FileResult {
    readonly count: number;
    readonly output: string;
}
interface FileJob {
    readonly row: string;
    readonly caseNumber: number;
}

// A worker owns its walk, fixes and rendering. Only immutable results cross the join.
function run(row: string, countOnly: boolean, program = 0, replayPrefix = '', recordPrefix = '', caseNumber = 0): FileResult {
    const output: string[] = [];
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
    const rowHeader = recordPrefix === '' && replayPrefix === '' ? '' : `${path}\n${hash(source.text)}\n`;
    if(replayPrefix !== '') {
        const transcript = readTextFile(`${replayPrefix}.${caseNumber}`);
        if(transcript.kind === 'Error') { replay = ''; replayError = transcript.message; }
        else if(!transcript.text.startsWith(rowHeader)) { replay = ''; replayError = 'transcript row differs'; }
        else { replay = transcript.text.slice(rowHeader.length); }
    }
    const checker = program === 0 && replayPrefix === '' ? undefined : new Checker(program, path, parser, source.text, replay, recordPrefix !== '');
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
            output.push(`refused ${refusal.rule} ${path} ${refusal.start} ${refusal.end} ${written(refusal.reason)}`);
        }
    }
    if(!countOnly && program === 0 && replayPrefix === '') { for(const skipped of linter.skipped) { output.push(skipped); } }
    if(countOnly) {
        return { count: linter.findings.length, output: output.length === 0 ? '' : output.join('\n') + '\n' };
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
        output.push(`${path}:${line}:${column}\n  ${finding.rule}  ${finding.message}\n`);
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
        output.push(
            `range ${start} ${end} ${finding.id} ${repair}\t${written(replacement)}\t${written(description)}\t${editStart} ${editEnd}`,
        );
        for(const extra of finding.extraFixes) {
            output.push(
                `fix-edit\t${offsets[extra.start] ?? panic('fix outside source')} ${offsets[extra.end] ?? panic('fix end outside source')}\t${written(extra.text)}`,
            );
        }
        if(complete) {
            for(const suggestion of finding.suggestions) {
                output.push(
                    `suggestion\t${written(suggestion.id)}\t${written(suggestion.message)}\t${suggestion.edits.length}`,
                );
                for(const edit of suggestion.edits) {
                    output.push(
                        `suggestion-edit\t${offsets[edit.start] ?? panic('suggestion outside source')} ${offsets[edit.end] ?? panic('suggestion end outside source')}\t${written(edit.text)}`,
                    );
                }
            }
        }
    }
    if(fields[6] === 'recovery') {
        output.push('recovery findings only');
    }
    else {
        const fixed = linter.fixed();
        for(const rejection of linter.rejected) {
            output.push(rejection);
        }
        if(linter.unconverged.length > 0) {
            output.push(`unconverged\t${linter.unconverged.join(',')}`);
        }
        output.push(`fixed\t${written(fixed)}`);
    }
    return { count: linter.findings.length, output: output.length === 0 ? '' : output.join('\n') + '\n' };
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
    const error = headerError;
    const jobs: FileJob[] = [];
    let caseNumber = 0;
    for(const row of rows) {
        if(row === '' || row.startsWith('program ')) { continue; }
        if(caseNumber % shardCount === shardIndex) { jobs.push({ row, caseNumber }); }
        caseNumber++;
    }
    const results = parallelMap(jobs, (job): FileResult => {
        const prefix = countOnly ? '' : `case ${job.caseNumber}\n`;
        // Checker handles are task-local too, never one program shared across files.
        const opened: ProgramResult = config === '' || replayPrefix !== '' ? { kind: 'Ok', value: 0 } : openProgram(config, []);
        const program = opened.kind === 'Ok' ? opened.value : 0;
        const refusal = error !== '' ? error : opened.kind === 'Error' ? opened.message : '';
        const result = refusal !== '' ? {
            count: 0,
            output: `refused ${job.row.split('\t')[1] ?? 'all'} ${job.row.split('\t')[0] ?? ''} 0 0 ${written(refusal)}\n`,
        } : run(job.row, countOnly, program, replayPrefix, recordPrefix, job.caseNumber);
        let suffix = '';
        if(program !== 0) {
            const released = releaseProgram(program);
            if(released.kind === 'Error') { suffix = `refused checker ${config} 0 0 ${written(released.message)}\n`; }
        }
        return { count: result.count, output: prefix + result.output + suffix };
    });
    let count = 0;
    for(const result of results) {
        count += result.count;
        if(result.output !== '') { console.log(result.output.slice(0, -1)); }
    }
    if(countOnly) {
        console.log(`${count}`);
    }
}
else {
    const result = run(first, false);
    if(result.output !== '') { console.log(result.output.slice(0, -1)); }
}
