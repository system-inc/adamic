'use strict';
const fs=require('node:fs'), path=require('node:path'), assert=require('node:assert/strict'), crypto=require('node:crypto'), ts=require('typescript');
const [base, candidate, output]=process.argv.slice(2);
const digest=s=>crypto.createHash('sha256').update(s).digest('hex');
const rows=[];
for (const file of ts.sys.readDirectory(path.join(base,'src'), ['.ts']).filter(f=>!f.endsWith('.d.ts')).sort()) {
 const relative=path.relative(base,file), a=fs.readFileSync(file,'utf8'),b=fs.readFileSync(path.join(candidate,relative),'utf8');
 for (const removeComments of [false,true]) {
  const options={compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments}};
  const before=ts.transpileModule(a,options).outputText,after=ts.transpileModule(b,options).outputText;
  if(relative==='src/compiler/sys.ts') {
   assert.equal(before.split('if (sys && sys.getEnvironmentVariable) {').length,2);
   assert.equal(after,before.replace('if (sys && sys.getEnvironmentVariable) {','if (sys) {'));
  } else assert.equal(after,before,relative);
  rows.push({file:relative,removeComments,before:digest(before),after:digest(after),identical:before===after});
 }
}
fs.writeFileSync(output,JSON.stringify({typescript:ts.version,files:rows.length/2,emits:rows.length,rows},null,2)+'\n');
console.log(JSON.stringify({files:rows.length/2,emits:rows.length,changed:rows.filter(r=>!r.identical).length,onlyExpectedCondition:true}));
