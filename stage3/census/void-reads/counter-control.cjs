'use strict';
const fs=require('node:fs'),path=require('node:path'),cp=require('node:child_process');
const scratch=process.argv[2],output=process.argv[3];fs.mkdirSync(scratch,{recursive:true});
const observer=fs.readFileSync(path.join(__dirname,'observe.cjs'),'utf8');
const source="const object={}; for(const value of [undefined,undefined,false,7,'x',null,object,object]) { if(!Object.is(globalThis.__adamicVoidRead(0,value),value))throw Error('changed result'); }";
function run(name,text){
 const hook=path.join(scratch,name+'.cjs'),directory=path.join(scratch,name);fs.writeFileSync(hook,text);fs.mkdirSync(directory);
 const result=cp.spawnSync('timeout',['10',process.execPath,'--require',hook,'-e',source],{env:{...process.env,VOID_READS_COVERAGE:directory},encoding:'utf8',timeout:15000});
 if(result.status!==0)throw Error(result.stderr);
 const records=fs.readdirSync(directory).map(f=>JSON.parse(fs.readFileSync(path.join(directory,f),'utf8')));
 const site=records[0].sites[0];
 return {hits:site.hits,valueKinds:site.values.map(v=>v.type).sort(),accepted:site.hits===8&&site.values.length===6};
}
const baseline=run('baseline',observer),mutant=run('mutant',observer.replace('record.hits++;',''));
if(!baseline.accepted||mutant.accepted)throw Error('counter mutant escaped');
fs.writeFileSync(output,JSON.stringify({baseline,mutant,mutantCaught:true,mutation:'Remove record.hits++ from the observer.',returnIdentityPreserved:true},null,2)+'\n');
console.log('counter control passed; increment-removal mutant caught');
