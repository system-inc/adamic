'use strict';
const fs=require('fs'),assert=require('assert/strict');
const s='/workspace/scratch/diag-read-view-2',r=s+'/repo',ts=require(r+'/stage3/api/node_modules/typescript');
const {compiler}=require(r+'/stage3/adapt/47-host-errors/adapt.cjs');
const results={};
for(const side of ['main','candidate']){
 const tree=s+'/'+side+'-tree',p=compiler(tree,ts),roots=p.getRootFileNames();assert.equal(roots.length,79);const ds=ts.getPreEmitDiagnostics(p).map(d=>({code:d.code,where:d.file?d.file.fileName.replace(tree+'/', '')+':'+JSON.stringify(d.file.getLineAndCharacterOfPosition(d.start)):null,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')}));results[side]={roots:roots.map(f=>f.replace(tree+'/','')),diagnostics:ds};
}
const key=d=>JSON.stringify([d.code,d.where,d.message]),baseline=new Set(results.main.diagnostics.map(key));results.new=results.candidate.diagnostics.filter(d=>!baseline.has(key(d)));fs.writeFileSync(s+'/raw/stock-tsc.json',JSON.stringify(results,null,2)+'\n');console.log(JSON.stringify({baseline:results.main.diagnostics.length,candidate:results.candidate.diagnostics.length,new:results.new.length}));assert.equal(results.new.length,0);
