// Counts a fixed stock-tsc workload. Static sites and executed calls are different units.
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const ts = require(process.env.CENSUS_TYPESCRIPT);
const output = process.argv[2];
const scratch = process.argv[3];
fs.mkdirSync(scratch,{recursive:true});
const sources={good:'const count: number = 3; const names: string[] = ["tsc"];\n',bad:'const count: number = "wrong"; function needsNumber(value: number): number { return value; } needsNumber("wrong");\n'};
for (const [name,text] of Object.entries(sources)) fs.writeFileSync(path.join(scratch,name+'.ts'),text);
const inputFiles=Object.fromEntries(Object.keys(sources).map(name=>[name,path.join(scratch,name+'.ts')]));
const originals = new Map();
const calls = {};
function wrap(object,key,family) {
 if (typeof object[key] !== 'function') return;
 const original=object[key];
 originals.set(family+'.'+key,{object,key,original});
 object[key]=function(...args) { const label=family+'.'+key; calls[label]=(calls[label]||0)+1; return Reflect.apply(original,this,args); };
}
for(const key of Object.keys(ts.sys)) wrap(ts.sys,key,'ts.sys');
for(const key of ['readFileSync','writeFileSync','statSync','lstatSync','realpathSync','readdirSync','existsSync','openSync','closeSync','readSync']) wrap(fs,key,'fs');
for(const key of ['resolve','normalize','join','dirname','basename','isAbsolute','relative']) wrap(path,key,'path');
for(const key of ['cwd','memoryUsage']) wrap(process,key,'process');
const results=[];
for(const name of ['good','bad']) {
 const before={...calls};
 const file=inputFiles[name];
 const program=ts.createProgram([file],{strict:true,noEmit:true,target:ts.ScriptTarget.ES2020,skipLibCheck:true,types:[]});
 const diagnostics=ts.getPreEmitDiagnostics(program).map(d=>({code:d.code,message:ts.flattenDiagnosticMessageText(d.messageText,'\n'),file:d.file?d.file.fileName.split('/').at(-1):null,line:d.file?d.file.getLineAndCharacterOfPosition(d.start).line+1:null,column:d.file?d.file.getLineAndCharacterOfPosition(d.start).character+1:null}));
 assert.deepEqual(diagnostics,name==='good'?[]:[
  {code:2322,message:"Type 'string' is not assignable to type 'number'.",file:'bad.ts',line:1,column:7},
  {code:2345,message:"Argument of type 'string' is not assignable to parameter of type 'number'.",file:'bad.ts',line:1,column:106}
 ]);
 results.push({name,source:sources[name],diagnostics,calls:Object.fromEntries(Object.entries(calls).map(([k,v])=>[k,v-(before[k]||0)]).filter(([,v])=>v))});
}
for(const {object,key,original} of originals.values()) object[key]=original;
fs.writeFileSync(output,JSON.stringify({typescript:ts.version,node:process.version,workload:'two independent ES2020 strict noEmit programs, cold compiler host per program',results},null,2)+'\n');
console.log('PASS: good [] and bad [2322, 2345]; host call counts recorded');
