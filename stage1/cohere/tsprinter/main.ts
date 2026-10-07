import { panic, programArguments, readTextFile } from 'adamic';
import { formatExpression } from './expressions.ts';
import { escaped, unescaped } from './transport.ts';
const args = programArguments();
const batch = args[0] === '--cases';
const input = readTextFile(args[batch ? 1 : 0] ?? panic('usage: main.ts <expression-file> | --cases <batch> [width]'));
if(input.kind === 'Error') panic(input.message);
const settings = { printWidth: args[2] === undefined ? 120 : Number(args[2]), tabWidth: 4, useTabs: false };
if(batch) {
    for(const line of input.text.split('\n')) {
        if(line === '') continue;
        const result = formatExpression(unescaped(line.slice(1)), settings);
        console.log(result.kind === 'Ok' ? `ok\t${escaped(result.text)}` : `notyet\t${result.reason}`);
    }
}
else {
    const result = formatExpression(input.text, settings);
    if(result.kind === 'NotYet') panic(`printer slice not yet: ${result.reason}`);
    console.log(result.text.slice(0, -1));
}
