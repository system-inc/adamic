'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict');
const adapter=require(process.argv[2]),raw=fs.readFileSync(process.argv[3],'utf8'),base=process.argv[4];
fs.mkdirSync(path.join(base,'src/compiler'),{recursive:true});const file=path.join(base,'src/compiler/sys.ts');
function call(input,restoring=false,fail=false) {
 fs.writeFileSync(file,input);const original=fs.writeFileSync;let writes=0;
 fs.writeFileSync=function(name,...args){if(name===file) writes++;return original.call(this,name,...args);};
 let result,error;try{result=adapter.rewrite(base,restoring);}catch(e){error=e;}finally{fs.writeFileSync=original;}
 if(fail){assert(error);assert.equal(writes,0);assert.equal(fs.readFileSync(file,'utf8'),input);}
 else {if(error)throw error;assert.equal(writes,result.files);}
 return {result,text:fs.readFileSync(file,'utf8'),writes};
}
const tests=[];
for(const [ending,text] of [['CRLF',raw],['LF',raw.replace(/\r\n/g,'\n')]]) {
 const a=call(text);assert.equal(a.writes,1);const replay=call(a.text);assert.equal(replay.writes,0);
 const restored=call(a.text,true);assert.equal(restored.text,text);assert.equal(call(text,true).writes,0);tests.push(ending+': adapt / no-write replay / exact mutant / no-write mutant replay');
}
for(const [label,text] of [
 ['owner drift',raw.replace('setCustomPollingValues(sys);','setCustomPollingValues(sys); /* drift */')],
 ['half adapted',raw.replace('if (sys && sys.getEnvironmentVariable) {','if (sys && true) {')],
 ['duplicate',raw+'\r\n'+adapter.owners[0].before.replace(/\n/g,'\r\n')+'\r\n'],
 ['mixed line endings',raw.replace('\r\n','\n')],
 ['mixed before and after',raw+'\r\n'+adapter.owners[0].after.replace(/\n/g,'\r\n')+'\r\n']
]) {call(text,false,true);tests.push(label+': refused with zero writes');}
console.log(JSON.stringify({status:'PASS',tests},null,2));
