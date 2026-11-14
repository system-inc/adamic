'use strict';
const fs=require('node:fs'),vm=require('node:vm'),crypto=require('node:crypto'),assert=require('node:assert/strict'),ts=require('typescript');
const file=process.argv[2]||'/tmp/env-presence-1965/baseline/src/compiler/sys.ts';
const text=fs.readFileSync(file,'utf8').replace(/\r\n/g,'\n');
const source=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true);
assert.equal(source.parseDiagnostics.length,0);
const helper=source.statements.filter(n=>ts.isFunctionDeclaration(n)&&n.name?.text==='setCustomPollingValues');
const block=source.statements.filter(n=>ts.isIfStatement(n)&&n.expression.getText(source)==='sys && sys.getEnvironmentVariable');
assert.equal(helper.length,1);assert.equal(block.length,1);
const before=block[0].getText(source),after=before.replace('if (sys && sys.getEnvironmentVariable) {','if (sys) {');
function run(label,changed) {
    const calls=[];let system;
    if(label==='present') system={getEnvironmentVariable(name){assert.equal(this,system);calls.push(name);return name==='NODE_ENV'?'development':'';}};
    else if(label==='hypothetical missing member')system={};
    const context={sys:system,__entered:false,defaultChunkLevels:{Low:32,Medium:64,High:256},pollingChunkSize:{sentinel:'chunks'},unchangedPollThresholds:{sentinel:'thresholds'},createPollingIntervalBasedLevels(){throw new Error('unexpected custom-level factory call for empty polling environment');},Debug:{setAssertionLevel(level){calls.push({assertionLevel:level});}},AssertionLevel:{Normal:1,None:0}};
    // Keep the real helper and block; the inserted marker observes block entry only.
    const statement=(changed?after:before).replace('{','{\n__entered=true;');
    const code=ts.transpileModule(helper[0].getText(source)+'\n'+statement,{compilerOptions:{target:ts.ScriptTarget.ES2024}}).outputText;
    let error;try{vm.runInNewContext(code,context);}catch(e){error={name:e.name,message:e.message};}
    return {entered:context.__entered,calls,...(error?{error}:{})};
}
console.log(JSON.stringify({source:'actual adapted sys.ts helper and top-level block',blockSha256:crypto.createHash('sha256').update(before).digest('hex')}));
for(const label of ['present','undefined','hypothetical missing member']) {
    const before=run(label,false),after=run(label,true);
    if(label!=='hypothetical missing member'){assert.deepEqual(before,after);assert.equal(before.error,undefined);}
    else {assert.equal(before.entered,false);assert.equal(after.entered,true);assert.equal(after.error.name,'TypeError');}
    console.log(JSON.stringify({label,before,after}));
}
