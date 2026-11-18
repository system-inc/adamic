const fs=require('node:fs');const ts=require('/workspace/depthgate2-base/stage3/api/node_modules/typescript');
const result={typescript:ts.version};
result.packageInputs=[];
for(const [name,version] of [[42,'1.0.0'],['pkg',7]]){
 const pkg={name,version,types:'index.d.ts'},text=JSON.stringify(pkg),filename='/project/node_modules/pkg/package.json';
 const files=new Map([[filename,text],['/project/node_modules/pkg/index.d.ts','export const x: number;']]);const trace=[];
 const host={fileExists:p=>files.has(p),readFile:p=>files.get(p),directoryExists:p=>['/project','/project/node_modules','/project/node_modules/pkg'].includes(p),getCurrentDirectory:()=>'/project',getDirectories:()=>[],trace:s=>trace.push(s),realpath:p=>p,useCaseSensitiveFileNames:true};
 const cache=ts.createModuleResolutionCache('/project',p=>p,{moduleResolution:ts.ModuleResolutionKind.NodeNext});
 const resolved=ts.resolveModuleName('pkg','/project/main.ts',{moduleResolution:ts.ModuleResolutionKind.NodeNext,module:ts.ModuleKind.NodeNext,traceResolution:true},host,cache);
 const cached=cache.getPackageJsonInfoCache().getPackageJsonInfo(filename);
 result.packageInputs.push({input:pkg,readJsonResult:ts.readJson(filename,host),cachedPackageJsonContent:cached?.contents?.packageJsonContent,resolved:resolved.resolvedModule,trace});
}
const chain={messageText:'cached witness',category:1,code:9999,repopulateInfo:42};
const buildInfo={fileNames:['./main.ts'],fileInfos:[{version:'v1',signature:'v1'}],root:[1],options:{incremental:true},semanticDiagnosticsPerFile:[[1,[{category:1,code:9999,file:false,start:0,length:0,messageText:chain}]]],version:ts.version};
const text=JSON.stringify(buildInfo),host={readFile:p=>p==='/project/cache.tsbuildinfo'?text:undefined,getCurrentDirectory:()=>'/project',useCaseSensitiveFileNames:()=>true};
const old=ts.readBuilderProgram({incremental:true,tsBuildInfoFile:'/project/cache.tsbuildinfo'},host);
if(!old)throw Error('build info not accepted');
const diagnostic=old.state.semanticDiagnosticsPerFile.get('/project/main.ts')[0];
result.buildInfoInput={input:buildInfo,readBuilderProgramAccepted:true,retainedDiagnostic:diagnostic,retainedRepopulateInfoType:typeof diagnostic.messageText.repopulateInfo};
if(result.packageInputs[0].readJsonResult.name!==42||result.packageInputs[1].cachedPackageJsonContent.version!==7||diagnostic.messageText.repopulateInfo!==42)throw Error('witness lost');
fs.writeFileSync('/workspace/family5-hazards.json',JSON.stringify(result,null,2)+'\n');console.log('Package JSON and accepted build-info cache preserve wrong-typed hidden fields.');
