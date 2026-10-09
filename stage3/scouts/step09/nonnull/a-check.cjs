// Match the fast gate's refusal-header contract, using only this unit's fixtures.
const fs=require('node:fs'),path=require('node:path'),cp=require('node:child_process'),assert=require('node:assert/strict');
const compiler=path.resolve(process.argv[2]),scratch=path.resolve(process.argv[3]);fs.mkdirSync(scratch,{recursive:true});
const results=[];
function accepts(header,r){return header.startsWith('// a-check: refused ')&&r.status===1&&r.stderr.includes('Adamic 0.1 refuses')&&r.stderr.includes(header.slice('// a-check: refused '.length));}
for(const row of JSON.parse(fs.readFileSync(path.join(__dirname,'fixtures.json')))){
 const source=path.join(__dirname,row.file),text=fs.readFileSync(source,'utf8'),header=text.split('\n')[0];
 const original=cp.spawnSync(compiler,['c',source],{encoding:'utf8'});assert.equal(original.error,undefined);assert.equal(original.signal,null);assert.ok(accepts(header,original));
 const changed='// a-check: checked\n'+text.slice(text.indexOf('\n')+1),mutant=path.join(scratch,row.file);fs.writeFileSync(mutant,changed);
 const mutated=cp.spawnSync(compiler,['c',mutant],{encoding:'utf8'});assert.equal(mutated.error,undefined);assert.equal(mutated.signal,null);assert.equal(mutated.status,1);assert.ok(mutated.stderr.includes('refuses the non-null assertion !'));assert.ok(!accepts(changed.split('\n')[0],mutated));
 fs.writeFileSync(path.join(scratch,row.file+'.original.log'),original.stdout+original.stderr);fs.writeFileSync(path.join(scratch,row.file+'.mutant.log'),mutated.stdout+mutated.stderr);
 results.push({file:row.file,expected:header,exit:original.status,diagnostic:original.stderr,passed:true,header_mutant:{header:'// a-check: checked',exit:mutated.status,diagnostic:mutated.stderr,caught:true}});
}
fs.writeFileSync(path.join(__dirname,'a-check-results.json'),JSON.stringify({compiler_revision:'45487a809f89885a3fc651cd590e7dabf31362dc',results},null,2)+'\n');console.log('3 main a-check refusal contracts passed; 3 actual checked-header mutants caught');
