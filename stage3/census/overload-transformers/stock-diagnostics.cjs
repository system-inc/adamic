'use strict';
const fs=require('fs'),path=require('path'),ts=require('../../api/node_modules/typescript');
const root=path.resolve(process.argv[2]);
const roots=[];function walk(d){for(const e of fs.readdirSync(d,{withFileTypes:true})){const p=path.join(d,e.name);if(e.isDirectory())walk(p);else if(p.endsWith('.ts'))roots.push(p);}}walk(path.join(root,'src/compiler'));roots.sort();
const options={strict:true,noUncheckedIndexedAccess:true,exactOptionalPropertyTypes:true,erasableSyntaxOnly:false,verbatimModuleSyntax:true,allowImportingTsExtensions:true,noEmit:true,module:ts.ModuleKind.ESNext,moduleDetection:ts.ModuleDetectionKind.Force,moduleResolution:ts.ModuleResolutionKind.Bundler,target:ts.ScriptTarget.ES2024,lib:['lib.es2024.d.ts'],types:['node'],typeRoots:[path.resolve(__dirname,'../../api/node_modules/@types')]};
const program=ts.createProgram(roots,options);
const diagnostics=ts.getPreEmitDiagnostics(program).map(d=>({file:d.file?path.relative(root,d.file.fileName):null,line:d.file?d.file.getLineAndCharacterOfPosition(d.start).line+1:null,column:d.file?d.file.getLineAndCharacterOfPosition(d.start).character+1:null,code:d.code,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')}));
console.log(JSON.stringify({typescript:ts.version,options,rootCount:roots.length,roots:roots.map(p=>path.relative(root,p)),diagnostics},null,2));
