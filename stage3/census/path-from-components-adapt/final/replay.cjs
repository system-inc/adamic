const fs=require('fs'),assert=require('assert/strict'),crypto=require('crypto');
const adapter=require('/workspace/path-components-adapt/stage3/adapt/99-correctness-path-from-components/adapt.cjs');
const [tree,out]=process.argv.slice(2),name=tree+'/src/compiler/path.ts',hash=b=>crypto.createHash('sha256').update(b).digest('hex');
const before=fs.readFileSync(name),mtime=fs.statSync(name,{bigint:true}).mtimeNs;let writes=0;const original=fs.writeFileSync;
fs.writeFileSync=function(...args){writes++;return original.apply(this,args)};const result=adapter.adapt(tree);fs.writeFileSync=original;
assert.equal(result.files,0);assert.equal(result.edits,0);assert.equal(writes,0);assert(before.equals(fs.readFileSync(name)));assert.equal(mtime,fs.statSync(name,{bigint:true}).mtimeNs);
fs.writeFileSync(out,JSON.stringify({result,writes,bytesUnchanged:true,mtimeUnchanged:true,sha256:hash(before)},null,2)+'\n');console.log(JSON.stringify(result));
