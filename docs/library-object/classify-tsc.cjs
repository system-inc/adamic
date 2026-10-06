// Check exact runner programs with the real typescript package, not a syntax approximation.
const fs = require('node:fs');
const ts = require(process.argv[2]);
const captures = fs.readFileSync(process.argv[3], 'utf8').trim().split('\n').map(JSON.parse);
const options = {strict:true,noUncheckedIndexedAccess:true,exactOptionalPropertyTypes:true,noImplicitReturns:true,noFallthroughCasesInSwitch:true,erasableSyntaxOnly:true,verbatimModuleSyntax:true,allowImportingTsExtensions:true,noEmit:true,module:ts.ModuleKind.ESNext,moduleDetection:ts.ModuleDetectionKind.Force,moduleResolution:ts.ModuleResolutionKind.Bundler,target:ts.ScriptTarget.ES2024,lib:['lib.es2024.d.ts'],types:[]};
const filename = '/tmp/object-tsc-program.ts';
const prelude = process.cwd()+'/internal/load/prelude.d.ts';
const cache = new Map();
const baseHost = ts.createCompilerHost(options);
const output = fs.openSync(process.argv[4], 'w');
let checked=0;
for (const capture of captures) {
 if (capture.result.kind !== 'refused') continue;
 const host = {...baseHost, getSourceFile(name,languageVersion,onError) {
  if(name===filename) return ts.createSourceFile(name,capture.program,languageVersion,true);
  if(!cache.has(name)) cache.set(name,baseHost.getSourceFile(name,languageVersion,onError));
  return cache.get(name);
 }};
 const program = ts.createProgram([filename,prelude],options,host);
 const diagnostics=ts.getPreEmitDiagnostics(program).map(d=>({code:d.code,message:ts.flattenDiagnosticMessageText(d.messageText,' '),line:d.file&&d.start!==undefined?d.file.getLineAndCharacterOfPosition(d.start).line+1:null,file:d.file?.fileName}));
 const adamicCode = /error TS(\d+)/.exec(capture.result.reason)?.[1];
 const sameCode=adamicCode!==undefined && diagnostics.some(d=>d.code===Number(adamicCode));
 fs.writeSync(output,JSON.stringify({...capture.result,tsc:diagnostics,sameCode})+'\n');
 if(++checked%100===0) console.error('checked '+checked);
}
fs.closeSync(output);
console.error('TypeScript '+ts.version+' checked '+checked);
