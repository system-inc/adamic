'use strict';
const fs=require('node:fs'),path=require('node:path'),vm=require('node:vm');
const ts=require(path.resolve(__dirname,'../../api/node_modules/typescript'));
const dir=path.resolve(__dirname,'counterexamples'),rows=[];
const options={strict:true,alwaysStrict:true,noEmit:true,target:ts.ScriptTarget.ES2020,module:ts.ModuleKind.CommonJS,types:[],skipLibCheck:true};
for(const name of fs.readdirSync(dir).filter(n=>n.endsWith('.ts')).sort()){
 const file=path.join(dir,name),text=fs.readFileSync(file,'utf8'),p=ts.createProgram([file],options);
 const diagnostics=ts.getPreEmitDiagnostics(p).map(d=>({code:d.code,line:d.file?d.file.getLineAndCharacterOfPosition(d.start).line+1:null,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')}));
 const output=[],javascript=ts.transpileModule(text,{compilerOptions:{...options,noEmit:false}}).outputText;
 let runtimeError;try{vm.runInNewContext(javascript,{console:{log:v=>output.push(JSON.parse(v))}},{timeout:1000});}catch(e){runtimeError=String(e);}
 rows.push({file:name,diagnostics,output,runtimeError,badPromiseObserved:output.length===1&&output[0].expectedType!==output[0].actualType});
}
fs.writeFileSync(process.argv[2],JSON.stringify({typescript:ts.version,node:process.version,rows},null,2)+'\n');
console.log(JSON.stringify(rows.map(r=>({file:r.file,diagnostics:r.diagnostics.length,badPromiseObserved:r.badPromiseObserved,runtimeError:r.runtimeError})),null,2));
if(rows.some(r=>r.diagnostics.length||!r.badPromiseObserved||r.runtimeError))process.exitCode=1;
