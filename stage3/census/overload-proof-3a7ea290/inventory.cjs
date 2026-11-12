const fs=require('fs');const ts=require('/workspace/op2-base/stage3/api/node_modules/typescript');
const rows=fs.readFileSync(process.argv[2],'utf8').trim().split('\n').map(JSON.parse),results=[];
for(const r of rows.slice(1,-1)){
 const sf=ts.createSourceFile(r.file,fs.readFileSync(r.file,'utf8'),ts.ScriptTarget.Latest,true);
 const where=n=>{const p=sf.getLineAndCharacterOfPosition(n.getStart(sf));return `${r.file}:${p.line+1}:${p.character+1}`;};
 const expected=sf.statements.map(where),observed=r.units.map(u=>u.where);
 if(JSON.stringify(expected)!==JSON.stringify(observed))throw Error('unit catalog differs: '+r.file);
 results.push({file:r.file,top_level_units:expected.length,all_units_present_in_source_order:true});
}
fs.writeFileSync(process.argv[3],JSON.stringify({files:results,total_units:results.reduce((n,r)=>n+r.top_level_units,0)},null,2)+'\n');
console.log('complete source catalog',results.length,'files');
