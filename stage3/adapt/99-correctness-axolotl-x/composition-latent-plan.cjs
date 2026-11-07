'use strict';
const e=require('./engine.cjs'),{fs,path,assert}=e;
const [tree,out]=process.argv.slice(2),units=require('./owned/latent/unit-map.json').before;
fs.mkdirSync(out,{recursive:true});const rows=[];
for(let i=0;i<units.length;i++){
 const row=units[i],source=e.parse(row.file,fs.readFileSync(path.join(tree,row.file),'utf8')),owner=e.ownerMap(source).get(row.name);assert(owner);
 const pos=source.getLineAndCharacterOfPosition(owner.start),actual=`${row.file}:${pos.line+1}:${pos.character+1}`;
 const plan={schema:1,files:[{file:row.file,all_units:false,units:[actual]}],expected_units:[actual],reference:row.reference,name:row.name};
 fs.writeFileSync(path.join(out,i+'.json'),JSON.stringify(plan,null,2)+'\n');rows.push({...row,actual});
}
fs.writeFileSync(path.join(out,'units.json'),JSON.stringify(rows,null,2)+'\n');
