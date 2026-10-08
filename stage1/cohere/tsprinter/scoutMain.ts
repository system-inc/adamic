import { panic, programArguments, readTextFile } from 'adamic';
import { formatFile } from './files.ts';
import { escaped, unescaped } from './transport.ts';
const input = readTextFile(programArguments()[0] ?? panic('scoutMain.ts <batch>'));
if(input.kind === 'Error') panic(input.message);
for(const line of input.text.split('\n')) {
    if(line === '') continue;
    const fields = line.split('\t');
    const path = fields[0] ?? panic('missing path');
    const source = unescaped(fields[1] ?? panic('missing source'));
    const result = formatFile(source, { printWidth: 120, tabWidth: 4, useTabs: false }, path);
    console.log(result.kind === 'Ok' ? `ok\t${escaped(result.text)}` : `notyet\t${result.reason}`);
}
