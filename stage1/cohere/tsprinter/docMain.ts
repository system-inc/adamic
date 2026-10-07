import { panic, programArguments, readTextFile } from 'adamic';
import { Documents } from './doc.ts';
import { escaped, unescaped } from './transport.ts';
const args = programArguments();
const input = readTextFile(args[0] ?? panic('usage: docMain.ts <doc-protocol>'));
if(input.kind === 'Error') panic(input.message);
let docs = new Documents({ printWidth: 120, tabWidth: 4, useTabs: false });
for(const line of input.text.split('\n')) {
    if(line === '') continue;
    const fields = line.split('\t');
    const kind = fields[0] ?? panic('missing doc kind');
    if(kind === 'reset') {
        docs = new Documents({
            printWidth: Number(fields[1] ?? ''),
            tabWidth: Number(fields[2] ?? ''),
            useTabs: fields[3] === '1',
        });
    }
    else if(kind === 'print') {
        console.log(`ok\t${escaped(docs.print(Number(fields[1] ?? '')))}`);
    }
    else {
        const children: number[] = [];
        for(const child of (fields[5] ?? '').split(',')) if(child !== '') children.push(Number(child));
        docs.add(
            kind,
            children,
            unescaped(fields[4] ?? ''),
            Number(fields[1] ?? ''),
            fields[3] ?? '',
            fields[2] === '1',
        );
    }
}
