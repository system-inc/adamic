const fs=require('node:fs'),path=require('node:path'),cp=require('node:child_process');
const ts=require('/workspace/adamic/stage3/api/node_modules/typescript');
const scratch='/workspace/scratch/ancestor', evidence='/workspace/adamic/stage3/census/for-each-ancestor';
const hash=s=>require('node:crypto').createHash('sha256').update(s).digest('hex');
const base=scratch+'/baseline';
function read(p){return fs.readFileSync(p,'utf8')}
let original=read(base+'/src/compiler/path.ts');
const start=original.indexOf('export function forEachAncestorDirectory<T>(directory: Path');
const impl=original.indexOf('export function forEachAncestorDirectory<T, P extends string>',start);
const end=original.indexOf('\n}',impl)+2;
const owner=original.slice(start,end),oldImpl=original.slice(impl,end);
if(start<0||impl<0||end<impl)throw Error('owner missing');
const stringImpl=oldImpl.replace('<T, P extends string>','<T>').replace('directory: P, callback: (directory: P)','directory: string, callback: (directory: string)').replace('getDirectoryPath(directory) as P','getDirectoryPath(directory)');
const nl=original.includes('\r\n')?'\r\n':'\n';
const pathImpl=oldImpl.replace('forEachAncestorDirectory<T, P extends string>','forEachAncestorDirectoryPath<T>').replaceAll('directory: P','directory: Path').replace('getDirectoryPath(directory) as P','getDirectoryPath(directory)');
const stringOverload=owner.slice(owner.indexOf('\n')+1,impl-start)+stringImpl;
const forms={baseline:owner,A:stringOverload,B:owner.slice(0,impl-start)+stringImpl.replace('callback: (directory: string)','callback: (directory: any)'),C:stringOverload+nl+nl+'/** @internal */'+nl+pathImpl,D:stringImpl};
const result={main:cp.execFileSync('git',['rev-parse','HEAD'],{cwd:'/workspace/adamic',encoding:'utf8'}).trim(),typescript:ts.version,candidates:{}};
for(const [name,form] of Object.entries(forms)){
 const root=name==='baseline'?base:scratch+'/'+name;
 if(name!=='baseline') {fs.rmSync(root,{recursive:true,force:true});fs.cpSync(base,root,{recursive:true,filter:s=>!s.includes('/.git')&&!s.includes('/node_modules')});fs.symlinkSync('/workspace/adamic/stage3/api/node_modules',root+'/node_modules','dir');fs.writeFileSync(root+'/src/compiler/path.ts',original.slice(0,start)+form+original.slice(end));}
 if(name==='C')for(const f of ['src/compiler/watchUtilities.ts','src/typingsInstallerCore/typingsInstaller.ts']){const p=root+'/'+f;fs.writeFileSync(p,read(p).replaceAll('forEachAncestorDirectory','forEachAncestorDirectoryPath'));}
 const configFile=root+'/src/compiler/tsconfig.json'; const r=ts.readConfigFile(configFile,ts.sys.readFile);const config=ts.parseJsonConfigFileContent(r.config,ts.sys,path.dirname(configFile),undefined,configFile);
 const p=ts.createProgram(config.fileNames,{...config.options,typeRoots:[root+'/node_modules/@types'],noEmit:false,noEmitOnError:false,declaration:true,emitDeclarationOnly:false,sourceMap:false,outDir:scratch+'/emit-'+name});
 const diagnostics=[...config.errors,...ts.getPreEmitDiagnostics(p)].map(d=>({code:d.code,file:d.file?.fileName,line:d.file&&d.start!=null?d.file.getLineAndCharacterOfPosition(d.start).line+1:undefined,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')}));
 const outputs={};const write=(file,data)=>{if(file.endsWith('.js')){const rel=path.relative(scratch+'/emit-'+name,file);outputs[rel]={bytes:Buffer.byteLength(data),sha256:hash(data)};fs.mkdirSync(path.dirname(file),{recursive:true});fs.writeFileSync(file,data);}};const emit=p.emit(undefined,write);
 const callerFile=root+'/src/typingsInstallerCore/typingsInstaller.ts';const callerProgram=ts.createProgram([...config.fileNames,callerFile],{...p.getCompilerOptions(),composite:false});const callerDiagnostics=ts.getPreEmitDiagnostics(callerProgram).map(d=>({code:d.code,file:d.file?.fileName,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')}));callerProgram.emit(undefined,write);
 const calls=[];for(const f of [...p.getSourceFiles(),callerProgram.getSourceFile(callerFile)]){if(f.isDeclarationFile)continue;function visit(n){if(ts.isCallExpression(n)&&/forEachAncestorDirectory(Path)?$/.test(n.expression.getText(f))){const checker=f.fileName===callerFile?callerProgram.getTypeChecker():p.getTypeChecker();const sig=checker.getResolvedSignature(n);const owners=[];for(let a=n.parent;a;a=a.parent)if(a.name&&(ts.isFunctionDeclaration(a)||ts.isMethodDeclaration(a)||ts.isClassDeclaration(a)))owners.push(a.name.getText(f));calls.push({owners,file:path.relative(root,f.fileName),line:f.getLineAndCharacterOfPosition(n.getStart()).line+1,expression:n.getText(f),signature:checker.signatureToString(sig)});}ts.forEachChild(n,visit)}visit(f)}
 result.candidates[name]={roots:config.fileNames,diagnostics,callerDiagnostics,emitSkipped:emit.emitSkipped,outputs,calls};fs.writeFileSync(evidence+'/'+name+'-path.ts.txt',read(root+'/src/compiler/path.ts'));console.log(name,config.fileNames.length,diagnostics.length,Object.keys(outputs).length);
 fs.writeFileSync(evidence+'/stock.json',JSON.stringify(result,null,2)+'\n');
}
