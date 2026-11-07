'use strict';
const fs=require('node:fs'),os=require('node:os'),path=require('node:path'),assert=require('node:assert/strict');
const a=require('./adapt.cjs'),pin=a.owner,base=fs.readFileSync(path.join(process.argv[2],'src/compiler/core.ts'),'utf8');
const dir=fs.mkdtempSync(path.join(os.tmpdir(),'core-assign-contract-'));fs.mkdirSync(path.join(dir,'src/compiler'),{recursive:true});const file=path.join(dir,'src/compiler/core.ts'),write=fs.writeFileSync;let writes=0;fs.writeFileSync=(...args)=>{writes++;return write(...args);};
const nl=base.includes('\r\n')?'\r\n':'\n',spell=s=>s.replace(/\n/g,nl),before=spell(pin.before),afterOwner=spell(pin.after);assert.equal(base.split(before).length,2);
function seed(s){write(file,s);writes=0;}
try{
 seed(base);assert.deepEqual(a.adapt(dir),{files:1,edits:1,javascriptChanging:false});assert.equal(writes,1);const after=fs.readFileSync(file,'utf8');assert.equal(after,base.replace(before,afterOwner));const stat=fs.statSync(file,{bigint:true});writes=0;assert.deepEqual(a.adapt(dir),{files:0,edits:0,javascriptChanging:false});assert.equal(writes,0);assert.equal(fs.statSync(file,{bigint:true}).mtimeNs,stat.mtimeNs);
 const upstream=base.replace(before,spell(pin.upstream));assert.equal(a.plan(file,upstream,'early').result,base);assert.equal(a.plan(file,base,'early').edit,null);assert.equal(a.plan(file,after,'early').edit,null);seed(upstream);assert.equal(a.adapt(dir).edits,1);assert.equal(fs.readFileSync(file,'utf8'),after);
 const mutants=[
  ['changed before body',base.replace('return hasOwnProperty.call(map, key);','return false;')],
  ['changed after body',after.replace('return hasOwnProperty.call(map, key);','return false;')],
  ['parameter only half',base.replace('hasProperty(map: object','hasProperty(map: T')],
  ['type parameter only half',base.replace('hasProperty(map: object','hasProperty<T extends object>(map: object')],
  ['changed generic constraint',after.replace('hasProperty<T extends object>','hasProperty<T extends MapLike<unknown>>')],
  ['changed key type',base.replace('hasProperty(map: object, key: string)','hasProperty(map: object, key: number)')],
  ['changed parameter name',base.replace('hasProperty(map: object','hasProperty(target: object')],
  ['missing owner',base.replace(before,'')],
  ['duplicate owner',base+nl+before],
  ['nested duplicate',base+nl+'function foreignContainer(){'+nl+before.replace('export ','')+nl+'}'],
  ['moved owner',base.replace(before,'function foreignContainer(){'+nl+before.replace('export ','')+nl+'}')],
  ['parse error',base+nl+'function {'],
 ];
 for(const [name,text] of mutants){seed(text);assert.throws(()=>a.adapt(dir),undefined,name);assert.equal(writes,0,name);assert.equal(fs.readFileSync(file,'utf8'),text,name);}
 const foreign=base.replace('export const emptyArray:', '// preserved foreign comment'+nl+'export const emptyArray:');assert.notEqual(foreign,base);seed(foreign);assert.equal(a.adapt(dir).edits,1);assert(fs.readFileSync(file,'utf8').includes('// preserved foreign comment'));
 for(const eol of ['\n','\r\n']){seed(base.replace(/\r\n/g,'\n').replace(/\n/g,eol));assert.equal(a.adapt(dir).edits,1);const text=fs.readFileSync(file,'utf8');assert(eol==='\n'?!text.includes('\r'):!text.replace(/\r\n/g,'').includes('\n'));}
 console.log(JSON.stringify({firstWrites:1,replayWrites:0,replayMtimePreserved:true,earlyPhaseOutputPreserved:true,upstreamAndObjectInputsAccepted:true,refusals:mutants.map(x=>x[0]),foreignEditPreserved:true,lfAndCrlfPreserved:true},null,2));
}finally{fs.writeFileSync=write;fs.rmSync(dir,{recursive:true,force:true});}
