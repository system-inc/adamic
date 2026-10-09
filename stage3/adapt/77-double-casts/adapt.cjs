'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),ts=require('typescript');
assert.equal(ts.version,'6.0.3');
assert.equal(process.argv.length,3,'usage: adapt.cjs <tree>');
const tree=path.resolve(process.argv[2]),sites=require('./sites.json'),plans=[];
function casts(text,file) {
 const source=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true),nodes=[];
 function visit(n){if(ts.isAsExpression(n))nodes.push(n);ts.forEachChild(n,visit);}visit(source);
 return {source,nodes};
}
for(const file of [...new Set(sites.map(s=>s.file))]) {
 const name=path.join(tree,file),before=fs.readFileSync(name,'utf8'),parsed=casts(before,file),edits=[];
 const matches=sites.filter(s=>s.file===file),seen=new Map();
 for(const s of matches) {
  const key=s.before,ordinal=seen.get(key)||0;seen.set(key,ordinal+1);
  const candidates=parsed.nodes.filter(n=>[s.before,s.after].includes(n.getText(parsed.source)));
  assert.equal(candidates.length,matches.filter(m=>m.before===s.before).length,'reviewed site drift: '+s.id);
  const node=candidates[ordinal];
  if(s.action==='checked-downcast'&&node.getText(parsed.source)===s.before)edits.push({start:node.getStart(parsed.source),end:node.end,text:s.after});
 }
 const known=new Set(matches.flatMap(s=>[s.before,s.after]));
 for(const n of parsed.nodes)if(ts.isAsExpression(n.expression)&&[ts.SyntaxKind.AnyKeyword,ts.SyntaxKind.UnknownKeyword].includes(n.expression.type.kind))assert(known.has(n.getText(parsed.source)),'unreviewed double cast: '+n.getText(parsed.source));
 let after=before;for(const e of edits.sort((a,b)=>b.start-a.start))after=after.slice(0,e.start)+e.text+after.slice(e.end);
 assert.equal(casts(after,file).source.parseDiagnostics.length,0,'syntax');
 for(const removeComments of [false,true]) {
  const compilerOptions={target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.ESNext,removeComments};
  assert.equal(ts.transpileModule(after,{fileName:file,compilerOptions}).outputText,ts.transpileModule(before,{fileName:file,compilerOptions}).outputText,'runtime bytes: '+file);
 }
 plans.push({name,before,after,changed:edits.length});
}
// Inventory the entire current compiler closure, including generated sources.
const knownRemaining=new Set(sites.filter(s=>s.action==='unresolved').map(s=>s.file+':'+s.before));
let bridges=0;
for(const entry of fs.readdirSync(path.join(tree,'src/compiler')).filter(f=>f.endsWith('.ts'))) {
 const file='src/compiler/'+entry,name=path.join(tree,file),plan=plans.find(p=>p.name===name),text=plan?plan.after:fs.readFileSync(name,'utf8'),parsed=casts(text,file);
 for(const n of parsed.nodes) {
  let inner=n.expression;while(ts.isParenthesizedExpression(inner))inner=inner.expression;
  if(ts.isAsExpression(inner)&&[ts.SyntaxKind.AnyKeyword,ts.SyntaxKind.UnknownKeyword].includes(inner.type.kind)) {
   bridges++;assert(knownRemaining.has(file+':'+n.getText(parsed.source)),'unreviewed closure bridge: '+file+':'+n.getText(parsed.source));
  }
 }
}
assert.equal(bridges,14,'closure bridge inventory drift');
for(const p of plans)assert.equal(fs.readFileSync(p.name,'utf8'),p.before,'concurrent edit');
for(const p of plans)if(p.after!==p.before)fs.writeFileSync(p.name,p.after);
console.log(JSON.stringify({adaptation:77,removed:plans.reduce((n,p)=>n+p.changed,0),reviewed:18,unresolved:14}));
