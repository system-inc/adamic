const fs=require('node:fs'),assert=require('node:assert/strict'),path=require('node:path');
const a=require('/workspace/scratch/path-mint-618/work/stage3/adapt/99-correctness-timer-closures/adapt.cjs');
const tree=process.argv[2],files=[...new Set(a.owners.map(p=>p.file))];
const snapshot=()=>files.map(f=>{const p=path.join(tree,f);return [f,fs.readFileSync(p,'utf8'),String(fs.statSync(p,{bigint:true}).mtimeNs)]});
const before=snapshot();const result=a.adapt(tree);assert.equal(result.files,0);assert.equal(result.edits,0);assert.deepEqual(snapshot(),before);console.log(JSON.stringify({result,files:files.length,bytesAndMtimesIdentical:true},null,2));
