'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict');
const root='/workspace/scratch/binding-views',dir=root+'/repo/stage3/adapt/99-correctness-binding-views';
const {adapt}=require(dir+'/adapt.cjs'),{rewrite}=require(dir+'/engine.cjs');
const pins=require(dir+'/mutant-owners.json'),tree=root+'/contracts';fs.mkdirSync(tree+'/src/compiler',{recursive:true});
function set(program,resolution){for(const [file,text] of [['program.ts',program],['resolutionCache.ts',resolution]])fs.writeFileSync(tree+'/src/compiler/'+file,text);}
const program=pins[1],res=pins[0],rows=[];
function refuse(name,p,r,fn=()=>adapt(tree)){
 set(p,r);let writes=0;const original=fs.writeFileSync;fs.writeFileSync=()=>{writes++;throw Error('unexpected write');};
 try{assert.throws(fn);assert.equal(writes,0);rows.push({name,refused:true,writes});}finally{fs.writeFileSync=original;}
}
set(program.after,res.before);assert.equal(adapt(tree).files,1);
let writes=0;const original=fs.writeFileSync;fs.writeFileSync=()=>{writes++;throw Error('replay wrote');};
try{assert.equal(adapt(tree).files,0);assert.equal(writes,0);rows.push({name:'adapter replay',writes});}finally{fs.writeFileSync=original;}
refuse('owner drift',program.after,res.before.replace('interface ResolutionCache', 'interface ResolutionCacheDrift'));
refuse('outside-owner drift',program.after,res.before+'\n');
refuse('duplicate owner',program.after,'interface ResolutionReadView {}\n'+res.after);
refuse('half-adapted resolution',program.after,res.before.replace('R extends ResolutionWithResolvedFileName','R extends ResolutionReadView'));
refuse('mixed resolution line endings',program.after,res.before.replace('\n','\r\n'));
refuse('mixed program line endings',program.after.replace('\n','\r\n'),res.before);
refuse('missing prior amendment',program.before,res.before);
refuse('half-restored tree',program.before,res.after,()=>rewrite(tree,pins,true));
set(program.after.replace(/\n/g,'\r\n'),res.before.replace(/\n/g,'\r\n'));assert.equal(adapt(tree).files,1);assert(!fs.readFileSync(tree+'/src/compiler/resolutionCache.ts','utf8').replace(/\r\n/g,'').includes('\n'));rows.push({name:'CRLF preserved',pass:true});
set(program.after,res.after);assert.equal(rewrite(tree,pins,true).files,2);
assert.equal(fs.readFileSync(tree+'/src/compiler/program.ts','utf8'),program.before);assert.equal(fs.readFileSync(tree+'/src/compiler/resolutionCache.ts','utf8'),res.before);
assert.equal(rewrite(tree,pins,true).files,0);rows.push({name:'mutant exact restoration and replay',pass:true});
fs.writeFileSync(root+'/evidence/contracts.json',JSON.stringify({pass:true,rows},null,2)+'\n');console.log(JSON.stringify(rows));
