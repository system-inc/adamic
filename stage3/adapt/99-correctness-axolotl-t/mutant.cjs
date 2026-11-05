'use strict';
const fs=require('fs'),path=require('path'),vm=require('vm'),assert=require('assert/strict');
const parser=require('typescript');
const tree=process.argv[2],mutant=process.argv.includes('--mutant');
const stock=require(path.join(tree,'built/local/typescript.js'));
let text=fs.readFileSync(path.join(tree,'src/compiler/debug.ts'),'utf8');
const check='Debug.assert(hasAntecedentList && (antecedent === undefined || Array.isArray(antecedent)), "Expected label antecedents.");';
assert.equal(text.split(check).length-1,1);
if(mutant)text=text.replace(check,'');
const js=parser.transpileModule(text,{compilerOptions:{target:parser.ScriptTarget.ES2024,module:parser.ModuleKind.CommonJS}}).outputText;
const box={exports:{},require:()=>stock,console};vm.runInNewContext(js,box);
box.exports.Debug.enableDebugInfo();
const bad={flags:stock.FlowFlags.BranchLabel,id:0,antecedent:{*[Symbol.iterator](){}}};
let result;
try{result={returned:box.exports.Debug.formatControlFlowGraph(bad)}}catch(e){result={threw:e.name,message:e.message}}
const detected=result.threw==='Error'&&result.message==='Debug Failure. False expression: Expected label antecedents.';
console.log(JSON.stringify({mutant,status:detected?'green':'red',expected:'reject a label whose antecedent is an iterable object rather than an array',result},null,2));
process.exitCode=detected?0:1;
