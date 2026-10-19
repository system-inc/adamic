'use strict';
const fs=require('node:fs'),path=require('node:path'),cp=require('node:child_process'),assert=require('node:assert/strict');
const [before,after,scratch,output]=process.argv.slice(2).map(p=>path.resolve(p));
fs.mkdirSync(scratch,{recursive:true});
const file=path.join(after,'src/compiler/core.ts'),original=fs.readFileSync(file,'utf8');
const log=fs.openSync(path.join(scratch,'source-scope.log'),'w');
let result;
try{
 fs.writeFileSync(file,original+'\n// exit79 unrelated-source mutant\n');
 result=cp.spawnSync(process.execPath,[path.join(__dirname,'source-scope.cjs'),before,after,path.join(scratch,'scope.json')],{timeout:120000,stdio:['ignore',log,log]});
 if(result.error)throw result.error;
}finally{fs.writeFileSync(file,original);fs.closeSync(log);}
assert.equal(result.status,1,'unrelated source change survived');
assert.match(fs.readFileSync(path.join(scratch,'source-scope.log'),'utf8'),/src\/compiler\/core.ts/);
const report={status:'pass',mutant:'append comment to unlisted core.ts',caught:true,check:'exact changed source set',checkerExit:result.status};
fs.writeFileSync(output,JSON.stringify(report,null,2)+'\n');console.log(JSON.stringify(report));
