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
    for(const value of result.ignored) { console.log(`ignore\t${written(value)}`); }
    for(const value of result.additional) { console.log(`additional\t${written(value)}`); }
}
