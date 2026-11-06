'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict');
const {prove}=require('./rename-proof.cjs'),[before,after]=process.argv.slice(2),a=fs.readFileSync(path.join(before,'built/local/_tsc.js'),'utf8'),b=fs.readFileSync(path.join(after,'built/local/_tsc.js'),'utf8');
const cases=[
 ['string literal',b+'\nconst observedName = "_q";\n',/string\/template literal/],
 ['property name',b+'\nconst observedProperty = object._q;\n',/property name/],
 ['name read',b.replace('_q.pop()','_q.name'),/\.name read/],
 ['unrenamed use',b.replace('_q = tracing','_p = tracing'),/declaration and all two uses/],
 ['new temporary escape',b.replace('_p = host.getGlobalTypingsCacheLocation','_p = foreign'),/introduced temporary used only|Expected values/],
];
for(const [name,text,expected] of cases){assert.notEqual(text,b);assert.throws(()=>prove(a,text),expected,name);}
console.log(JSON.stringify({mutants:cases.map(([name])=>name),allRefused:true},null,2));
