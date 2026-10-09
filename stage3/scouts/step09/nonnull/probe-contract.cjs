// An unconditional stock table initializer is independently known to evaluate undefined.
const assert=require('node:assert/strict');
module.exports=function(data,sites){
 const row=sites.find(s=>s.file==='src/compiler/scanner.ts'&&s.line===4097&&s.column===24);
 assert.ok(row && row.expression==='undefined!' && row.status==='checked');
 const id=row.file+':'+row.line+':'+row.column+'@'+row.start+'-'+row.end;
 assert.equal(data.nullish[id+':undefined'],1,'missing nullish observation of unconditional Script_Extensions initializer');
};
