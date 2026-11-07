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
const checker=p.getTypeChecker(), sys=p.getSourceFile(root+'/src/compiler/sys.ts');const requireWitness=[];
function witness(n){if(ts.isPropertyAssignment(n)&&n.name.getText(sys)==='module'&&n.initializer.getText(sys)==='require(modulePath)'){const at=sys.getLineAndCharacterOfPosition(n.initializer.getStart(sys));requireWitness.push({expression:n.getText(sys),line:at.line+1,column:at.character+1,valueType:checker.typeToString(checker.getTypeAtLocation(n.initializer)),contextualType:checker.getContextualType(n.initializer)?checker.typeToString(checker.getContextualType(n.initializer)):null});}ts.forEachChild(n,witness);}witness(sys);

const solution=[];
const solutionHost=ts.createSolutionBuilderHost(ts.sys,undefined,d=>solution.push(diag(d)),d=>console.log(ts.flattenDiagnosticMessageText(d.messageText,'\n')),d=>console.log(ts.flattenDiagnosticMessageText(d.messageText,'\n')));
const builder=ts.createSolutionBuilder(solutionHost,[root+'/src/tsconfig.json'],{force:true,verbose:true});
const solutionExit=builder.build();
fs.writeFileSync(out,JSON.stringify({version:ts.version,adamicRoots:roots.length,options,adamic,solution,solutionExit,requireWitness},null,2)+'\n');
console.log(JSON.stringify({adamic:adamic.length,solution:solution.length,solutionExit}));
