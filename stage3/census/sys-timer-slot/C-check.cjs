'use strict';
const ts=require('typescript'),fs=require('fs'),path=require('path');
const [root,repo,out]=process.argv.slice(2);if(ts.version!=='6.0.3')throw Error(ts.version);
function diag(d){let r={code:d.code,category:ts.DiagnosticCategory[d.category],message:ts.flattenDiagnosticMessageText(d.messageText,'\n')};if(d.file){let l=d.file.getLineAndCharacterOfPosition(d.start);Object.assign(r,{file:path.relative(root,d.file.fileName),line:l.line+1,column:l.character+1});}return r;}
const roots=ts.sys.readDirectory(root+'/src/compiler',['.ts'],undefined,['**/*']).sort();
const options={strict:true,noUncheckedIndexedAccess:true,exactOptionalPropertyTypes:true,erasableSyntaxOnly:false,verbatimModuleSyntax:true,allowImportingTsExtensions:true,noEmit:true,module:ts.ModuleKind.ESNext,moduleDetection:ts.ModuleDetectionKind.Force,moduleResolution:ts.ModuleResolutionKind.Bundler,target:ts.ScriptTarget.ES2024,lib:['lib.es2024.d.ts'],types:[],typeRoots:[repo+'/internal/load/node_types/node_modules/@types']};
// Reproduce the Node-enabled loader prelude: console is supplied by Node.
const prelude=fs.readFileSync(repo+'/internal/load/prelude.d.ts','utf8').replace(/declare const console: \{[\s\S]*?\};/, 'declare var console: Console;');
const preludeFile=root+'/adamic-measurement-prelude.d.ts';
const host=ts.createCompilerHost(options),read=host.readFile,exists=host.fileExists;host.readFile=f=>{if(f===preludeFile)return prelude;let text=read(f);if(!text)return text;
if(/lib\.(es5|es2021\.string|es2015\.symbol\.wellknown)\.d\.ts$/.test(f)){text=text.replaceAll('replacer: (substring: string, ...args: any[]) => string','replacer: (substring: string, ...args: any[]) => unknown');}if(f.endsWith('/lib.es5.d.ts')){text=text.replaceAll('interface RegExpExecArray extends Array<string>','interface RegExpExecArray extends Array<string | undefined>').replaceAll('interface RegExpMatchArray extends Array<string>','interface RegExpMatchArray extends Array<string | undefined>').replaceAll('split(separator: string | RegExp, limit?: number): string[];','split(separator: string, limit?: number): string[];\n    split(separator: RegExp, limit?: number): (string | undefined)[];');}
if(f.endsWith('/lib.es2018.regexp.d.ts'))text=text.replaceAll('[key: string]: string','[key: string]: string | undefined');
if(f.endsWith('/lib.es2015.symbol.wellknown.d.ts'))text=text.replaceAll('[Symbol.split](string: string, limit?: number): string[]','[Symbol.split](string: string, limit?: number): (string | undefined)[]').replaceAll('split(splitter: { [Symbol.split](string: string, limit?: number): (string | undefined)[]; }, limit?: number): string[];','split(splitter: { [Symbol.split](string: string, limit?: number): (string | undefined)[]; }, limit?: number): (string | undefined)[];');return text;};host.fileExists=f=>f===preludeFile||exists(f);
const node=repo+'/internal/load/node_types/node_modules/@types/node/index.d.ts';
const p=ts.createProgram([...roots,preludeFile,node],options,host);
const adamic=ts.getPreEmitDiagnostics(p).map(diag);
const headProbes=[],checker=p.getTypeChecker(),sys=p.getSourceFile(root+'/src/compiler/sys.ts');
const delta=root.endsWith('-b')?2:0;
function inspect(n){if(ts.isShorthandPropertyAssignment(n)&&['setTimeout','clearTimeout'].includes(n.name.text)){
 const point=sys.getLineAndCharacterOfPosition(n.getStart()),logicalLine=point.line+1+delta;
 if([1501,1502,1599,1600].includes(logicalLine)){
  const context=checker.getContextualType(n.parent),targetProp=context?.getProperty(n.name.text),source=checker.getTypeAtLocation(n);
  const target=targetProp&&checker.getNonNullableType(checker.getTypeOfSymbolAtLocation(targetProp,n));
  headProbes.push({line:point.line+1,mainLine:logicalLine,name:n.name.text,sourceType:checker.typeToString(source,n,ts.TypeFormatFlags.NoTruncation),targetType:target&&checker.typeToString(target,n,ts.TypeFormatFlags.NoTruncation),fits:target&&checker.isTypeAssignableTo(source,target),sourceSignatures:source.getCallSignatures().map(sig=>({signature:checker.signatureToString(sig,n,ts.TypeFormatFlags.NoTruncation),file:sig.declaration?.getSourceFile().fileName,line:sig.declaration?sig.declaration.getSourceFile().getLineAndCharacterOfPosition(sig.declaration.getStart()).line+1:undefined}))});
 }
}ts.forEachChild(n,inspect);}inspect(sys);if(headProbes.length!==4)throw Error('expected four provider probes');

const solution=[];
const solutionHost=ts.createSolutionBuilderHost(ts.sys,undefined,d=>solution.push(diag(d)),d=>console.log(ts.flattenDiagnosticMessageText(d.messageText,'\n')),d=>console.log(ts.flattenDiagnosticMessageText(d.messageText,'\n')));
const builder=ts.createSolutionBuilder(solutionHost,[root+'/src/tsconfig.json'],{force:true,verbose:true});
const solutionExit=builder.build();
fs.writeFileSync(out,JSON.stringify({version:ts.version,adamicRoots:roots.length,options,adamic,solution,solutionExit,headProbes},null,2)+'\n');
console.log(JSON.stringify({adamic:adamic.length,solution:solution.length,solutionExit}));
