// The manifest is the full Go construction census. Eligibility is an explicit slice boundary;
// every eligible row must independently parse, lower and match, and every other row is a decline.
import { panic, readTextFile } from 'adamic';
import { lowerSourceAt } from './lower.ts';
import { ConstructedHIR } from './core.ts';
import { HIRFile, ForFunction } from './cache.ts';
import { SymbolSnapshot } from './symbol.ts';
import { dump } from './dump.ts';
export function constructionCoverage(path: string, cached: boolean = false): void {
    const manifest = readTextFile(path);
    if(manifest.kind === 'Error') { panic(manifest.message); }
    for(const line of manifest.text.split('\n')) {
        if(line === '') { continue; }
        const fields = line.split('\t');
        const key = fields[0] ?? panic('missing key');
        console.log(`case\t${key}`);
        if(fields[7] !== 'true') { console.log('decline'); continue; }
        const source = readTextFile(fields[1] ?? panic('missing source path'));
        if(source.kind === 'Error') { panic(source.message); }
        let symbols: SymbolSnapshot | undefined;
        if(fields[4] === 'true') {
            const facts = readTextFile(fields[8] ?? panic('missing symbol snapshot'));
            if(facts.kind === 'Error') { panic(facts.message); }
            symbols = new SymbolSnapshot(facts.text);
        }
        const start = Number.parseInt(fields[10] ?? fields[2] ?? '', 10); const end = Number.parseInt(fields[11] ?? fields[3] ?? '', 10);
        const nestedPath = fields[12] ?? '';
        let fn: ConstructedHIR;
        if(cached && symbols !== undefined) {
            const file = new HIRFile(source.text, symbols); const node = file.at(start, end);
            const first = ForFunction(file, node) ?? panic(`eligible cache construction declined: ${key}`);
            if(first !== ForFunction(file, node) || first !== ForFunction(file, node)) { panic('ForFunction cache identity changed'); }
            let index = first.root;
            if(nestedPath !== '') { for(const part of nestedPath.split(',')) { index = first.arena.read(index).functions[Number.parseInt(part, 10)] ?? panic('missing cached nested path'); } }
            fn = new ConstructedHIR(first.arena, index);
        }
        else {
            if(cached) { const file = new HIRFile(source.text, undefined); if(ForFunction(file, file.at(start, end)) !== undefined) { panic('checker-less cache constructed'); } }
            fn = lowerSourceAt(source.text, start, end, symbols, nestedPath) ?? panic(`eligible construction declined: ${key}`);
        }
        console.log(dump(fn).trimEnd());
    }
}
