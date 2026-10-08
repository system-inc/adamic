// The manifest is the full Go construction census. Eligibility is an explicit slice boundary;
// every eligible row must independently parse, lower and match, and every other row is a decline.
import { panic, readTextFile } from 'adamic';
import { lowerSourceAt } from './lower.ts';
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
        const fn = lowerSourceAt(source.text, Number.parseInt(fields[2] ?? '', 10), Number.parseInt(fields[3] ?? '', 10)) ?? panic(`eligible construction declined: ${key}`);
        console.log(dump(fn).trimEnd());
    }
}
