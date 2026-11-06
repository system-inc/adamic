'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict');
const {prove}=require('./rename-proof.cjs'),[before,after]=process.argv.slice(2);
const a=fs.readFileSync(path.join(before,'built/local/_tsc.js'),'utf8'),b=fs.readFileSync(path.join(after,'built/local/_tsc.js'),'utf8');
const cases=[
 ['string literal',b+'\nconst observedName = "maybeBind3";\n',/string\/template literal/],
 ['property name',b+'\nconst observedProperty = object.maybeBind3;\n',/property name/],
 ['name read',b.replace('maybeBind3(left)','maybeBind3.name'),/declaration and all two uses|unreviewed use|\.name read/],
 ['unrenamed call',b.replace('maybeBind3(left)','maybeBind2(left)'),/declaration and all two uses/],
];
for(const [name,text,expected] of cases){assert.notEqual(text,b);assert.throws(()=>prove(a,text),expected,name);}
console.log(JSON.stringify({mutants:cases.map(([name])=>name),allRefused:true},null,2));
