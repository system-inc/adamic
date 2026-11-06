'use strict';
// On a separate copied accepted source tree, restore exactly the original bridge.
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict');
const {owners,parse}=require('./adapt.cjs');
assert.equal(process.argv.length,3,'usage: node mutant.cjs MUTANT_TREE');
const file=path.join(process.argv[2],'src/compiler/core.ts');
const raw=fs.readFileSync(file,'utf8'),text=raw.replace(/\r\n/g,'\n'),pin=owners[0];
assert.equal(text.split(pin.after).length,2,'expected exact adapted owner');
const mutated=text.replace(pin.after,pin.before);parse(file,mutated);
fs.writeFileSync(file,raw.includes('\r\n')?mutated.replace(/\n/g,'\r\n'):mutated);
console.log(JSON.stringify({file,restored:'compareStringsCaseSensitive as any as Comparer<T>',edits:1}));
