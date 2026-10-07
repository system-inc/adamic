// Map immutable upstream ledger identities to adapted AST casts. Ambiguous or
// changed groups remain unknown; offsets are never guessed from line numbers.
const fs=require('node:fs'),path=require('node:path'),cp=require('node:child_process'),assert=require('node:assert/strict');
const ts=require('/home/agent/.cache/adamic-stage3/api/node_modules/typescript');
assert.equal(ts.version,'6.0.3');
const [adapted,output]=process.argv.slice(2);
const sites=JSON.parse(fs.readFileSync('stage3/interface-downcasts/shape-conformance-sites.json'));
assert.equal(sites.length,2936);
function inventory(text,file) {
 const source=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true), rows=[];
 function owner(node) {
  const parts=[];
  for(let p=node.parent;p;p=p.parent) if(ts.isFunctionLike(p)) {
   const name=p.name?.getText(source)|| (ts.isVariableDeclaration(p.parent)?p.parent.name.getText(source):ts.isPropertyAssignment(p.parent)?p.parent.name.getText(source):'<anonymous>');
   parts.unshift(name);
  }
  return parts.join('/')||'<module>';
 }
 function visit(node) {
  if(ts.isAsExpression(node)) rows.push({start:node.getStart(source),end:node.end,byte_start:Buffer.byteLength(text.slice(0,node.getStart(source))),byte_end:Buffer.byteLength(text.slice(0,node.end)),bodies:(()=>{const bodies=[];for(let p=node.parent;p;p=p.parent)if(ts.isFunctionLike(p)&&p.body)bodies.push({start:Buffer.byteLength(text.slice(0,p.body.pos)),end:Buffer.byteLength(text.slice(0,p.body.end))});return bodies})(),key:owner(node)+'\u0000'+node.getText(source).replace(/\s+/g,' '),owner:owner(node),text:node.getText(source)});
  ts.forEachChild(node,visit);
 }
 visit(source);return rows;
}
const files=[...new Set(sites.map(s=>s.file))], mapped=[];
for(const file of files) {
 const original=cp.execFileSync('git',['--git-dir=/home/agent/.cache/adamic-stage3/typescript.git','show','050880ce:'+file],{encoding:'utf8',maxBuffer:16*1024*1024});
 const old=inventory(original,file),next=inventory(fs.readFileSync(path.join(adapted,file),'utf8'),file);
 const oldGroups=new Map(),newGroups=new Map();
 for(const [rows,groups] of [[old,oldGroups],[next,newGroups]]) for(const row of rows) {if(!groups.has(row.key))groups.set(row.key,[]);groups.get(row.key).push(row);}
 for(const site of sites.filter(s=>s.file===file)) {
  const origin=old.find(n=>n.start===site.start&&n.end===site.end);assert(origin);assert.equal(origin.text,site.text);
  const a=oldGroups.get(origin.key),b=newGroups.get(origin.key)||[];
  const match=a.length===b.length?b[a.indexOf(origin)]:undefined;
  mapped.push({...site,measurement:'measured on a checker-rejected program',adapted:match?{start:match.byte_start,end:match.byte_end,owner:match.owner,text:match.text,bodies:match.bodies}:null,mapping:match?'exact AST text and owner with equal occurrence counts':'changed or ambiguous adapted cast group'});
 }
}
assert.equal(mapped.length,2936);fs.writeFileSync(output,JSON.stringify(mapped));
console.log(JSON.stringify({measurement:'measured on a checker-rejected program',sites:mapped.length,mapped:mapped.filter(r=>r.adapted).length,unmapped:mapped.filter(r=>!r.adapted).length}));
