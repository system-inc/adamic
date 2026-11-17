'use strict';
const fs=require('fs'),path=require('path');const ts=require(process.env.TYPESCRIPT_API ? path.resolve(process.env.TYPESCRIPT_API) : path.resolve(__dirname, '../../api/node_modules/typescript'));
const tree=path.resolve(process.argv[2]),out=process.argv[3],core=tree+'/src/compiler/core.ts',text=fs.readFileSync(core,'utf8'),sf=ts.createSourceFile(core,text,ts.ScriptTarget.Latest,true);
const roots=ts.sys.readDirectory(tree+'/src/compiler',['.ts'],undefined,['**/*']);
const options={strict:true,target:ts.ScriptTarget.ES2020,module:ts.ModuleKind.NodeNext,moduleResolution:ts.ModuleResolutionKind.NodeNext,skipLibCheck:true,noEmit:true,types:['node'],typeRoots:[process.env.UPSTREAM_TYPES_ROOT || path.resolve(process.env.ADAMIC_TYPESCRIPT_SOURCE, 'node_modules/@types')],strictBindCallApply:false,useUnknownInCatchVariables:false};
function diagnostics(replacement){
 const host=ts.createCompilerHost(options),read=host.readFile;host.readFile=f=>path.resolve(f)===core?replacement:read(f);
 const p=ts.createProgram(roots,options,host);
 return ts.getPreEmitDiagnostics(p).map(d=>({file:d.file?path.relative(tree,d.file.fileName):null,line:d.file?d.file.getLineAndCharacterOfPosition(d.start).line+1:null,code:d.code,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')}));
}
const baseline=diagnostics(text);fs.writeFileSync(out+'.baseline.json',JSON.stringify(baseline,null,2));
const results={roots:roots.map(f=>path.relative(tree,f)),baselineCount:baseline.length,candidates:{}};
for(const name of ['arrayFrom','arrayToMap','reduceLeft','sameMap']){
 const nodes=sf.statements.filter(n=>ts.isFunctionDeclaration(n)&&n.name?.text===name),edits=[];
 for(const n of nodes){
  if(!n.body&&(name!=='sameMap'||!n.parameters[0].type.getText(sf).startsWith('readonly ')))edits.push({start:n.getFullStart(),end:n.end,after:''});
  if(n.body&&name==='arrayFrom'){const t=n.typeParameters.at(-1);edits.push({start:t.end,end:t.end,after:' = never'});}
  if(n.body&&name==='arrayToMap'){const t=n.typeParameters.at(-1);edits.push({start:t.end,end:t.end,after:' = never'});}
 }
 let next=text;for(const e of edits.sort((a,b)=>b.start-a.start))next=next.slice(0,e.start)+e.after+next.slice(e.end);
 const js=removeComments=>t=>ts.transpileModule(t,{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments}}).outputText;
 const all=diagnostics(next),keys=new Set(baseline.map(d=>JSON.stringify([d.file,d.code,d.message]))),newDiagnostics=all.filter(d=>!keys.has(JSON.stringify([d.file,d.code,d.message])));
 results.candidates[name]={description:name==='sameMap'?'drop mutable overloads, retain readonly overloads':`drop all overloads${name==='reduceLeft'?'':', default unused mapped binder to never'}`,edits,javascriptIdentical:[false,true].every(r=>js(r)(text)===js(r)(next)),diagnosticCount:all.length,newDiagnostics};
 fs.writeFileSync(out,JSON.stringify(results,null,2)+'\n');console.log(name,newDiagnostics.length);
}
