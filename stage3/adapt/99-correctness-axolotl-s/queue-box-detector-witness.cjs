'use strict';
const fs=require('fs'),vm=require('vm'),ts=require('typescript'),assert=require('assert/strict');
const file=process.argv[2]+'/built/local/typescript.js',text=fs.readFileSync(file,'utf8'),source=ts.createSourceFile(file,text,99,true,ts.ScriptKind.JS);let owner;
function visit(n){if(ts.isFunctionDeclaration(n)&&n.name?.text==='createQueue')owner=n.getText(source);ts.forEachChild(n,visit);}visit(source);assert(owner);
assert(owner.includes('return box.value;'));
const mutant=owner.replace('return box.value;','return "wrong payload";');assert.notEqual(mutant,owner);
const rows=[],context={__queueBoxRecord:(method,equal,classification)=>rows.push({method,equal,classification})};vm.createContext(context);vm.runInContext(mutant+'\nglobalThis.queue=createQueue;',context);
const queue=context.queue(['first']);assert.throws(()=>queue.dequeue(),/Queue box counterexample: dequeue/);assert.equal(rows.filter(r=>!r.equal).length,1);
console.log(JSON.stringify({seedExcludedFromOracle:true,inputClass:'dense string array, observed by the whole lane',input:['first'],mutantPayload:'wrong payload',detectorCaught:true,rows},null,2));
