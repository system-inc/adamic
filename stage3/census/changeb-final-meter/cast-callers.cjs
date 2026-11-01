// Source calls are catalogued separately from measured reachability.
const ts=require('/workspace/adamic/stage3/api/node_modules/typescript'),fs=require('fs'),path=require('path');
const root='/workspace/scratch/tree',out=__dirname,raw='/workspace/scratch/changeb-final-meter';
const rows={};
for(const side of ['A','B']){
 const records=fs.readFileSync(raw+'/'+side+'.jsonl','utf8').trim().split('\n').map(JSON.parse);
 const units={};for(const r of records.slice(1))for(const u of r.units)units[u.where.replace(root+'/','')]={...u,findings:r.findings.filter(f=>f.unit===u.where&&f.phase==='lowering')};rows[side]=units;
}
const calls=[];
for(const file of Object.keys(JSON.parse(fs.readFileSync(out+'/source-sha256.json')))){
 const filename=root+'/src/compiler/'+file,sf=ts.createSourceFile(filename,fs.readFileSync(filename,'utf8'),ts.ScriptTarget.Latest,true);
 function loc(n){const p=sf.getLineAndCharacterOfPosition(n.getStart(sf));return 'src/compiler/'+file+':'+(p.line+1)+':'+(p.character+1)}
 for(const top of sf.statements){
  function visit(n){
   if(ts.isCallExpression(n)&&ts.isIdentifier(n.expression)&&['tryCast','cast'].includes(n.expression.text)){
    let fn=n.parent;while(fn&&fn!==top&&!ts.isFunctionLike(fn))fn=fn.parent;
    let name=top.name?.getText(sf)|| (ts.isVariableStatement(top)?top.declarationList.declarations.map(d=>d.name.getText(sf)).join(', '):ts.SyntaxKind[top.kind]);
    const unit_where=loc(top),item={callee:n.expression.text,call_where:loc(n),call_source:n.getText(sf),enclosing_function:fn?.name?.getText(sf)||null,unit_where,unit_name:name};
    for(const side of ['A','B']){const u=rows[side][unit_where];if(!u)throw Error('missing unit '+unit_where);item[side]={status:u.status,instantiated:u.instantiated,outcome:u.status==='attempted'&&!u.findings.length?'lowers':u.findings.length?u.findings[0].kind:u.status,stops:u.findings.map(f=>({where:f.where.replace(root+'/',''),reason:f.reason}))};}
    calls.push(item);
   }ts.forEachChild(n,visit);
  }visit(top);
 }
}
fs.writeFileSync(out+'/cast-callers.json',JSON.stringify(calls,null,2)+'\n');
let lines=['Static direct call sites; containing top-level unit outcomes do not prove that a call was reached. Generic units are deferred and diagnosed bodies are skipped.','','| Callee / call | Predicate argument | Containing unit | A outcome | B outcome / first stop |','|---|---|---|---|---|'];
for(const r of calls)lines.push(`| ${r.callee} ${r.call_where} | ${r.call_source.replace(/\n/g,' ').replace(/\|/g,'&#124;')} | ${r.unit_name} (${r.unit_where}) | ${r.A.outcome} | ${r.B.outcome}: ${r.B.stops.map(s=>s.where+' '+s.reason.replace(/\|/g,'&#124;')).join('; ')} |`);
fs.writeFileSync(out+'/cast-callers.md',lines.join('\n')+'\n');
console.log(JSON.stringify({direct_calls:calls.length,units:new Set(calls.map(c=>c.unit_where)).size,outcomes:calls.reduce((a,r)=>(a[r.B.outcome]=(a[r.B.outcome]||0)+1,a),{})}));
