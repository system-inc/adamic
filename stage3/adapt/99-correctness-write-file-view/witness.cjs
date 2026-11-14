'use strict';
// Execute the real pinned owner; only its filesystem helper is a focused stub.
const assert=require('node:assert/strict'),vm=require('node:vm'),ts=require('typescript');
const {owners,parse,transpile,digest}=require('./adapt.cjs');
function run(text,ownCall){
 const events=[],sys={marker:'initial'},f=function(path,data,bom){events.push(['target',this===sys,this.marker,path,data,bom]);};
 const override=label=>function(receiver,...args){events.push([label,receiver===sys,...args]);return Reflect.apply(f,receiver,args);};
 if(ownCall)f.call=override('own-call');
 Object.defineProperty(sys,'writeFile',{configurable:true,get(){events.push(['getter',this===sys]);return f;},set(value){events.push(['setter',this===sys]);Object.defineProperty(sys,'writeFile',{value,writable:true,configurable:true});}});
 sys.directoryExists=function(path){events.push(['directoryExists',this===sys,path]);return false;};
 sys.createDirectory=function(path){events.push(['createDirectory',this===sys,path]);};
 const source=parse('owner.ts',text),captures=[];
 function visit(node){if(ts.isVariableStatement(node)&&node.declarationList.declarations.some(d=>d.name.getText(source)==='originalWriteFile'))captures.push(node);ts.forEachChild(node,visit);}
 visit(source);assert.equal(captures.length,1);const point=captures[0].end;
 // This hook observes the private capture in memory; it is not in the adapter's output.
 const observed=text.slice(0,point)+'\n    witnessCapture(originalWriteFile, sys);'+text.slice(point);
 const context={witnessCapture(value,receiver){assert.equal(value,f);assert.equal(receiver,sys);events.push(['capture',value===f,receiver===sys]);},writeFileEnsuringDirectories(path,data,bom,write,create,exists){events.push(['ensure',path,data,bom]);if(!exists('parent'))create('parent');return write(path,data,bom);}};
 const patch=vm.runInNewContext(transpile(observed).replace(/^export /gm,'')+'\npatchWriteFileEnsuringDirectory;',context);
 patch(sys);const patched=sys.writeFile;patched('first','data',false);
 sys.marker='changed';sys.writeFile=()=>events.push(['later-writer']);if(ownCall)f.call=override('later-own-call');
 patched('second','data2',true);
 assert.equal(events.filter(e=>e[0]==='getter').length,1);assert(!events.some(e=>e[0]==='later-writer'));
 return events;
}
const pin=owners[0];assert.equal(digest(pin.before),pin.ownerBeforeSha256);assert.equal(digest(pin.after),pin.ownerAfterSha256);
const cases=[];for(const ownCall of [false,true]){const before=run(pin.before,ownCall),after=run(pin.after,ownCall);assert.deepEqual(after,before);cases.push({ownCall,before,after});}
console.log(JSON.stringify({owner:pin.owner,source:'whole pinned owners with an in-memory private capture hook',receiverSensitive:true,getterReadOnce:true,laterReassignment:true,identityPreserved:true,ownCallPreserved:true,cases},null,2));
