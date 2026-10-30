'use strict';
// Compare the two fully composed trees; undoing the retype must fail this check.
const fs = require('node:fs'), path = require('node:path'), cp = require('node:child_process');
const assert = require('node:assert/strict'), crypto = require('node:crypto'), ts = require('typescript');
const {owner, plan} = require('./classes.cjs');
const rules = require('./class-rules.json').brands;
const [before, after, evidence] = process.argv.slice(2);
assert.equal(ts.version,'6.0.3');
assert(before && after && evidence,'usage: retype-check.cjs BEFORE AFTER EVIDENCE');
const walk = d => fs.readdirSync(d,{withFileTypes:true}).flatMap(e => e.isDirectory()?walk(path.join(d,e.name)):[path.join(d,e.name)]).sort();
const contracts = [...new Set(rules.map(r=>r.file))].flatMap(file => plan(cp.execFileSync('git',['-C',before,'show','HEAD:'+file],{encoding:'utf8',maxBuffer:32*1024*1024}),file).contracts);
function project(text,file) {
 const source=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true), edits=[];
 function visit(n) {
  if ((ts.isPropertySignature(n)||ts.isPropertyDeclaration(n)) && n.type && contracts.some(c=>c.key===n.name.text && c.owner===owner(n))) {
   assert.equal(n.type.kind,ts.SyntaxKind.VoidKeyword,'baseline brand must be void');
   edits.push([n.type.getStart(source),n.type.end]);
  }
  ts.forEachChild(n,visit);
 }
 visit(source);
 for (const [a,b] of edits.sort((a,b)=>b[0]-a[0])) text=text.slice(0,a)+'undefined'+text.slice(b);
 return {text,count:edits.length};
}
const files=walk(path.join(before,'src/compiler')).filter(f=>f.endsWith('.ts')).map(f=>path.relative(before,f));
assert.deepEqual(files,walk(path.join(after,'src/compiler')).filter(f=>f.endsWith('.ts')).map(f=>path.relative(after,f)));
let count=0; const changed=[];
const hash=s=>crypto.createHash('sha256').update(s).digest('hex');
for (const file of files) {
 const old=fs.readFileSync(path.join(before,file),'utf8'), actual=fs.readFileSync(path.join(after,file),'utf8'), expected=project(old,file);
 assert.equal(actual,expected.text,'only reviewed void-to-undefined annotations: '+file);
 count+=expected.count;
 if (old!==actual) changed.push({file,sites:expected.count,before:hash(old),after:hash(actual)});
}
assert.equal(count,36); assert.equal(changed.length,8);
const api='tests/baselines/reference/api/typescript.d.ts';
const original=fs.readFileSync(path.join(before,api),'utf8'), actual=fs.readFileSync(path.join(after,api),'utf8'), expected=project(original,api);
assert.equal(expected.count,27); assert.equal(actual,expected.text,'exact 27 API annotation changes');
const file=changed[0].file, undone=fs.readFileSync(path.join(before,file),'utf8');
assert.throws(()=>assert.equal(undone,project(undone,file).text),/AssertionError/);
fs.mkdirSync(evidence,{recursive:true});
fs.writeFileSync(path.join(evidence,'retype-mutant.json'),JSON.stringify({test:'retype-check.cjs: only reviewed void-to-undefined annotations',mutation:'undo the retype, restoring baseline void annotations',file,result:'red: source differs from independent undefined projection',baselineIsUndoOfRetype:true},null,2)+'\n');
fs.writeFileSync(path.join(evidence,'retype.json'),JSON.stringify({sourceSites:count,changed,publicLines:expected.count,otherCompilerSourcesIdentical:true,apiBefore:hash(original),apiAfter:hash(actual)},null,2)+'\n');
console.log(JSON.stringify({sourceSites:count,publicLines:expected.count,mutant:'killed'}));
