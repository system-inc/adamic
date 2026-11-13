const fs=require('fs'),ts=require('/workspace/timer-ambient-delete/stage3/api/node_modules/typescript'),assert=require('assert/strict');
const before=process.argv[2]||'/workspace/scratch/timer-main',after=process.argv[3]||'/workspace/scratch/timer-after';
function read(file){return fs.readFileSync(file,'utf8').trim().split('\n').map(JSON.parse);}
function ledger(root,file){
 const rows=read(file);assert.deepEqual(rows[0].diagnostics,[]);const row=rows.find(r=>r.file===root+'/src/compiler/sys.ts');assert(row);
 const text=fs.readFileSync(row.file,'utf8').replace(/\r\n/g,'\n'),source=ts.createSourceFile(row.file,text,99,true);
 const units=new Map(row.units.map(u=>[u.where,u]));
 function scope(where){const [,file,line,col]=where.match(/^(.*):(\d+):(\d+)$/);if(file!==row.file)return file.replace(root,'');const p=source.getPositionOfLineAndCharacter(+line-1,+col-1);let best=[];
 function visit(n,parents){if(p<n.pos||p>=n.end)return;let chain=parents;if(n.name&&(ts.isFunctionDeclaration(n)||ts.isClassDeclaration(n)||ts.isInterfaceDeclaration(n)||ts.isTypeAliasDeclaration(n)||ts.isMethodDeclaration(n)||ts.isVariableDeclaration(n)||ts.isEnumDeclaration(n)))chain=[...parents,n.name.getText(source)];if(chain.length>best.length)best=chain;ts.forEachChild(n,c=>visit(c,chain));}visit(source,[]);return best.join('.')||units.get(where)?.name||'module';}
 return row.findings.map(f=>({...f,where:f.where.replace(root+'/',''),unit:f.unit.replace(root+'/',''),unitOwner:scope(f.unit),findingOwner:scope(f.where)}));
}
const b=ledger(before,process.argv[4]||'/workspace/scratch/timer-census-before.jsonl'),a=ledger(after,process.argv[5]||'/workspace/scratch/timer-census-after-final.jsonl');
const key=f=>JSON.stringify([f.unitOwner,f.findingOwner,f.kind,f.phase,f.reason]);
const buckets=new Map();for(const f of b){const k=key(f);if(!buckets.has(k))buckets.set(k,[]);buckets.get(k).push(f);}
const added=[],kept=[];for(const f of a){const q=buckets.get(key(f));if(q?.length)kept.push({before:q.shift(),after:f});else added.push(f);}
const gone=[...buckets.values()].flat();
const result={method:'Pair findings by named unit scope, named finding scope, kind, phase and reason, preserving multiplicity. Raw before/after locations are retained for all pairs; this avoids treating line shifts as new findings.',before:b.length,after:a.length,retained:kept.length,gone,new:added,retainedPairs:kept};
fs.writeFileSync(process.argv[6]||'/workspace/scratch/timer-census-delta.json',JSON.stringify(result,null,2)+'\n');console.log(JSON.stringify({before:b.length,after:a.length,retained:kept.length,gone,new:added},null,2));
