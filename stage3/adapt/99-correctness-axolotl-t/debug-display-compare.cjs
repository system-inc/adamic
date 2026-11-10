'use strict';
const fs=require('fs'),path=require('path'),assert=require('node:assert/strict'),crypto=require('crypto'),ts=require('/workspace/adamic/stage3/api/node_modules/typescript');
const before='/workspace/scratch/debug-display-before',after='/workspace/scratch/debug-display-candidate',unit='/workspace/adamic/stage3/adapt/99-correctness-axolotl-t';
const names=['attachFlowNodeDebugInfoWorker','attachFlowNodeDebugInfo','attachNodeArrayDebugInfoWorker','attachNodeArrayDebugInfo','enableDebugInfo'];
const hash=b=>crypto.createHash('sha256').update(b).digest('hex');
function owners(text,file){const sf=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true,ts.ScriptKind.JS),result=new Map();function visit(n){if(ts.isFunctionDeclaration(n)&&n.name){const name=names.find(name=>new RegExp('^'+name+'[0-9]*$').test(n.name.text));if(name){assert(!result.has(name),file+' duplicate '+name);result.set(name,{start:n.getStart(sf),end:n.end,text:n.getText(sf)});}}ts.forEachChild(n,visit);}visit(sf);return result;}
const rows=[];
for(const file of fs.readdirSync(path.join(before,'built/local')).filter(f=>f.endsWith('.js')).sort()){
 const left=fs.readFileSync(path.join(before,'built/local',file),'utf8'),right=fs.readFileSync(path.join(after,'built/local',file),'utf8');
 if(left===right){rows.push({file,identical:true,sha256:hash(left)});continue;}
 const a=owners(left,file),b=owners(right,file);assert.equal(a.size,5);assert.equal(b.size,5);let expected=left;
 for(const[name,owner]of [...a].sort((x,y)=>y[1].start-x[1].start))expected=expected.slice(0,owner.start)+b.get(name).text+expected.slice(owner.end);
 if(right!==expected){fs.writeFileSync('/workspace/scratch/debug-display-expected-'+file,expected);rows.push({file,identical:false,only_five_owners:false,before_sha256:hash(left),after_sha256:hash(right)});continue;}rows.push({file,identical:false,only_five_owners:true,before_sha256:hash(left),after_sha256:hash(right)});
}
const declarations=[];for(const file of ['typescript.d.ts','tsserverlibrary.d.ts']){const a=fs.readFileSync(path.join(before,'built/local',file)),b=fs.readFileSync(path.join(after,'built/local',file));assert(a.equals(b),file);declarations.push({file,bytes:a.length,sha256:hash(a),identical:true});}
const walk=root=>fs.readdirSync(root,{withFileTypes:true}).flatMap(e=>e.isDirectory()?walk(path.join(root,e.name)):[path.join(root,e.name)]);
const changed=[];for(const p of walk(path.join(before,'src'))){const rel=path.relative(before,p),a=fs.readFileSync(p),b=fs.readFileSync(path.join(after,rel));if(!a.equals(b))changed.push(rel);}assert.deepEqual(changed,['src/compiler/debug.ts']);
fs.writeFileSync(path.join(unit,'debug-display-normal-comparison.json'),JSON.stringify({status:rows.every(r=>r.identical||r.only_five_owners)?'pass':'blocked',javascript:rows,declarations,changed_source:changed},null,2)+'\n');console.log(JSON.stringify({files:rows.length,changed:rows.filter(r=>!r.identical).map(r=>r.file),declarations:'identical',source:changed}));

if(rows.some(r=>!r.identical&&!r.only_five_owners))process.exitCode=1;
