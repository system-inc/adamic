'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto'),ts=require('typescript');
const pins=require('./owners.json'),hash=s=>crypto.createHash('sha256').update(s).digest('hex');
function parse(file,text){const source=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true);assert.equal(source.parseDiagnostics.length,0,file+': parse error');return source;}
function ownerPath(n){const result=[];for(let p=n;p;p=p.parent)if((ts.isFunctionLike(p)||ts.isClassDeclaration(p))&&p.name)result.unshift(p.name.getText(p.getSourceFile()));return result;}
function inspect(file,text){const source=parse(file,text),nodes=[];function visit(n){nodes.push(n);ts.forEachChild(n,visit);}visit(source);const states=[],edits=[];
 for(const pin of pins.filter(p=>p.file===file)){let matches;
  if(pin.kind==='import')matches=source.statements.filter(n=>ts.isImportDeclaration(n)&&n.moduleSpecifier.text===pin.module);
  else if(pin.kind==='declaration')matches=source.statements.filter(n=>n.name?.text===pin.path[0]);
  else matches=nodes.filter(n=>(ts.isFunctionDeclaration(n)||ts.isMethodDeclaration(n))&&n.body&&n.name&&JSON.stringify(ownerPath(n))===JSON.stringify(pin.path));
  assert(matches.length===1||pin.after===''&&matches.length===0,file+': missing or duplicate '+(pin.path?.join('.')||pin.module));
  const node=matches[0],spelling=node?node.getText(source):'',adapted=spelling===pin.after;assert(spelling===pin.before||adapted,file+': changed or half-adapted owner '+(pin.path?.join('.')||pin.module));
  assert.equal(hash(spelling),adapted?pin.afterSha256:pin.beforeSha256,'pin digest');states.push(adapted);
  if(!adapted)edits.push({start:node.getStart(source),end:node.end,after:pin.after});
 }
 assert(states.length&&states.every(s=>s===states[0]),file+': half-adapted file');return {adapted:states[0],edits};
}
function adapt(tree){assert.equal(ts.version,'6.0.3');const plans=[...new Set(pins.map(p=>p.file))].map(file=>{const name=path.join(tree,file),raw=fs.readFileSync(name,'utf8'),text=raw.replace(/\r\n/g,'\n'),state=inspect(file,text);let after=text,boundary=text.length;for(const edit of state.edits.sort((a,b)=>b.start-a.start)){assert(edit.end<=boundary,'overlapping owners');boundary=edit.start;after=after.slice(0,edit.start)+edit.after+after.slice(edit.end);}assert(inspect(file,after).adapted,'invalid planned result');if(raw.includes('\r\n'))after=after.replace(/\n/g,'\r\n');return {name,raw,after,...state};});assert(plans.every(p=>p.adapted===plans[0].adapted),'half-adapted tree');for(const p of plans)assert.equal(fs.readFileSync(p.name,'utf8'),p.raw,'source changed while planning');if(!plans[0].adapted)for(const p of plans)fs.writeFileSync(p.name,p.after);return {files:plans[0].adapted?0:plans.length,edits:plans.reduce((n,p)=>n+p.edits.length,0),javascriptChanging:true};}
module.exports={adapt,inspect,pins};if(require.main===module){assert.equal(process.argv.length,3,'usage: node adapt.cjs TREE');console.log(JSON.stringify(adapt(path.resolve(process.argv[2])),null,2));}
