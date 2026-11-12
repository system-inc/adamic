const fs=require('fs');const ts=require('/workspace/latent-whole-passes/stage3/api/node_modules/typescript');
const header=JSON.parse(fs.readFileSync('/workspace/ldpi/entry-shallow/census.jsonl','utf8').split('\n')[0]);
const files=header.sources.map(path=>{const text=fs.readFileSync(path,'utf8');return {path,bytes:Buffer.byteLength(text),units:ts.createSourceFile(path,text,ts.ScriptTarget.Latest,true).statements.length};});
const groups=Array.from({length:6},(_,id)=>({id,weight:0,files:[]}));
for(const f of files.sort((a,b)=>(b.bytes/1024+b.units*2)-(a.bytes/1024+a.units*2))){groups.sort((a,b)=>a.weight-b.weight);groups[0].files.push(f);groups[0].weight+=f.bytes/1024+f.units*2;}
groups.sort((a,b)=>a.id-b.id);fs.writeFileSync('/workspace/ldpi/chunks.json',JSON.stringify({files:files.length,units:files.reduce((s,f)=>s+f.units,0),groups},null,2)+'\n');
