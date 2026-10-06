import { createRequire } from 'node:module';
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
const [library,path] = process.argv.slice(2);
const require = createRequire(join(library,'package.json'));
if(require('yaml/package.json').version !== '2.9.0') throw new Error('expected yaml 2.9.0');
const { Parser } = require('yaml');
const {Composer,isPair,isMap,isSeq,isAlias}=require('yaml');
const hex=text=>{let result='';for(let i=0;i<text.length;i++)result+=text.charCodeAt(i).toString(16).padStart(4,'0');return result;};
const { parse } = await import(join(library,'node_modules/yaml-unist-parser/dist/parse.mjs'));
let currentText='';
const point=p=>{
 let offset=p.offset;const code=currentText.charCodeAt(offset),previous=currentText.charCodeAt(offset-1);
 if(code>=0xdc00&&code<=0xdfff&&previous>=0xd800&&previous<=0xdbff)offset--;
 return `${p.line},${p.column},${offset}`;
};
function outline(n){
 if(n==null)return '-';
 const parent=n._parent==null?'-':`${n._parent.type},${point(n._parent.position.start)},${point(n._parent.position.end)}`;
 const list=items=>`[${(items??[]).map(item=>outline(item)+',').join('')}]`;
 return `${n.type}|${point(n.position.start)}|${point(n.position.end)}|${parent}|${hex(n.value??'')}|${n.chomping??''}|${n.indent??-1}|${n.directivesEndMarker?1:0}|${n.documentEndMarker?1:0}|${hex(n.name??'')}|${(n.parameters??[]).map(hex).join(',')}|${list(n.children)}|${outline(n.tag)}|${outline(n.anchor)}|${list(n.middleComments)}|${list(n.leadingComments)}|${outline(n.trailingComment)}|${list(n.endComments)}|${outline(n.indicatorComment)}|${list(n.comments)}`;
}
const unescape=text=>text.replace(/\\([\\nrt])/g,(_,char)=>char==='n'?'\n':char==='r'?'\r':char==='t'?'\t':char);
let number=0;
for(const line of readFileSync(path,'utf8').split('\n')) {
    if(line==='')continue;
    const tab=line.indexOf('\t');const size=Number(line.slice(0,tab));const text=unescape(line.slice(tab+1));
    currentText=text;console.log(`case ${number++}`);
    try { console.log(outline(parse(text,{uniqueKeys:false}))); }
    catch(error) {
      if(error.name==='YAMLSyntaxError')console.log(`syntax|${error.code}|${hex(error.message)}|${point(error.position.start)}|${point(error.position.end)}`);
      else console.log(`throw|${error.name}|${hex(error.message)}`);
    }
}
