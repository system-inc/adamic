const fs=require('fs'),path=require('path'),assert=require('assert/strict'),ts=require('/workspace/adamic/stage3/api/node_modules/typescript');
const repo='/workspace/adamic',s='/workspace/scratch/core-finish',source='/workspace/scratch/core-cast/cut-source',tree=s+'/contract-tree';fs.mkdirSync(tree+'/src/compiler',{recursive:true});fs.mkdirSync(tree+'/src/deprecatedCompat',{recursive:true});
for(const rel of ['src/compiler/debug.ts','src/compiler/core.ts','src/deprecatedCompat/deprecate.ts'])fs.copyFileSync(source+'/'+rel,tree+'/'+rel);
const cast=require(repo+'/stage3/adapt/99-correctness-core-cast-message/adapt.cjs');cast.adapt(tree);
const move=require(repo+'/stage3/adapt/99-correctness-core-getfunctionname-move/adapt.cjs');assert.equal(move.adapt(tree).edits,4);
const d=fs.readFileSync(tree+'/src/compiler/debug.ts','utf8'),c=fs.readFileSync(tree+'/src/deprecatedCompat/deprecate.ts','utf8'),oldD=fs.readFileSync(source+'/src/compiler/debug.ts','utf8');const mtime=fs.statSync(tree+'/src/compiler/debug.ts').mtimeMs;assert.equal(move.adapt(tree).edits,0);assert.equal(fs.statSync(tree+'/src/compiler/debug.ts').mtimeMs,mtime);
const reports=[];
for(const dir of ['30-indexed-reads','41-explicit-any-remaining','45-regex-captures','99-correctness-core-debug-messages']){
 const home=require(repo+'/stage3/adapt/'+dir+'/getfunctionname-home.cjs'),stage=dir.startsWith('30')?'30':dir.startsWith('41')?'41':dir.startsWith('45')?'45':'finished';
 assert(home.inspect(ts,d,c,stage).moved);assert(!home.inspect(ts,oldD,fs.readFileSync(source+'/src/deprecatedCompat/deprecate.ts','utf8'),stage).moved);
 for(const [label,debug,local] of [['local drift',d,c.replace('return func.name;','return "drift";')],['local half adapted',d,c.replace('match[1]!','match[1]')],['debug drift',oldD.replace('return func.name;','return "drift";'),fs.readFileSync(source+'/src/deprecatedCompat/deprecate.ts','utf8')],['duplicate homes',oldD,c],['missing homes',d,'']])assert.throws(()=>home.inspect(ts,debug,local,stage),undefined,label);
 reports.push({adapter:dir,bothExactHomesAccepted:true,driftAtBothHomesRefused:true,halfAppliedLocalRefused:true,duplicateAndMissingHomesRefused:true});
}
assert.throws(()=>move.inspect(d.replace(/\r\n/g,'\n'),c.replace('getFunctionName(func), options','Debug.getFunctionName(func), options').replace(/\r\n/g,'\n')),/half-applied/);
const road=require(repo+'/stage3/adapt/99-correctness-core-debug-messages/adapt.cjs');assert.equal(road.adapt(tree).edits,0);assert.throws(()=>road.plan(d,c.replace('return func.name;','return "drift";')),/drift/);
// The real 45 capture proof still runs with the helper at its new home.
for(const rel of ['checker','emitter','parser','semver','sourcemap','utilities','utilitiesPublic','types'])fs.copyFileSync(source+'/src/compiler/'+rel+'.ts',tree+'/src/compiler/'+rel+'.ts');
const fortyFive=require(repo+'/stage3/adapt/45-regex-captures/adapt.cjs');const result=fortyFive.adapt(tree,ts);assert.equal(result.edits,0);assert.equal(fs.readFileSync(tree+'/src/compiler/debug.ts','utf8'),d);
fs.writeFileSync(s+'/raw/contracts.json',JSON.stringify({passed:true,moveEdits:4,idempotentWithoutWrites:true,bodyBytesIdentical:true,strictRecognition:reports,roadBRelocatedGuardRejectsDrift:true,real45ReplayEdits:0},null,2)+'\n');console.log('move, exact bytes, replay, body drift and half-state contracts passed');
