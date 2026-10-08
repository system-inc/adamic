// The manifest is the full Go construction census. Eligibility is an explicit slice boundary;
// every eligible row must independently parse, lower and match, and every other row is a decline.
import { panic, readTextFile } from 'adamic';
import { lowerSourceAt } from './lower.ts';
import { SymbolSnapshot } from './symbol.ts';
import { dump } from './dump.ts';
export function constructionCoverage(path: string): void {
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
        const fn = lowerSourceAt(source.text, Number.parseInt(fields[10] ?? fields[2] ?? '', 10), Number.parseInt(fields[11] ?? fields[3] ?? '', 10), symbols, fields[12] ?? '') ?? panic(`eligible construction declined: ${key}`);
        console.log(dump(fn).trimEnd());
    }
}
