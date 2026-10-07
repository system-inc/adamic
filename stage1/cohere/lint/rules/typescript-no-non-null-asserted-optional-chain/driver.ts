import { panic, programArguments, readTextFile } from 'adamic';
import { written } from '../../../../typescript/parser/nodes.ts';
import { Parser } from '../../../../typescript/parser/parser.ts';
import { Scanner } from '../../../../typescript/scanner/scanner.ts';
import { Settings } from '../../settings.ts';
import { Linter } from '../../lint.ts';
import { diagnostics, enableCompleteSuggestions } from './diagnostic.ts';
function run(row: string, countOnly: boolean): number {
    const fields = row.split('\t');
    const path = fields[0] ?? panic('missing path');
    const source = readTextFile(path);
    if(source.kind === 'Error') { panic(source.message); }
    while(diagnostics.length > 0) { diagnostics.pop(); }
    const parser = new Parser(source.text, path);
    const scanner = new Scanner(source.text);
    const settings = new Settings();
    settings.load(fields[5] ?? '');
    const linter = new Linter(source.text, parser, scanner, fields[1] ?? 'all', '', '', false, settings);
    linter.run();
    if(countOnly) { return diagnostics.length; }
    const offsets: number[] = [0];
    let bytes = 0;
    for(let index = 0; index < source.text.length; index++) {
        const code = source.text.codePointAt(index) ?? 0;
        if(code > 65535) { offsets.push(bytes); index++; }
        bytes += code < 128 ? 1 : code < 2048 ? 2 : code < 65536 ? 3 : 4;
        offsets.push(bytes);
    }
    diagnostics.sort((left, right) => left.start - right.start);
    if(linter.findings.length !== diagnostics.length) { panic('selected unvalidated rule'); }
    for(const diagnostic of diagnostics) {
        console.log(`finding ${diagnostic.rule} ${diagnostic.id} ${offsets[diagnostic.start] ?? panic('start outside source')} ${offsets[diagnostic.end] ?? panic('end outside source')}\t${written(diagnostic.message)}`);
        console.log(`fixes 0`);
        console.log(`suggestions ${diagnostic.suggestions.length}`);
        for(const suggestion of diagnostic.suggestions) {
            console.log(`suggestion ${suggestion.id}\t${written(suggestion.message)}`);
            console.log(`edits ${suggestion.edits.length}`);
            for(const edit of suggestion.edits) {
                console.log(`edit ${offsets[edit.start] ?? panic('edit start outside source')} ${offsets[edit.end] ?? panic('edit end outside source')}\t${written(edit.text)}`);
            }
        }
    }
    console.log(`fixed\t${written(source.text)}`);
    return diagnostics.length;
}
enableCompleteSuggestions();
const args = programArguments();
const path = args[1] ?? panic('usage: --manifest <path> [--count]');
const input = readTextFile(path);
if(input.kind === 'Error') { panic(input.message); }
const countOnly = args.includes('--count');
let ordinal = 0;
let total = 0;
for(const row of input.text.split('\n')) {
    if(row === '') { continue; }
    if(!countOnly) { console.log(`case ${ordinal}`); }
    total += run(row, countOnly);
    ordinal++;
}
if(countOnly) { console.log(`${total}`); }
