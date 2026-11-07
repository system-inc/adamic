'use strict';
const fs=require('fs'),vm=require('vm'),ts=require('/workspace/adamic/stage3/api/node_modules/typescript'),assert=require('assert/strict');
const pins=JSON.parse(fs.readFileSync('/workspace/scratch/levenshtein/measured-owners.json','utf8'));
function load(text){const js=ts.transpileModule(text+'\nglobalThis.fn=levenshteinWithMax;',{compilerOptions:{target:ts.ScriptTarget.ES2020,module:ts.ModuleKind.None}}).outputText;const ctx={};vm.runInNewContext(js,ctx);return ctx.fn;}
const before=load(pins.before),after=load(pins.after);
const cases=[['equal strings','same','same',3],['empty s1','','abc',3],['empty s2','abc','',3],['case only','a','A',1],['give up','abc','xyz',1],['max below one','a','A',0.2],['zero max equal','a','a',0],['negative max','a','a',-1],['empty both','','',0],['ordinary finite distance','cat','cut',3]];
const countPath='/workspace/scratch/levenshtein/COUNT.json';
if (fs.existsSync(countPath)) {
 const count=JSON.parse(fs.readFileSync(countPath,'utf8'));
 if (count.complete) {
  const wanted=new Set(Object.keys(count.classes).filter(k=>count.classes[k]>0));wanted.add('finiteReturn');const examples=new Map();
  outer:for(const piece of count.pieces){const dir='/workspace/scratch/levenshtein/count-'+String(piece.piece).padStart(2,'0')+'/calls';for(const name of fs.readdirSync(dir).sort()){for(const line of fs.readFileSync(dir+'/'+name,'utf8').trim().split('\n')){if(!line)continue;const r=JSON.parse(line);for(const key of wanted)if(!examples.has(key)&&(key==='finiteReturn'?r.original!=='undefined':r[key]))examples.set(key,r);if(examples.size===wanted.size)break outer;}}}
  assert.equal(examples.size,wanted.size,'represent every observed class');for(const [key,r] of examples)cases.push(['observed '+key,r.s1,r.s2,r.max]);
 }
}
const rows=cases.map(([name,...input])=>{const a=before(...input),b=after(...input);assert(Object.is(a,b),name);return {name,input,before:a===undefined?'undefined':a,after:b===undefined?'undefined':b,equal:true};});
console.log(JSON.stringify({node:process.version,rows},null,2));
