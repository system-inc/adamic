const fs=require('fs'),path=require('path'),assert=require('assert/strict'),crypto=require('crypto');
const s=__dirname,before=path.join(s,'main-adapted'),after=path.join(s,'lane-final/adapted-tree'),hash=x=>crypto.createHash('sha256').update(x).digest('hex');
function inventory(root){return fs.readdirSync(path.join(root,'src'),{recursive:true,withFileTypes:true}).filter(e=>e.isFile()).map(e=>path.relative(root,path.join(e.parentPath,e.name))).sort();}
const a=inventory(before),b=inventory(after);assert.deepEqual(a,b);assert.equal(a.length,747);
const entries=a.map(file=>({file,before:hash(fs.readFileSync(path.join(before,file))),after:hash(fs.readFileSync(path.join(after,file)))}));
assert.deepEqual(entries.filter(x=>x.before!==x.after).map(x=>x.file),['src/compiler/sys.ts']);
const sys=path.join(after,'src/compiler/sys.ts');assert.deepEqual(fs.readFileSync(sys),fs.readFileSync(path.join(s,'candidate-sys.ts')));
const old=[fs.readFileSync(sys),String(fs.statSync(sys,{bigint:true}).mtimeNs)],write=fs.writeFileSync;let writes=0;fs.writeFileSync=(...args)=>{writes++;return write(...args)};
let replay;try{replay=require(path.join(s,'work/stage3/adapt/99-wasm-crypto-import/adapt.cjs')).adapt(after);assert.equal(writes,0);assert.equal(replay.files,0);assert.deepEqual([fs.readFileSync(sys),String(fs.statSync(sys,{bigint:true}).mtimeNs)],old)}finally{fs.writeFileSync=write}
fs.writeFileSync(path.join(s,'final-source.json'),JSON.stringify({files:a.length,changed:['src/compiler/sys.ts'],candidateBytesMatch:true,replay,writes,bytesAndMtimePreserved:true,inventory:entries},null,2)+'\n');
console.log(JSON.stringify({files:a.length,changed:['src/compiler/sys.ts'],candidateBytesMatch:true,zeroWriteReplay:true}));
