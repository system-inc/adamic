import {readFileSync} from 'node:fs';
import {createHash} from 'node:crypto';
import {pathToFileURL} from 'node:url';
const tree=process.argv[2],inputs=JSON.parse(readFileSync(process.argv[3],'utf8'));
const parser=await import(pathToFileURL(tree+'/src/compiler/parser.ts'));
const types=await import(pathToFileURL(tree+'/src/compiler/types.ts'));
const {ScriptTarget,ScriptKind,SyntaxKind,JSDocParsingMode}=types;
function scriptKind(name){return name.endsWith('.tsx')?ScriptKind.TSX:name.endsWith('.jsx')?ScriptKind.JSX:/\.(js|mjs|cjs)$/.test(name)?ScriptKind.JS:ScriptKind.TS;}
function count(file){
 const rows=[],counts={nodes:0,jsdoc_roots:0,jsdoc_descendants:0,tags:0,type_expressions:0,links:0};const pending=[[file,false]];
 while(pending.length){const [node,inDoc]=pending.pop(),doc=inDoc||node.kind===SyntaxKind.JSDoc;
  counts.nodes++;if(doc)counts.jsdoc_descendants++;
  if(node.kind===SyntaxKind.JSDoc)counts.jsdoc_roots++;
  if(node.kind>=SyntaxKind.FirstJSDocTagNode&&node.kind<=SyntaxKind.LastJSDocTagNode)counts.tags++;
  if(node.kind===SyntaxKind.JSDocTypeExpression)counts.type_expressions++;
  if([SyntaxKind.JSDocLink,SyntaxKind.JSDocLinkCode,SyntaxKind.JSDocLinkPlain].includes(node.kind))counts.links++;
  rows.push([node.kind,node.pos,node.end,node.flags,node.text,node.escapedText,node.tagName?.escapedText,node.comment===undefined?null:typeof node.comment==='string'?node.comment:node.comment.map(p=>[p.kind,p.text])]);
  const children=[];for(const d of node.jsDoc??[])children.push([d,true]);parser.forEachChild(node,c=>{children.push([c,doc]);});
  for(let i=children.length-1;i>=0;i--)pending.push(children[i]);
 }
 const diagnostics=[file.parseDiagnostics??[],file.jsDocDiagnostics??[]].map(ds=>ds.map(d=>[d.code,d.start,d.length,d.messageText]));
 return {...counts,parse_diagnostics:diagnostics[0].length,jsdoc_diagnostics:diagnostics[1].length,sha256:createHash('sha256').update(JSON.stringify([rows,diagnostics])).digest('hex')};
}
function parse(input,mode=JSDocParsingMode.ParseAll){return parser.createSourceFile(input.name,input.text,{languageVersion:ScriptTarget.Latest,jsDocParsingMode:mode},false,scriptKind(input.name));}
const rows=inputs.map(input=>{
 const observed=count(parse(input));if(input.reference_nodes!==undefined&&observed.nodes!==input.reference_nodes)throw Error('reference node count drift '+input.name+' '+observed.nodes+' vs '+input.reference_nodes);
 return {name:input.name,group:input.group,raw_sha256:input.raw_sha256,...observed};
});
const options=[];
for(const kind of ['TS','TSX','JS','JSX'])for(const mode of ['ParseAll','ParseNone','ParseForTypeErrors','ParseForTypeInfo'])for(const link of [false,true]){
 const source=(link?'/** @see Target */':'/** @param {string} x */')+'\nfunction f(x) {}';
 const sf=parser.createSourceFile('mode.'+kind.toLowerCase(),source,{languageVersion:ScriptTarget.Latest,jsDocParsingMode:JSDocParsingMode[mode]},false,ScriptKind[kind]);options.push({kind,mode,link,...count(sf)});
}
const timings={};
for(const group of ['compiler','parser','jsx','salsa','remaining','directed']){
 const selection=inputs.filter(i=>i.group===group);for(const i of selection)parse(i);
 const rounds=[];
 for(let round=0;round<5;round++){
  const pair={};for(const mode of round%2?['ParseNone','ParseAll']:['ParseAll','ParseNone']){
   const stats={calls:0,ms:0};globalThis.__jsdocTiming=stats;let total=0;
   for(const i of selection){const start=performance.now();parse(i,JSDocParsingMode[mode]);total+=performance.now()-start;}
   globalThis.__jsdocTiming=undefined;pair[mode]={parse_ms:total,jsdoc_ms:stats.ms,calls:stats.calls};
  }rounds.push(pair);
 }timings[group]={files:selection.length,rounds};
}
const typeInputs=['{*}','{!number}','{?}','{?number}','{...number=}'];
const typeTrees=typeInputs.map(text=>{const r=parser.parseJSDocTypeExpressionForTests(text);return {text,tree:count(r.jsDocTypeExpression),diagnostics:r.diagnostics.map(d=>d.code)};});
console.log(JSON.stringify({source:'actual adapted parser; stock API used only for transpilation',timing_instrumented:process.env.JSDOC_TIMING==='1',node:process.version,rows,options,timings,typeTrees},null,2));
