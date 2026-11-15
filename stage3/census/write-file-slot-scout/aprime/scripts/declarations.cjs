const fs=require('fs'),path=require('path'),cp=require('child_process'),crypto=require('crypto'),ts=require('/workspace/scratch/write-file-slot-scout/stage3/api/node_modules/typescript');
const r='/workspace/scratch/write-file-slot-aprime'; const report={};
for(const name of ['candidate']){
const root=path.join(r,name),configName=path.join(root,'src/typescript/tsconfig.json'),loaded=ts.readConfigFile(configName,ts.sys.readFile),config=ts.parseJsonConfigFileContent(loaded.config,ts.sys,path.dirname(configName));
const out=path.join(r,'declarations',name);fs.mkdirSync(out,{recursive:true});
const options={...config.options,composite:false,incremental:false,isolatedDeclarations:false,declaration:true,declarationMap:false,sourceMap:false,emitDeclarationOnly:true,stripInternal:true,noEmit:false,noEmitOnError:false,rootDir:path.join(root,'src'),outDir:out,typeRoots:['/workspace/scratch/write-file-slot-scout/stage3/api/node_modules/@types']};
const program=ts.createProgram(config.fileNames,options);const emitted=program.emit();
const entry=path.join(out,'typescript/typescript.d.ts'),bundle=path.join(out,'typescript.d.ts');
const cmd=['node','/workspace/scratch/write-file-slot-typescript/scripts/dtsBundler.mjs','--entrypoint',entry,'--output',bundle];const bundled=cp.spawnSync(cmd[0],cmd.slice(1),{cwd:'/workspace/scratch/write-file-slot-typescript',encoding:'utf8',maxBuffer:50e6});fs.writeFileSync(out+'/bundler.log',bundled.stdout+bundled.stderr);
report[name]={stripInternal:true,emitSkipped:emitted.emitSkipped,emitDiagnostics:emitted.diagnostics.map(d=>ts.flattenDiagnosticMessageText(d.messageText,'\n')),command:cmd,bundlerExit:bundled.status};
if(fs.existsSync(bundle))report[name].sha256=crypto.createHash('sha256').update(fs.readFileSync(bundle)).digest('hex');console.log(name,JSON.stringify(report[name]));}
fs.writeFileSync(r+'/declarations.json',JSON.stringify(report,null,2)+'\n');
