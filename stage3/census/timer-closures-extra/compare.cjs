'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict');
const ts=require('/workspace/adamic/stage3/api/node_modules/typescript');
const D=__dirname,trees={before:'/workspace/scratch/timer-main-tip-final',after:'/workspace/scratch/timer-after-tip'};
function ledger(label,file){
 const root=trees[label],rows=fs.readFileSync(path.join(D,label+'-'+file.replace('.ts','')+'.jsonl'),'utf8').trim().split('\n').map(JSON.parse);
 assert.deepEqual(rows[0].diagnostics,[]);assert.equal(rows[0].latent_mode,'full');const row=rows.find(r=>r.file===root+'/src/compiler/'+file);assert(row);const trailer=rows.find(r=>r.status==='relation_report');assert(trailer);assert.deepEqual(trailer.findings,[]);
 const cache=new Map();function source(name){if(!cache.has(name)){const text=fs.readFileSync(name,'utf8').replace(/\r\n/g,'\n');cache.set(name,ts.createSourceFile(name,text,99,true));}return cache.get(name);}
 const units=new Map(row.units.map(u=>[u.where,u]));
 function scope(where){const [,name,line,col]=where.match(/^(.*):(\d+):(\d+)$/);const s=source(name),p=s.getPositionOfLineAndCharacter(+line-1,+col-1);let best=[];
  function visit(n,parents){if(p<n.pos||p>=n.end)return;let chain=parents;if(n.name&&(ts.isFunctionDeclaration(n)||ts.isClassDeclaration(n)||ts.isInterfaceDeclaration(n)||ts.isTypeAliasDeclaration(n)||ts.isMethodDeclaration(n)||ts.isVariableDeclaration(n)||ts.isEnumDeclaration(n)))chain=[...parents,n.name.getText(s)];if(chain.length>best.length)best=chain;ts.forEachChild(n,c=>visit(c,chain));}visit(s,[]);return best.join('.')||units.get(where)?.name||'module';}
 function line(where){const [,name,n]=where.match(/^(.*):(\d+):(\d+)$/);return source(name).text.split('\n')[+n-1];}
 const findings=row.findings.map(f=>({...f,where:f.where.replace(root+'/',''),unit:f.unit.replace(root+'/',''),unitOwner:scope(f.unit),findingOwner:scope(f.where),sourceLine:line(f.where)}));
 const counts={};for(const f of findings)counts[f.kind]=(counts[f.kind]||0)+1;
 return {counts,total:findings.length,units:row.units.length,findings};
}
const output={method:'Same named-scope pairing as the sys.ts proof: unit scope, finding scope, kind, phase and reason, preserving multiplicity. All paired raw positions are retained, so line shifts are not reported as new findings.',files:{}};
for(const file of ['watchPublic.ts','tsbuildPublic.ts']){
 const before=ledger('before',file),after=ledger('after',file),key=f=>JSON.stringify([f.unitOwner,f.findingOwner,f.kind,f.phase,f.reason]);const buckets=new Map();for(const f of before.findings){const k=key(f);if(!buckets.has(k))buckets.set(k,[]);buckets.get(k).push(f);}
 const added=[],retained=[];for(const f of after.findings){const q=buckets.get(key(f));if(q?.length)retained.push({before:q.shift(),after:f});else added.push(f);}const gone=[...buckets.values()].flat();
 output.files[file]={before:{counts:before.counts,total:before.total,units:before.units},after:{counts:after.counts,total:after.total,units:after.units},retained:retained.length,gone,new:added,retainedPairs:retained};
 console.log(file,JSON.stringify({before:output.files[file].before,after:output.files[file].after,retained:retained.length,gone:gone.length,new:added.length}));
 for(const [kind,findings] of [['GONE',gone],['NEW',added]])for(const f of findings)console.log(JSON.stringify({kind,...f}));
}
fs.writeFileSync(path.join(D,'comparison.json'),JSON.stringify(output,null,2)+'\n');
