import { panic, programArguments, readTextFile } from 'adamic';
import { written } from '../../typescript/parser/nodes.ts';
import { Parser } from '../../typescript/parser/parser.ts';
import { Scanner } from '../../typescript/scanner/scanner.ts';
import { Linter } from './lint.ts';
import { Settings } from './settings.ts';

// Test only: see Linter.junkRows.
const junkRows = programArguments().includes('--junk-rows');

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
        settings.read('mode', fields[2] ?? ''),
        settings.read('null', fields[3] ?? ''),
        settings.read('allowemptycatch', fields[4] === 'true' ? 'true' : 'false') === 'true',
        settings,
    );
    linter.junkRows = junkRows;
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
    let count = 0;
    let caseNumber = 0;
    for(const row of manifest.text.split('\n')) {
        if(row === '') {
            continue;
        }
        if(caseNumber % shardCount === shardIndex) {
            if(!countOnly) {
                console.log(`case ${caseNumber}`);
            }
            count += run(row, countOnly);
        }
        caseNumber++;
    }
    if(countOnly) {
        console.log(`${count}`);
    }
}
else {
    run(first, false);
}
