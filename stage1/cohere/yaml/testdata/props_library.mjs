import { createRequire } from 'node:module';
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
const [library,path] = process.argv.slice(2);
const require = createRequire(join(library,'package.json'));
if(require('yaml/package.json').version !== '2.9.0') throw new Error('expected yaml 2.9.0');
const { Parser } = require('yaml');
const { resolveProps } = require(join(library,'node_modules/yaml/dist/compose/resolve-props.js'));
const hex=text=>{let result='';for(let i=0;i<text.length;i++)result+=text.charCodeAt(i).toString(16).padStart(4,'0');return result;};
const describe=token=>token==null?'-':`${token.type},${token.offset},${hex(token.source)}`;
function resolve(tokens,flow,indicator,next,offset,parentIndent,startOnNewline) {
    const errors=[],warnings=[];
    const onError=(source,code,message,warning)=>{
        const pos=typeof source==='number'?[source,source+1]:Array.isArray(source)?source.slice(0,2):[source.offset,source.offset+(typeof source.source==='string'?source.source.length:1)];
        (warning?warnings:errors).push(`${pos.join(',')},${code},${hex(message)}`);
    };
    const result=resolveProps(tokens,{flow:flow||undefined,indicator,next,offset,parentIndent,startOnNewline,onError});
    console.log(`${describe(result.comma)}|${describe(result.found)}|${result.spaceBefore?1:0}|${hex(result.comment)}|${result.hasNewline?1:0}|${describe(result.anchor)}|${describe(result.tag)}|${describe(result.newlineAfterProp)}|${result.end}|${result.start}|${errors.join(';')}|${warnings.join(';')}`);
}
function visit(token) {
    if(token==null)return;
    if(token.type==='document') resolve(token.start,'','doc-start',token.value??token.end?.[0],token.offset,0,true);
    else {
        const flow=token.type==='flow-collection'?(token.start.source==='{'?'flow map':'flow sequence'):'';
        for(const item of token.items??[]) {
            const next=item.key??item.sep?.[0];
            resolve(item.start,flow,token.type==='block-seq'?'seq-item-ind':'explicit-key-ind',token.type==='block-seq'?item.value:next,token.offset,token.indent,flow==='');
            if(item.sep!=null)resolve(item.sep,flow,'map-value-ind',item.value,token.offset,token.indent,false);
        }
    }
    visit(token.value);
    for(const item of token.items??[]){visit(item.key);visit(item.value);}
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
