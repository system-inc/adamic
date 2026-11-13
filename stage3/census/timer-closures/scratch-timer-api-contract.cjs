const ts=require('/workspace/adamic/stage3/api/node_modules/typescript'),path=require('path'),assert=require('assert/strict');
const tree=path.resolve(process.argv[2]);
const file='/workspace/scratch/timer-public-contract.ts';
const text=`import type { WatchHost } from ${JSON.stringify(tree+'/built/local/typescript.js')};
type Timers=Pick<WatchHost,"setTimeout"|"clearTimeout">;
const variadic = {
 setTimeout(callback:(...args:unknown[])=>void, ms:number, ...args:unknown[]):unknown { callback(...args); return ms; },
 clearTimeout(_id:unknown):void {}
};
const host:Timers=variadic;
host.setTimeout!((label?:unknown)=>{void label;},1,{label:1});
host.setTimeout!(()=>{},1,"label","extra");
`;
const options={strict:true,noEmit:true,skipLibCheck:true,target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,moduleResolution:ts.ModuleResolutionKind.Bundler,types:[]};
const host=ts.createCompilerHost(options),original=host.getSourceFile.bind(host);host.getSourceFile=(name,...args)=>name===file?ts.createSourceFile(file,text,99,true):original(name,...args);
const p=ts.createProgram([file],options,host),d=ts.getPreEmitDiagnostics(p).map(d=>({file:d.file?.fileName,line:d.file?.getLineAndCharacterOfPosition(d.start).line+1,code:d.code,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')}));
assert.deepEqual(d.map(d=>[d.file,d.line,d.code]),[[file,9,2554]]);
console.log(JSON.stringify({version:ts.version,variadicImplementerFits:true,optionalLabelCallbackAccepted:true,extraBeyondLabelRejected:d},null,2));
