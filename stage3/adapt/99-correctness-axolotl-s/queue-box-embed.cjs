'use strict';
// Invoke with the frozen in-memory preload once to materialize its exact wrapper in scratch bundles.
const fs=require('fs'),path=require('path'),crypto=require('crypto'),assert=require('assert/strict'),ts=require('typescript');
const [originalTree,tree]=process.argv.slice(2),hash=s=>crypto.createHash('sha256').update(s).digest('hex');
function span(text,file){const sf=ts.createSourceFile(file,text,99,true,ts.ScriptKind.JS),found=[];function visit(n){if(ts.isFunctionDeclaration(n)&&n.name?.text==='createQueue')found.push(n);ts.forEachChild(n,visit);}visit(sf);assert.equal(found.length,1,file);return {start:found[0].getStart(sf),end:found[0].end};}
const records=[];
for(const relative of fs.readdirSync(path.join(tree,'built/local'),{recursive:true}).filter(n=>n.endsWith('.js'))){
 const file=path.join(tree,'built/local',relative),original=fs.readFileSync(path.join(originalTree,'built/local',relative),'utf8');
 if(!original.includes('function createQueue('))continue;
 const transformed=fs.readFileSync(file,'utf8'),a=span(original,file),b=span(transformed,file);
 assert.notEqual(original,transformed);assert(original.slice(a.start,a.end).includes('items.slice()'));
 const wrapper=transformed.slice(b.start,b.end);assert(wrapper.includes('Missing queue recorder'));assert(wrapper.includes('function originalQueue'));assert(wrapper.includes('function boxedQueue'));
 const compact=ts.transpileModule(wrapper,{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments:true}}).outputText.trim().replace(/\r?\n/g,' ');
 const lines=original.slice(a.start,a.end).split('\n').length-1;
 const after=original.slice(0,a.start)+compact+'\n'.repeat(lines)+original.slice(a.end);
 assert.equal(after.split('\n').length,original.split('\n').length);
 const parsed=ts.createSourceFile(file,after,99,true,ts.ScriptKind.JS);assert.equal(parsed.parseDiagnostics.length,0);
 fs.writeFileSync(file,after);records.push({file:relative,ownerOnly:true,lineCountPreserved:true,before:hash(original),after:hash(after),originalOwner:hash(original.slice(a.start,a.end)),wrapper:hash(compact)});
}
assert.deepEqual(records.map(r=>r.file).sort(),['_tsc.js','run.js','typescript.js']);
console.log(JSON.stringify({measurementOnly:true,originalResultsReturned:true,records},null,2));
