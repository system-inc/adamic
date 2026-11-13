'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto');
const ts=require('/workspace/scratch/sourcefile-like-position/stage3/api/node_modules/typescript');
const [before,after,destination]=process.argv.slice(2);
function emit(tree,side){
 const config=ts.getParsedCommandLineOfConfigFile(path.join(tree,'src/typescript/tsconfig.json'),{}, {...ts.sys,onUnRecoverableConfigFileDiagnostic:d=>{throw Error(ts.flattenDiagnosticMessageText(d.messageText,'\n'));}});
 assert(config); assert.equal(config.errors.length,0);
 const options={...config.options,composite:false,incremental:false,tsBuildInfoFile:undefined,stripInternal:true,declaration:true,emitDeclarationOnly:true,declarationMap:false,sourceMap:false,noEmitOnError:true,outDir:path.join(destination,side)};
 const program=ts.createProgram({rootNames:config.fileNames,options});
 const diagnostics=ts.getPreEmitDiagnostics(program);
 if(diagnostics.length)process.stderr.write(ts.formatDiagnosticsWithColorAndContext(diagnostics,{getCanonicalFileName:f=>f,getCurrentDirectory:()=>tree,getNewLine:()=> '\n'}));
 assert.equal(diagnostics.length,0,side+': declaration diagnostics');
 const result=program.emit();assert.equal(result.emitSkipped,false);assert.equal(result.diagnostics.length,0);
 return {root:options.outDir,sourceFiles:program.getSourceFiles().length};
}
const a=emit(before,'main'),b=emit(after,'trial');
const names=(root,sub='')=>fs.readdirSync(path.join(root,sub),{withFileTypes:true}).flatMap(e=>e.isDirectory()?names(root,path.join(sub,e.name)):[path.join(sub,e.name)]).sort();
const files=names(a.root);assert.deepEqual(files,names(b.root));
const digest=crypto.createHash('sha256');
for(const name of files){const x=fs.readFileSync(path.join(a.root,name)),y=fs.readFileSync(path.join(b.root,name));assert(x.equals(y),name+': declaration bytes changed');digest.update(name+'\n').update(x);}
console.log(JSON.stringify({status:'pass',typescript:ts.version,stripInternal:true,declarations:files.length,sourceFiles:[a.sourceFiles,b.sourceFiles],sha256:digest.digest('hex')},null,2));
