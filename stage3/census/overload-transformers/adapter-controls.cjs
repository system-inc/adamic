'use strict';
const fs=require('fs'),os=require('os'),path=require('path'),assert=require('node:assert/strict');
const a=require('./rejected-get-properties/adapt.cjs'), pin=a.owners[0];
const root=fs.mkdtempSync(path.join(os.tmpdir(),'overload-controls-'));const file=path.join(root,'src/compiler',pin.file);fs.mkdirSync(path.dirname(file),{recursive:true});
function rejected(text,label){fs.writeFileSync(file,text);assert.throws(()=>a.adapt(root),undefined,label);assert.equal(fs.readFileSync(file,'utf8'),text);}
try {
 for(const ending of ['\n','\r\n']){const raw=pin.before.replace(/\n/g,ending);fs.writeFileSync(file,raw);assert.equal(a.adapt(root).files,1);assert.equal(fs.readFileSync(file,'utf8'),pin.after.replace(/\n/g,ending));assert.equal(a.adapt(root).files,0);assert.equal(a.rewrite(root,true).files,1);assert.equal(fs.readFileSync(file,'utf8'),raw);}
 rejected(pin.before+'\n'+pin.before,'duplicate');
 rejected(pin.before.replace('return filter','return /* drift */ filter'),'body drift');
 rejected(pin.before.replace('requireInitializer: boolean','requireInitializer: false'),'half-adapted');
 rejected(pin.before.replace(/\n/g,'\r\n')+'\n','mixed endings');
 console.log(JSON.stringify({lf:true,crlf:true,replayZeroWrites:true,restoresExact:true,duplicatesRefused:true,driftRefused:true,halfAdaptedRefused:true,mixedEndingsRefused:true,javascriptIdentical:true}));
}finally{fs.rmSync(root,{recursive:true,force:true});}
