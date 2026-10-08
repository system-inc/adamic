import { createRequire } from 'node:module';
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
const [library,path] = process.argv.slice(2);
const require = createRequire(join(library,'package.json'));
if(require('yaml/package.json').version !== '2.9.0') throw new Error('expected yaml 2.9.0');
const { Parser } = require('yaml');
const base=join(library,'node_modules/yaml/dist/compose');
const { resolveFlowScalar } = require(join(base,'resolve-flow-scalar.js'));
const { resolveBlockScalar } = require(join(base,'resolve-block-scalar.js'));
const hex=text=>{let result='';for(let i=0;i<text.length;i++)result+=text.charCodeAt(i).toString(16).padStart(4,'0');return result;};
function visit(token) {
    if(token == null) return;
    if(['scalar','single-quoted-scalar','double-quoted-scalar','block-scalar'].includes(token.type)) {
        for(const root of [true,false]) {
            const errors=[];
            const onError=(source,code,message)=>{
                let pos;
                if(typeof source==='number') pos=[source,source+1];
                else if(Array.isArray(source)) pos=source.slice(0,2);
                else pos=[source.offset,source.offset+(typeof source.source==='string'?source.source.length:1)];
                errors.push(`${pos.join(',')},${code},${hex(message)}`);
            };
            const result=token.type==='block-scalar'?resolveBlockScalar({atRoot:root,options:{strict:true}},token,onError):resolveFlowScalar(token,true,onError);
            console.log(`${token.type}|${token.offset}|${root?1:0}|${hex(result.value)}|${result.type??''}|${result.range.join(',')}|${hex(result.comment??'')}|${errors.join(';')}`);
        }
    }
    visit(token.value);
    for(const item of token.items??[]) {visit(item.key);visit(item.value);}
}
const unescape=text=>text.replace(/\\([\\nrt])/g,(_,char)=>char==='n'?'\n':char==='r'?'\r':char==='t'?'\t':char);
let number=0;
for(const line of readFileSync(path,'utf8').split('\n')) {
    if(line==='')continue;
    const tab=line.indexOf('\t');const size=Number(line.slice(0,tab));const text=unescape(line.slice(tab+1));
    const parser=new Parser();console.log(`case ${number++}`);
    const emit=(source,incomplete)=>{for(const token of parser.parse(source,incomplete))visit(token);};
    if(size===0)emit(text,false);
    else {for(let start=0;start<text.length;start+=size)emit(text.slice(start,start+size),true);emit('',false);}
}
