const inspector = require('node:inspector');
const fs = require('node:fs');
const session = new inspector.Session();session.connect();
const post = (name,args={})=>new Promise((ok,no)=>session.post(name,args,(e,r)=>e?no(e):ok(r)));
(async()=>{
 await post('HeapProfiler.startSampling',{samplingInterval:32768,includeObjectsCollectedByMajorGC:true,includeObjectsCollectedByMinorGC:true});
 const ts=require(process.env.TYPESCRIPT_PACKAGE + '/lib/typescript.js');
 const root=process.env.ADAMIC_REPOSITORY || process.cwd();
 const config=ts.readConfigFile(root+'/tsconfig.json',ts.sys.readFile);
 const options=ts.convertCompilerOptionsFromJson(config.config.compilerOptions,root).options;
 const files=[root+'/internal/load/prelude.d.ts',root+'/'+process.argv[2]];
 const start=process.hrtime.bigint();
 let program=ts.createProgram(files,options);
 const diagnostics=ts.getPreEmitDiagnostics(program);
 if(diagnostics.length){console.error(ts.formatDiagnosticsWithColorAndContext(diagnostics,{getCurrentDirectory:()=>root,getCanonicalFileName:x=>x,getNewLine:()=> '\n'}));process.exitCode=1;return;}
 const {profile}=await post('HeapProfiler.stopSampling');
 let sampled=0;function walk(n){sampled+=n.selfSize;for(const c of n.children)walk(c)}walk(profile.head);
 global.gc();
 const live=process.memoryUsage().heapUsed;
 const count=program.getSourceFiles().length;
 program=undefined;global.gc();
 console.log(JSON.stringify({date:new Date().toISOString(),ts:ts.version,node:process.version,files:count,wallSeconds:Number(process.hrtime.bigint()-start)/1e9,sampledAllocatedBytes:sampled,postGcHeapBytes:live,afterProgramHeapBytes:process.memoryUsage().heapUsed,rssKiB:process.resourceUsage().maxRSS}));
 session.disconnect();
})();
