// Only yaml 2.9.0, with no port code or generated Adamic output.
import { createRequire } from 'node:module';
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
const [library, path] = process.argv.slice(2);
const require = createRequire(join(library, 'package.json'));
const version = require('yaml/package.json').version;
if (version !== '2.9.0') throw new Error(`expected yaml 2.9.0, got ${version}`);
const {Schema}=require('yaml');
const tags=['core','yaml-1.1'].flatMap(schema=>new Schema({schema,merge:true}).tags.filter(tag=>tag.test));
const unescape=text=>text.replace(/\\([\\nrt])/g,(_,char)=>char==='n'?'\n':char==='r'?'\r':char==='t'?'\t':char);
let number=0;
for(const line of readFileSync(path,'utf8').split('\n')) {
 if(line==='')continue;
 const text=unescape(line.slice(line.indexOf('\t')+1)).trim();
 console.log(`${number++}|${tags.map(tag=>tag.test.test(text)?'1':'0').join('')}`);
}
