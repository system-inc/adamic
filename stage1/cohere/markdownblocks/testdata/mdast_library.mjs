// Expose the pinned original from-markdown entry/config by reference, unchanged.
import fs from 'node:fs';
import vm from 'node:vm';
let bundle=fs.readFileSync(process.argv[2]+'/plugins/markdown.js','utf8');
const anchor='function _i(e){';
if(bundle.split(anchor).length!==2 || !bundle.includes('function xg(){'))throw new Error('pinned mdast anchor');
bundle=bundle.replace(anchor,'globalThis.adamicFromMarkdown=fi;globalThis.adamicCompileMdast=UD;globalThis.adamicSliceChunks=ND;globalThis.adamicSerializeChunks=RD;globalThis.adamicFrontMatter=Ve;globalThis.adamicParseMarkdown=_i;globalThis.adamicMarkdownConfig=xg;'+anchor);
const context={module:{exports:{}},exports:{}};
vm.runInNewContext(bundle,context);
const encode=text=>text.replace(/[\\\n\r\t]/g,c=>c==='\\'?'\\\\':c==='\n'?'\\n':c==='\r'?'\\r':'\\t');
const optional=value=>value===null||value===undefined?'-':'+'+value;
const canonical=root=>{
 const rows=[],stack=[root];
 while(stack.length){
  const n=stack.pop(),p=n.position;
  const fields=[n.type,'children'in n?n.children.length:-1,'value'in n,n.value===null,n.value??'',n.depth??0,n.ordered??false,typeof n.start==='number'?optional(n.start):'-',n.spread??false,typeof n.checked==='boolean'?n.checked:'-',optional(n.lang),optional(n.meta),n.url??'',optional(n.title),optional(n.alt),optional(n.label),n.identifier??'',n.referenceType??'',(n.align??[]).map(x=>x??'').join(','),p?[p.start.line,p.start.column,p.start.offset,p.end.line,p.end.column,p.end.offset].join(','):'0,0,0,0,0,0'];
  fields.push(n.type==='frontMatter'?'+'+n.language:'-',n.type==='frontMatter'?optional(n.explicitLanguage):'-',n.type==='frontMatter'?n.value:'',n.type==='frontMatter'?n.startDelimiter:'',n.type==='frontMatter'?n.endDelimiter:'',n.type==='frontMatter'?n.raw:'');
  if(n.type==='frontMatter'){fields[2]=false;fields[4]='';}
  rows.push(fields.map(x=>encode(''+x)).join('\t'));
  if(n.children)for(let i=n.children.length-1;i>=0;i--)stack.push(n.children[i]);
 }
 return rows.join('\n');
};
const decode=text=>text.replace(/\\([\s\S])/g,(_,c)=>c==='n'?'\n':c==='r'?'\r':c==='t'?'\t':c);
const tokenPoint=value=>{const f=value.split(',').map(Number);return {line:f[0],column:f[1],offset:f[2],_index:f[3],_bufferIndex:f[4]}};
if(process.argv[3]==='--error'){
 const paragraph={type:'paragraph',start:{line:1,column:1,offset:0},end:{line:1,column:1,offset:0}},strong={...paragraph,type:'strong'};
 const name=process.argv[4],source={sliceSerialize:()=>''};
 const events=name==='unclosed'?[['enter',paragraph,source]]:name==='not-open'?[['exit',paragraph,source]]:[['enter',paragraph,source],['exit',strong,source]];
 try{context.adamicCompileMdast(context.adamicMarkdownConfig())(events);process.stdout.write('no error\n')}catch(error){process.stdout.write(error.message+'\n')}
 process.exit(0);
}
const output=[];
if(process.argv[3]==='--events'){
 let source='',tokens=[],contexts=[],events=[];
 for(const line of fs.readFileSync(process.argv[4],'utf8').split('\n')){
  const f=line.split('\t');
  if(f[0]==='S')source=decode(f[1]);
  else if(f[0]==='C'){
   const chunks=f[1].split(';').filter(Boolean).map(v=>{if(v[0]==='C')return Number(v.slice(1))===0?null:Number(v.slice(1));const units=v.slice(1)===''?[]:v.slice(1).split(',').map(Number);let text='';for(let i=0;i<units.length;i+=8192)text+=String.fromCharCode(...units.slice(i,i+8192));return text});
   contexts.push({sliceSerialize:(token,expand)=>context.adamicSerializeChunks(context.adamicSliceChunks(chunks,token),expand)});
  }else if(f[0]==='T')tokens.push({type:f[1],start:tokenPoint(f[2]),end:tokenPoint(f[3]),_spread:f[4]==='true',_align:f[5]===''?[]:f[5].split(',')});
  else if(f[0]==='E')events.push([f[1]==='true'?'enter':'exit',tokens[Number(f[2])],contexts[Number(f[3])]]);
  else if(f[0]==='F'){
   const tree=context.adamicCompileMdast(context.adamicMarkdownConfig())(events),{frontMatter,content}=context.adamicFrontMatter(source);
   if(frontMatter){const p=frontMatter;tree.children.unshift({...p,type:'frontMatter',position:{start:{line:p.start.line,column:p.start.column+1,offset:p.start.index},end:{line:p.end.line,column:p.end.column+1,offset:p.end.index}}})}
   output.push(encode('content\t'+encode(content)+'\n'+canonical(tree)));source='';tokens=[];contexts=[];events=[];
  }else if(line!=='')throw new Error('event oracle protocol');
 }
}else{
 for(const line of fs.readFileSync(process.argv[3],'utf8').split('\n')){
  if(!line)continue;const c=JSON.parse(line),source=c.text.replace(/\r\n?/g,'\n').replace(/^\uFEFF/,'');
  output.push(encode('content\t'+encode(context.adamicFrontMatter(source).content)+'\n'+canonical(context.adamicParseMarkdown(source))));
 }
}
process.stdout.write(output.join('\n')+'\n');
