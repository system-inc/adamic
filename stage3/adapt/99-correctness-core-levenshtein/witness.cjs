'use strict';
const fs=require('fs'),vm=require('vm'),ts=require('/workspace/adamic/stage3/api/node_modules/typescript'),assert=require('assert/strict');
const pins=JSON.parse(fs.readFileSync('/workspace/scratch/levenshtein/measured-owners.json','utf8'));
function load(text){const js=ts.transpileModule(text+'\nglobalThis.fn=levenshteinWithMax;',{compilerOptions:{target:ts.ScriptTarget.ES2020,module:ts.ModuleKind.None}}).outputText;const ctx={};vm.runInNewContext(js,ctx);return ctx.fn;}
const before=load(pins.before),after=load(pins.after);
const cases=[['equal strings','same','same',3],['empty s1','','abc',3],['empty s2','abc','',3],['case only','a','A',1],['give up','abc','xyz',1],['max below one','a','A',0.2],['zero max equal','a','a',0],['negative max','a','a',-1],['empty both','','',0],['ordinary finite distance','cat','cut',3]];
const rows=cases.map(([name,...input])=>{const a=before(...input),b=after(...input);assert(Object.is(a,b),name);return {name,input,before:a===undefined?'undefined':a,after:b===undefined?'undefined':b,equal:true};});
console.log(JSON.stringify({node:process.version,rows},null,2));
