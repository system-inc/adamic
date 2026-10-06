import fs from 'node:fs';
import { createRequire } from 'node:module';
const require = createRequire(`${process.argv[2]}/package.json`);
if (require('prettier/package.json').version !== '3.9.6') throw Error('expected prettier 3.9.6');
const print = require('prettier/plugins/postcss').printers.postcss.print;
const decode = text => text.replace(/\\([nrt\\])/g, (_, c) => ({n:'\n',r:'\r',t:'\t','\\':'\\'})[c]);
const encode = text => text.replace(/[\n\r\t\\]/g, c => ({'\n':'\\n','\r':'\\r','\t':'\\t','\\':'\\\\'})[c]);
for (const line of fs.readFileSync(process.argv[3], 'utf8').split('\n')) {
 if (!line) continue;
 const node = {type:'selector-string',value:decode(line.slice(1))};
 console.log(encode(print({node}, {parser:'css',singleQuote:line[0]==='s'}, () => { throw Error('unexpected child'); })));
}
