import { panic, programArguments, readTextFile } from 'adamic';
import { OptionsJson } from '../../helpers/options_json.ts';
import { written } from '../../../../typescript/parser/nodes.ts';
import { decode } from './options.ts';
const input = readTextFile(programArguments()[0] ?? panic('cases path'));
if(input.kind === 'Error') { panic(input.message); }
const cases = new OptionsJson(input.text);
const root = cases.parse();
for(const row of cases.node(root).children) {
    const result = decode(cases.node(row).text);
    console.log(result.valid ? 'valid' : 'invalid');
    if(!result.valid) { continue; }
    const keys: string[] = [];
    for(const [key, requires] of result.requirements) { keys.push(key); }
    keys.sort((left, right) => {
        let a = 0; let b = 0;
        while(a < left.length && b < right.length) {
            const x = left.codePointAt(a) ?? 0; const y = right.codePointAt(b) ?? 0;
            if(x !== y) { return x < y ? -1 : 1; }
            a += x > 65535 ? 2 : 1; b += y > 65535 ? 2 : 1;
        }
        return left.length - right.length;
    });
    for(const key of keys) {
        console.log(`key\t${written(key)}`);
        for(const value of result.requirements.get(key) ?? panic('missing configured key')) { console.log(`requires\t${written(value)}`); }
    }
}
