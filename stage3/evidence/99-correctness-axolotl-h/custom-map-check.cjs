'use strict';
const assert=require('node:assert/strict');
const before=require(process.argv[2]),candidate=require(process.argv[3]);
function observe(ts){
    const plugins=[{name:'a',toJSON(){return undefined;}},{name:'b',toJSON(){return undefined;}}];
    plugins.map=function(callback,thisArg){return Array.prototype.map.call(this,callback,thisArg).filter(value=>value!==undefined);};
    return ts.generateTSConfig({plugins},'\n');
}
const a=observe(before),b=observe(candidate);
console.log(JSON.stringify({before:a.split('\n').find(s=>s.includes('"plugins"')),candidate:b.split('\n').find(s=>s.includes('"plugins"'))}));
assert.equal(b,a,'custom map observed changed callback returns');
