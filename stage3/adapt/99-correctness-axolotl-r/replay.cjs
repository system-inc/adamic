'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict');
const adapter=require('./adapt.cjs'),tree=path.resolve(process.argv[2]),file=path.join(tree,'src/compiler/core.ts');
const before=fs.readFileSync(file),mtime=fs.statSync(file,{bigint:true}).mtimeNs,write=fs.writeFileSync;let writes=0;
fs.writeFileSync=(...args)=>{writes++;return write(...args);};
const result=adapter.adapt(tree);assert.deepEqual(result,{files:0,edits:0,javascriptChanging:false});assert.equal(writes,0);
assert(before.equals(fs.readFileSync(file)));assert.equal(fs.statSync(file,{bigint:true}).mtimeNs,mtime);
console.log(JSON.stringify({...result,writes,bytesAndMtimePreserved:true},null,2));
