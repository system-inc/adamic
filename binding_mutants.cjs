'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict');
const ts=require('/workspace/adamic/stage3/api/node_modules/typescript');
const {compiler,nodes}=require('/workspace/adamic/stage3/adapt/47-host-errors/adapt.cjs');
const {bindings}=require('/workspace/adamic/stage3/adapt/47-host-errors/tracing-fs-import.cjs');
const tree='/workspace/scratch/tracing-after',p=compiler(tree,ts),file=path.join(tree,'src/compiler/tracing.ts'),text=fs.readFileSync(file,'utf8');
function check(text){const h=ts.createCompilerHost(p.getCompilerOptions(),true),get=h.getSourceFile.bind(h);h.getSourceFile=(name,...args)=>path.resolve(name)===file?ts.createSourceFile(name,text,ts.ScriptTarget.Latest,true):get(name,...args);const q=ts.createProgram(p.getRootFileNames(),p.getCompilerOptions(),h);return bindings(ts,q,q.getSourceFile(file),nodes,true);}
const selected=process.argv[2];
const probes=[['module collision',text+'\nconst fs = 1;\n'],['parameter shadow',text.replace('export function startTracing(tracingMode: Mode,','export function startTracing(fs: typeof import("node:fs"), tracingMode: Mode,')],['bare namespace value',text+'\nconst bare = fs;\n']];
if(selected==='clean') console.log(JSON.stringify({qualifiedUses:check(text).qualified}));
else {const [name,changed]=probes[Number(selected)];assert.notEqual(changed,text);let error;try{check(changed);}catch(e){error=e.message;}assert(error,name+' escaped');console.log(JSON.stringify({name,rejected:true,error}));}
