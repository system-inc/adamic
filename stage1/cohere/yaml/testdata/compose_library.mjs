import { createRequire } from 'node:module';
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
const [library,path] = process.argv.slice(2);
const require = createRequire(join(library,'package.json'));
if(require('yaml/package.json').version !== '2.9.0') throw new Error('expected yaml 2.9.0');
const { Parser } = require('yaml');
const {Composer,isPair,isMap,isSeq,isAlias}=require('yaml');
const hex=text=>{let result='';for(let i=0;i<text.length;i++)result+=text.charCodeAt(i).toString(16).padStart(4,'0');return result;};
const token=t=>t==null?'-':`${t.type},${t.offset},${hex(t.source??'')}`;
function outline(n){
 if(n==null)return '-';
 const kind=isPair(n)?'Pair':isAlias(n)?'Alias':isMap(n)?'Map':isSeq(n)?'Seq':'Scalar';
 return `${kind}|${n.constructor.name}|${n.range?.join(',')??'-'}|${isPair(n)?'-':token(n.srcToken)}|${isPair(n)&&n.srcToken?1:0}|${'anchor' in n?1:0}|${hex(n.anchor??'')}|${hex(n.tag??'')}|${hex(n.comment??'')}|${hex(n.commentBefore??'')}|${n.spaceBefore?1:0}|${n.flow?1:0}|${'source' in n?1:0}|${hex(n.source??'')}|${n.type??''}|${n.format??''}|${n.minFractionDigits??0}|[${(n.items??[]).map(item=>outline(item)+',').join('')}]|${isPair(n)?outline(n.key):'-'}|${isPair(n)?outline(n.value):'-'}`;
}
function describe(doc){
 const d=doc.directives;
 const tags=Object.entries(d.tags).map(([key,value])=>`${hex(key)}=${hex(value)}`).sort();
 const errors=list=>list.map(e=>`${e.pos.join(',')},${e.code},${hex(e.message)}`).join(';');
 return `${d.yaml.version}|${d.yaml.explicit?1:0}|${d.docStart?1:0}|${d.docEnd?1:0}|${tags.join(';')}|${doc.range.join(',')}|${hex(doc.commentBefore??'')}|${hex(doc.comment??'')}|${outline(doc.contents)}|errors:${errors(doc.errors)}|warnings:${errors(doc.warnings)}`;
}
const unescape=text=>text.replace(/\\([\\nrt])/g,(_,char)=>char==='n'?'\n':char==='r'?'\r':char==='t'?'\t':char);
let number=0;
for(const line of readFileSync(path,'utf8').split('\n')) {
    if(line==='')continue;
    const tab=line.indexOf('\t');const size=Number(line.slice(0,tab));const text=unescape(line.slice(tab+1));
    const parser=new Parser();console.log(`case ${number++}`);
    const tokens=[];const emit=(source,incomplete)=>{for(const token of parser.parse(source,incomplete))tokens.push(token);};
    if(size===0)emit(text,false);
    else {for(let start=0;start<text.length;start+=size)emit(text.slice(start,start+size),true);emit('',false);}
 for(const doc of new Composer({keepSourceTokens:true,uniqueKeys:false,merge:true}).compose(tokens,true,text.length))console.log(describe(doc));
}
