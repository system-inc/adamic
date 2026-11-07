'use strict';
const fs=require('node:fs'),ts=require('/workspace/adamic/stage3/api/node_modules/typescript'),assert=require('node:assert/strict');
const s='/workspace/scratch/elementat-declaration',result={};
for(const variant of ['baseline','candidate']){
 const tree=s+'/'+variant,rows=fs.readFileSync(s+'/'+variant+'-census.jsonl','utf8').trim().split('\n').map(JSON.parse),metadata=rows[0],core=rows.find(r=>r.file?.endsWith('/core.ts'));
 assert.equal(metadata.checker_rejected,false);assert.deepEqual(metadata.diagnostics,[]);assert(core);
 const file=tree+'/src/compiler/core.ts',sf=ts.createSourceFile(file,fs.readFileSync(file,'utf8'),ts.ScriptTarget.Latest,true);
 const relevant={};
 for(const name of ['elementAt','toOffset']){
  const node=sf.statements.find(n=>ts.isFunctionDeclaration(n)&&n.name?.text===name||ts.isVariableStatement(n)&&n.declarationList.declarations.some(d=>d.name.getText(sf)===name));assert(node);
  const start=sf.getLineAndCharacterOfPosition(node.getStart(sf)).line+1,end=sf.getLineAndCharacterOfPosition(node.end).line+1;
  const within=where=>{const m=where?.match(/core\.ts:(\d+):\d+$/);return !!m&&Number(m[1])>=start&&Number(m[1])<=end;};
  relevant[name]={start,end,units:core.units.filter(u=>within(u.where)),findings:core.findings.filter(f=>within(f.where)||within(f.unit))};
 }
 const normalize=f=>({...f,where:f.where.replace(tree+'/src/compiler/',''),unit:f.unit.replace(tree+'/src/compiler/','')});
 const unique=new Map(core.findings.map(f=>[JSON.stringify([f.where,f.kind,f.reason,f.phase]),normalize(f)]));
 result[variant]={metadata,units:core.units.length,unitStatuses:core.units.reduce((a,u)=>(a[u.status]=(a[u.status]||0)+1,a),{}),findings:core.findings.length,distinctSites:unique.size,distinctBlockerSites:[...unique.values()].filter(f=>f.kind==='NotYet'||f.kind==='Refused').length,sites:[...unique.values()],relevant};
}
assert(result.baseline.relevant.elementAt.findings.some(f=>f.kind==='NotYet'&&f.reason==='a function returning T | undefined'));
assert.equal(result.candidate.relevant.elementAt.findings.length,0);assert.equal(result.candidate.relevant.toOffset.findings.length,0);
result.f9Gone=true;result.candidateElementAtDeferred=result.candidate.relevant.elementAt.units.some(u=>u.status==='deferred_generic_instantiation');assert(result.candidateElementAtDeferred);
fs.writeFileSync(s+'/census-summary.json',JSON.stringify(result,null,2)+'\n');console.log(JSON.stringify({f9Gone:true,relevant:result.candidate.relevant},null,2));
