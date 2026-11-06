'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict');
const file=path.join(process.argv[2],'src/compiler/core.ts'),before=fs.readFileSync(file),needle=Buffer.from('array.length === 0 ? emptyArray as readonly T[] : array.slice().sort(comparer)');
const text=before.toString('utf8');assert.equal(text.split(needle.toString()).length,2);fs.writeFileSync(file,text.replace(needle.toString(),'array.length === 0 ? emptyArray : array.slice().sort(comparer)'));
console.log('Restored exactly the mutable toSorted arm');
