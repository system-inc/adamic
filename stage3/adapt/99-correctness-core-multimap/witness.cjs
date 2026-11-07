'use strict';
const fs=require('node:fs'),vm=require('node:vm'),assert=require('node:assert/strict');const bodies=JSON.parse(fs.readFileSync(process.argv[2],'utf8'));
const functions=vm.runInThisContext('(function(){'+bodies.dependencies+';return {oldAdd:('+bodies.add+'),oldRemove:('+bodies.remove+'),freeAdd:('+bodies.freeAdd+'),freeRemove:('+bodies.freeRemove+')};})()');
const value={},other={},cases=[['missing',[],value],['empty',[['k',[]]],value],['singleton-present',[['k',[value]]],value],['singleton-absent',[['k',[other]]],value],['multiple-present',[['k',[value,other,value]]],value],['multiple-absent',[['k',[other,other]]],value],['undefined',[['k',[undefined,other]]],undefined],['NaN',[['k',[NaN,other]]],NaN],['sparse',[['k',[,other]]],undefined]];
const rows=[];
for(const [name,entries,v]of cases)for(const op of ['Add','Remove']){const clone=()=>new Map(entries.map(([k,vs])=>[k,vs.slice()])),a=clone(),b=clone();a.add=functions.oldAdd;a.remove=functions.oldRemove;
 const ar=Reflect.apply(functions['old'+op],a,['k',v]),br=functions['free'+op](b,'k',v);
 assert.deepEqual(ar,br);assert.deepEqual([...a],[...b]);if(op==='Add'){assert.equal(ar,a.get('k'));assert.equal(br,b.get('k'));}
 assert.deepEqual(Object.keys(a),['add','remove']);assert.deepEqual(Object.keys(b),[]);
 rows.push({class:name,operation:op,equalReturnContents:true,equalMapContents:true,addReturnsStoredArray:op==='Add',legacyOwnKeys:Object.keys(a),plainOwnKeys:Object.keys(b)});
}
fs.writeFileSync(process.argv[3],JSON.stringify(rows,null,2)+'\n');console.log(JSON.stringify({witnesses:rows.length,passed:true,knownOwnKeyDifference:['add','remove']}));
