'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto'),ts=require('typescript');
const pins=require('./owners.json');
const normalize=text=>text.replace(/\r\n/g,'\n');
const hash=text=>crypto.createHash('sha256').update(text).digest('hex');
function ownerPath(node){const result=[];for(;node;node=node.parent)if(ts.isFunctionLike(node)&&node.name)result.unshift(node.name.getText(node.getSourceFile()));return result;}
function inspect(file,raw){
 const source=ts.createSourceFile(file,raw,ts.ScriptTarget.Latest,true);assert.equal(source.parseDiagnostics.length,0,file+': parse error');
 const nodes=[];function visit(n){nodes.push(n);ts.forEachChild(n,visit);}visit(source);
 const states=[],edits=[];
 for(const pin of pins.filter(p=>p.file===file)){
  let matches,start,end;
  if(pin.kind==='coreGroup'){
   matches=source.statements.filter(n=>ts.isFunctionDeclaration(n)&&['or','orPredicate2','orPredicate3'].includes(n.name?.text));
   assert.equal(matches.length,5,file+': missing, duplicate or half-adapted or declaration group');
   const names=matches.map(n=>n.name.text);assert(names.every(n=>n==='or')||JSON.stringify(names)===JSON.stringify(['orPredicate2','orPredicate3','or','or','or']),file+': half-adapted helper declarations');
   start=matches[0].getStart(source);end=matches.at(-1).end;
  }else{
   if(pin.kind==='import')matches=source.statements.filter(n=>ts.isImportDeclaration(n)&&n.moduleSpecifier.text===pin.module);
   else if(pin.kind==='variable')matches=source.statements.filter(n=>ts.isVariableStatement(n)&&n.declarationList.declarations.some(d=>d.name.getText(source)===pin.path[0]));
   else matches=nodes.filter(n=>ts.isFunctionLike(n)&&n.name&&n.body&&JSON.stringify(ownerPath(n))===JSON.stringify(pin.path));
   assert.equal(matches.length,1,file+': missing or duplicate owner '+(pin.path?.join('/')||pin.module));start=matches[0].getStart(source);end=matches[0].end;
  }
  const spelling=normalize(raw.slice(start,end));
  const adapted=spelling===pin.after;
  assert(spelling===pin.before||adapted,file+': changed or half-adapted owner '+(pin.path?.join('/')||pin.module||'or declaration group'));
  assert.equal(hash(spelling),adapted?pin.afterSha256:pin.beforeSha256,file+': owner hash mismatch');
  states.push(adapted);
  if(!adapted){const eol=raw.slice(start,end).includes('\r\n')?'\r\n':'\n';edits.push({start,end,after:pin.after.replace(/\n/g,eol)});}
 }
 assert(states.length&&states.every(s=>s===states[0]),file+': half-adapted file');return {adapted:states[0],edits};
}
function adapt(tree){
 assert.equal(ts.version,'6.0.3');
 const plans=[...new Set(pins.map(p=>p.file))].map(file=>{
  const name=path.join(tree,file),raw=fs.readFileSync(name,'utf8'),state=inspect(file,raw);let after=raw,boundary=raw.length;
  for(const e of state.edits.sort((a,b)=>b.start-a.start)){assert(e.end<=boundary,file+': overlapping owners');boundary=e.start;after=after.slice(0,e.start)+e.after+after.slice(e.end);}
  assert(inspect(file,after).adapted,file+': incomplete planned result');return {name,raw,after,...state};
 });
 assert(plans.every(p=>p.adapted===plans[0].adapted),'half-adapted tree');
 for(const plan of plans)assert.equal(fs.readFileSync(plan.name,'utf8'),plan.raw,'source changed while planning');
 if(!plans[0].adapted)for(const plan of plans)fs.writeFileSync(plan.name,plan.after);
 return {files:plans[0].adapted?0:plans.length,owners:plans.reduce((n,p)=>n+p.edits.length,0),calls:plans[0].adapted?0:19,javascriptChanging:true};
}
module.exports={adapt,inspect,pins};
if(require.main===module){assert.equal(process.argv.length,3,'usage: node adapt.cjs TREE');console.log(JSON.stringify(adapt(path.resolve(process.argv[2])),null,2));}
