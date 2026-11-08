'use strict';
const ts=require('typescript'),fs=require('fs');
if(ts.version!=='6.0.3')throw Error(ts.version);
const root='/workspace/scratch/sys-timer-slot/main',file=root+'/C-inference.ts';
const signatures={prefix:'...args: [...A, ...unknown[]]',union:'...args: A | [...A, ...unknown[]]',fullTupleNoInfer:'...args: A'};
const result=[];
for(const [spelling,rest] of Object.entries(signatures))for(const ambient of [true,false]){
const source=`export {};\n${ambient?'declare function setTimeout(handler: () => void, timeout: number): NodeJS.Timeout;\ndeclare function clearTimeout(handle: NodeJS.Timeout | undefined): void;':''}
interface Slot {setTimeout?<A extends unknown[]>(callback: (...args: ${spelling==='fullTupleNoInfer'?'NoInfer<A>':'A'}) => void, ms: number, ${rest}): unknown;clearTimeout?(id:unknown):void;}
declare const slot: Slot;
declare const queue: {pollScheduled: unknown};
slot.setTimeout!((_label:string, q:typeof queue)=>{},250,'label',queue);
slot.setTimeout!(()=>{},250,'pollQueue');
slot.setTimeout!(()=>{},1000,'timerToUpdateChildWatches');
const provider: Slot = {setTimeout,clearTimeout};
`;
const options={strict:true,noEmit:true,skipLibCheck:false,target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.NodeNext,moduleResolution:ts.ModuleResolutionKind.NodeNext,types:['node'],typeRoots:[root+'/node_modules/@types'],lib:['lib.es2024.d.ts']};
const host=ts.createCompilerHost(options),read=host.readFile,exists=host.fileExists;host.readFile=f=>f===file?source:read(f);host.fileExists=f=>f===file||exists(f);host.getCurrentDirectory=()=>root;
const p=ts.createProgram([file],options,host),c=p.getTypeChecker(),sf=p.getSourceFile(file),calls=[];
function walk(n){if(ts.isCallExpression(n)){const sig=c.getResolvedSignature(n);calls.push({line:sf.getLineAndCharacterOfPosition(n.getStart()).line+1,text:n.getText(),signature:sig&&c.signatureToString(sig,n,ts.TypeFormatFlags.NoTruncation)});}ts.forEachChild(n,walk);}walk(sf);
const diagnostics=ts.getPreEmitDiagnostics(p).map(d=>({file:d.file?.fileName,line:d.file?d.file.getLineAndCharacterOfPosition(d.start).line+1:undefined,code:d.code,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')}));
result.push({spelling,rest,ambient,source,calls,diagnostics});
}
fs.writeFileSync(process.argv[2],JSON.stringify({typescript:ts.version,attempts:result},null,2)+'\n');console.log(JSON.stringify(result.map(r=>({spelling:r.spelling,ambient:r.ambient,diagnostics:r.diagnostics}))));
