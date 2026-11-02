'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),os=require('node:os');
const {adapt}=require('/workspace/scratch/adamic-99d/stage3/adapt/99-correctness-axolotl-d/adapt.cjs');
const before=process.argv[2],after=process.argv[3];
const relative='src/compiler/commandLineParser.ts';
const old=fs.readFileSync(path.join(before,relative),'utf8'), good=fs.readFileSync(path.join(after,relative),'utf8');
assert.equal(adapt(after).files,0);assert.equal(fs.readFileSync(path.join(after,relative),'utf8'),good);
const results=[];
for(const kind of ['missing helper','changed helper','import only','call only','catch drift','duplicate imports']) {
 const scratch=fs.mkdtempSync(path.join(os.tmpdir(),'99d-contract-'));fs.mkdirSync(path.join(scratch,'src/compiler'),{recursive:true});
 let text=old,helper=fs.readFileSync(path.join(after,'src/compiler/hostErrors.ts'),'utf8');
 if(kind==='changed helper') helper=helper.replace('Reflect.get','Reflect.has');
 if(kind==='import only') text='import { propertyOf } from "./hostErrors.js";\r\n'+old;
 if(kind==='call only') text=old.replace('fileName, e.message);','fileName, String(propertyOf(e, "message")));');
 if(kind==='catch drift') text=old.replace('fileName, e.message);','fileName, e.stack);');
 if(kind==='duplicate imports') text='import { propertyOf } from "./hostErrors.js";\r\nimport { errorCode } from "./hostErrors.js";\r\n'+old;
 fs.writeFileSync(path.join(scratch,relative),text);
 if(kind!=='missing helper') fs.writeFileSync(path.join(scratch,'src/compiler/hostErrors.ts'),helper);
 assert.throws(()=>adapt(scratch));assert.equal(fs.readFileSync(path.join(scratch,relative),'utf8'),text);results.push({kind,refused:true,noWrite:true});fs.rmSync(scratch,{recursive:true});
}
console.log(JSON.stringify({idempotent:true,results},null,2));
