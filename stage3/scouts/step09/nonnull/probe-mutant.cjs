const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),contract=require('./probe-contract.cjs');
const data=JSON.parse(fs.readFileSync(process.argv[2])),sites=JSON.parse(fs.readFileSync(path.join(__dirname,'sites.json')));
contract(data,sites);
const row=sites.find(s=>s.file==='src/compiler/scanner.ts'&&s.line===4097&&s.column===24);
const key=row.file+':'+row.line+':'+row.column+'@'+row.start+'-'+row.end+':undefined';
const mutant=structuredClone(data);delete mutant.nullish[key];
assert.throws(()=>contract(mutant,sites),/missing nullish observation/);
fs.writeFileSync(path.join(__dirname,'probe-mutant.json'),JSON.stringify({input:path.basename(data.input),mutant:'omit real undefined observation from unconditional table initializer',site:key,caught_by:'independent known-nullish initializer contract',caught:true},null,2)+'\n');
console.log('actual probe-log omission mutant caught by independent unconditional undefined initializer contract');
