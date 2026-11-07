'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),ts=require('typescript');
const [before,after,mutant,scratch,output]=process.argv.slice(2);
function read(label,tree){
 const rows=fs.readFileSync(path.join(scratch,'site-'+label+'.jsonl'),'utf8').trim().split('\n').map(JSON.parse);
 assert.deepEqual(rows[0].diagnostics,[],'checker diagnostics '+label);
 const sf=ts.createSourceFile('core.ts',fs.readFileSync(path.join(tree,'src/compiler/core.ts'),'utf8'),ts.ScriptTarget.Latest,true),sites=[],raw=new Set(),occurrences=new Map();
 for(const finding of rows.flatMap(r=>r.findings||[])){
  if(finding.kind==='Boundary')continue;
  const match=finding.where?.match(/:(\d+):(\d+)$/);if(!match)continue;
  const rawKey=JSON.stringify([finding.where,finding.kind,finding.reason,finding.phase]);if(raw.has(rawKey))continue;raw.add(rawKey);
  const pos=sf.getPositionOfLineAndCharacter(+match[1]-1,+match[2]-1),node=ts.getTokenAtPosition(sf,pos),owners=[];
  for(let p=node;p;p=p.parent)if(ts.isFunctionLike(p)&&p.name)owners.unshift(p.name.getText(sf));
  const record={owner:owners.join('.'),kind:finding.kind,reason:finding.reason,phase:finding.phase,snippet:node.parent.getText(sf)};
  const key=JSON.stringify(record),occurrence=occurrences.get(key)||0;occurrences.set(key,occurrence+1);sites.push({...record,occurrence,where:finding.where});
 }
 return {sites,units:rows.flatMap(r=>r.units||[]).length};
}
const a=read('before',before),b=read('after',after),m=read('mutant',mutant),key=r=>JSON.stringify([r.owner,r.kind,r.reason,r.phase,r.snippet,r.occurrence]),as=new Set(a.sites.map(key)),bs=new Set(b.sites.map(key));
const removed=a.sites.filter(r=>!bs.has(key(r))),added=b.sites.filter(r=>!as.has(key(r))),restored=m.sites.filter(r=>r.owner==='getAllKeys');
assert.equal(removed.length,1,'unexpected removed sites');assert.equal(added.length,0,'new core site');
assert.equal(removed[0].owner,'getAllKeys');assert.equal(removed[0].kind,'Refused');assert.equal(removed[0].reason,'Object.getPrototypeOf');assert.equal(removed[0].phase,'lowering');
assert.deepEqual(restored.map(key),removed.map(key),'mutant failed to restore F10');assert.deepEqual(new Set(m.sites.map(key)),as,'mutant differs beyond restoring original owner');
fs.writeFileSync(output,JSON.stringify({passed:true,beforeSites:a.sites.length,afterSites:b.sites.length,mutantSites:m.sites.length,beforeUnits:a.units,afterUnits:b.units,removed,added,mutantRestored:restored,checkerDiagnostics:0,referenceSite:'F10, core-sites-af42d3dc/REPORT.md'},null,2)+'\n');console.log(JSON.stringify({before:a.sites.length,after:b.sites.length,removed:removed.length,newSites:added.length,mutantRestored:restored.length}));
