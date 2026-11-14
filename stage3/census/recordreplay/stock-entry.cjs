const ts=require('/workspace/scratch/recordreplay/head/stage3/api/node_modules/typescript');const fs=require('fs'),path=require('path'),crypto=require('crypto');
const tree=process.argv[2],out=process.argv[3];
const cfg=ts.readConfigFile(path.join(tree,'src/tsc/tsconfig.json'),ts.sys.readFile);const parsed=ts.parseJsonConfigFileContent(cfg.config,ts.sys,path.join(tree,'src/tsc'),{noEmit:true,incremental:false,composite:false,emitDeclarationOnly:false,isolatedDeclarations:false});
parsed.options.types=['node'];
parsed.options.typeRoots=['/workspace/scratch/recordreplay/head/stage3/api/node_modules/@types'];
const p=ts.createProgram(parsed.fileNames,parsed.options);const ds=ts.getPreEmitDiagnostics(p);let emits={};
for(const f of p.getSourceFiles().filter(f=>f.fileName.startsWith(tree+'/src/')&&!f.isDeclarationFile)) { const js=ts.transpileModule(f.text,{compilerOptions:{target:ts.ScriptTarget.ES2020,module:ts.ModuleKind.CommonJS,removeComments:false}}).outputText;emits[path.relative(tree,f.fileName)]=crypto.createHash('sha256').update(js).digest('hex'); }
const result={version:ts.version,roots:parsed.fileNames.length,diagnostics:ds.map(d=>({code:d.code,file:d.file&&path.relative(tree,d.file.fileName),start:d.start,text:ts.flattenDiagnosticMessageText(d.messageText,'\n')})),javascript:emits};fs.writeFileSync(out,JSON.stringify(result,null,2)+'\n');console.log('diagnostics',ds.length,'source JS files',Object.keys(emits).length);
