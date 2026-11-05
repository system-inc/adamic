'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto'),ts=require('typescript');
const owners=require('./owners.json');
const digest=text=>crypto.createHash('sha256').update(text).digest('hex');
function parse(file,text) {const source=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true);assert.equal(source.parseDiagnostics.length,0,file+': parse error');return source;}
function ancestry(node) {const result=[];for(let p=node.parent;p;p=p.parent)if(ts.isFunctionLike(p))result.unshift(p.name?.text ?? '<'+ts.SyntaxKind[p.kind]+'>');return result;}
function locate(file,text) {
 const source=parse(file,text),states=[],edits=[];
 for(const pin of owners.filter(p=>p.file===file)) {
  const found=[];function visit(node){if(ts.isFunctionDeclaration(node)&&node.name?.text===pin.owner)found.push(node);ts.forEachChild(node,visit);} visit(source);
  assert([pin.beforeDeclarations,pin.afterDeclarations].includes(found.length),pin.owner+': owner count drift');
  for(const node of found)assert.deepEqual(ancestry(node),pin.ancestors,pin.owner+': moved or nested owner');
  let container=source;
  for(const ancestor of pin.ancestors){const matches=container.statements.filter(n=>ts.isFunctionDeclaration(n)&&n.name?.text===ancestor&&n.body);assert.equal(matches.length,1,pin.owner+': missing or duplicate ancestor');container=matches[0].body;}
  assert(found.every(n=>n.parent===container),pin.owner+': owner is not directly in its pinned container');
  assert(found.every(n=>n.parent===found[0].parent),pin.owner+': split declarations');
  const start=found[0].getStart(source),end=found.at(-1).end,spelling=text.slice(start,end),hash=digest(spelling);
  assert([pin.ownerBeforeSha256,pin.ownerAfterSha256].includes(hash),pin.owner+': changed or half-adapted owner');
  const adapted=hash===pin.ownerAfterSha256;
  assert.equal(spelling,adapted?pin.after:pin.before,pin.owner+': exact text drift');
  states.push(adapted);if(!adapted)edits.push({start,end,after:pin.after});
 }
 assert(states.length&&states.every(s=>s===states[0]),file+': half-adapted file');
 return {adapted:states[0],edits};
}
function transpile(text,removeComments=false) {return ts.transpileModule(text,{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments}}).outputText;}
function adapt(tree) {
 assert.equal(ts.version,'6.0.3');
 const plans=[...new Set(owners.map(p=>p.file))].map(file=>{
  const name=path.join(tree,'src/compiler',file),raw=fs.readFileSync(name,'utf8'),text=raw.replace(/\r\n/g,'\n'),state=locate(file,text);
  let after=text;
  for(const edit of [...state.edits].sort((a,b)=>b.start-a.start))after=after.slice(0,edit.start)+edit.after+after.slice(edit.end);
  assert(locate(file,after).adapted,file+': planned result invalid');
  if(file==='core.ts')for(const removeComments of [false,true])assert.equal(transpile(after,removeComments),transpile(text,removeComments),'core.ts JavaScript changed');
  if(raw.includes('\r\n'))after=after.replace(/\n/g,'\r\n');
  return {name,raw,after,...state};
 });
 assert(plans.every(p=>p.adapted===plans[0].adapted),'half-adapted tree');
 for(const plan of plans)assert.equal(fs.readFileSync(plan.name,'utf8'),plan.raw,'source changed while planning');
 if(!plans[0].adapted)for(const plan of plans)fs.writeFileSync(plan.name,plan.after);
 return {files:plans[0].adapted?0:plans.length,edits:plans.reduce((sum,p)=>sum+p.edits.length,0),javascriptChanging:true};
}
module.exports={adapt,owners,parse,transpile,digest};
if(require.main===module){assert.equal(process.argv.length,3,'usage: node adapt.cjs TREE');console.log(JSON.stringify(adapt(path.resolve(process.argv[2])),null,2));}
