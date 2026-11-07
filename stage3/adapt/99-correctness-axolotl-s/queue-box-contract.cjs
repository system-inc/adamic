'use strict';
const fs=require('fs'),path=require('path'),assert=require('assert/strict'),ts=require('typescript');
const adapter=require('./adapt.cjs'),q=adapter.owners.find(p=>p.owner==='createQueue'),clear=adapter.owners.find(p=>p.owner==='clear');
const candidate=fs.readFileSync(__dirname+'/queue-box-proposed.ts','utf8');
function tokens(text){const scanner=ts.createScanner(ts.ScriptTarget.Latest,true,ts.LanguageVariant.Standard,text),result=[];for(let token=scanner.scan();token!==ts.SyntaxKind.EndOfFileToken;token=scanner.scan())result.push([token,scanner.getTokenText()]);return result;}
assert.deepEqual(tokens(adapter.transpile(q.after)),tokens(adapter.transpile(candidate)),'measured and applied JavaScript tokens');
assert.equal(q.after.split('\n').length,q.previousAfter.split('\n').length);
function nested(text,name){const source=adapter.parse('core.ts',text),owner=source.statements.find(ts.isFunctionDeclaration);return owner.body.statements.find(n=>ts.isFunctionDeclaration(n)&&n.name?.text===name).getText(source);}
assert.equal(nested(q.previousAfter,'isEmpty'),nested(q.after,'isEmpty'));
function truncation(text){return text.slice(text.indexOf('        // If more than half'),text.indexOf('        return '));}
assert.equal(truncation(q.previousAfter),truncation(q.after));
let casts=0;const source=adapter.parse('core.ts',q.after);function visit(n){if(ts.isAsExpression(n)||ts.isTypeAssertionExpression(n))casts++;ts.forEachChild(n,visit);}visit(source);assert.equal(casts,0);
const raw=fs.readFileSync(path.join(process.argv[2],'src/compiler/core.ts'),'utf8'),dir=fs.mkdtempSync('/workspace/scratch/queue-box/migration-contract-');fs.mkdirSync(path.join(dir,'src/compiler'),{recursive:true});
const file=path.join(dir,'src/compiler/core.ts'),write=fs.writeFileSync;let writes=0;fs.writeFileSync=(...args)=>{writes++;return write(...args);};
const eol=raw.includes('\r\n')?'\r\n':'\n',spell=s=>s.replaceAll('\n',eol);
try{
 write(file,raw);assert.deepEqual(adapter.adapt(dir),{files:1,edits:2,javascriptChanging:true});assert.equal(writes,1);
 const after=fs.readFileSync(file,'utf8'),mtime=fs.statSync(file,{bigint:true}).mtimeNs;assert.equal(after,raw.replace(spell(q.previousAfter),spell(q.after)).replace(spell(clear.previousAfter),spell(clear.after)));
 writes=0;assert.deepEqual(adapter.adapt(dir),{files:0,edits:0,javascriptChanging:true});assert.equal(writes,0);assert.equal(fs.readFileSync(file,'utf8'),after);assert.equal(fs.statSync(file,{bigint:true}).mtimeNs,mtime);
 const fixtures=[raw.replace('elements.splice(newLength);','elements.splice(newLength - 1);'),raw.replace(spell(clear.previousAfter),spell(clear.before)),raw.replace('const elements: (T | undefined)[] = items?.slice() ?? [];','const elements: ({ value: T } | undefined)[] = [];')];
 for(const fixture of fixtures){assert.notEqual(fixture,raw);write(file,fixture);writes=0;assert.throws(()=>adapter.adapt(dir));assert.equal(writes,0);assert.equal(fs.readFileSync(file,'utf8'),fixture);}
 console.log(JSON.stringify({legacyMigration:{files:1,edits:2,writes:1},replay:{files:0,edits:0,writes:0,bytesAndMtimePreserved:true},legacyRefusals:fixtures.length,measuredAndAppliedJavascriptTokensIdentical:true,lineCountPreserved:true,isEmptyUnchanged:true,truncationUnchanged:true,afterCasts:casts},null,2));
}finally{fs.writeFileSync=write;fs.rmSync(dir,{recursive:true,force:true});}
