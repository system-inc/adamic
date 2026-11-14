'use strict';
const fs=require('fs'),path=require('path'),crypto=require('crypto');
const ts=require('/workspace/scratch/recordreplay/head/stage3/api/node_modules/typescript');
const tree=process.argv[2],out=process.argv[3];
const p=ts.createProgram([path.join(tree,'src/tsc/tsc.ts')],{strict:true,target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,moduleResolution:ts.ModuleResolutionKind.Bundler,noEmit:true,types:['node'],typeRoots:['/workspace/scratch/recordreplay/head/stage3/api/node_modules/@types']});
const sources=p.getSourceFiles().filter(f=>!f.isDeclarationFile&&f.fileName.startsWith(tree+'/src/'));
fs.writeFileSync(out,JSON.stringify(sources.map(f=>({file:path.relative(tree,f.fileName),sha256:crypto.createHash('sha256').update(f.text).digest('hex'),topLevelStatements:f.statements.length})).sort((a,b)=>a.file.localeCompare(b.file)),null,2)+'\n');
console.log('resolved implementation files',sources.length);
