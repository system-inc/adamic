'use strict';
const fs=require('node:fs'),vm=require('node:vm'),assert=require('node:assert/strict'),ts=require('typescript');
const source=fs.readFileSync(process.argv[2]+'/src/compiler/core.ts','utf8');
const file=ts.createSourceFile('core.ts',source,ts.ScriptTarget.Latest,true);
const nodes=file.statements.filter(n=>ts.isFunctionDeclaration(n)&&n.name?.text==='concatenate');
const owner=source.slice(nodes[0].getStart(file),nodes.at(-1).end).replaceAll('export function','function');
const wrappers=`
function mutable(a: number[] | undefined, b: number[]) { return concatenate(a, b); }
function readonly(a: readonly number[] | undefined, b: readonly number[]) { return concatenate(a, b); }
`;
function types(text){
 const name='/concatenate-witness.ts',options={strict:true,types:[],target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,moduleResolution:ts.ModuleResolutionKind.Bundler};
 const host=ts.createCompilerHost(options),read=host.readFile,exists=host.fileExists;
 host.readFile=f=>f===name?text:read(f);host.fileExists=f=>f===name||exists(f);
 const p=ts.createProgram([name],options,host),c=p.getTypeChecker(),s=p.getSourceFile(name);
 const diagnostics=ts.getPreEmitDiagnostics(p);assert.equal(diagnostics.length,0,diagnostics.map(d=>ts.flattenDiagnosticMessageText(d.messageText,'\n')).join('\n'));
 return s.statements.filter(n=>ts.isFunctionDeclaration(n)&&['mutable','readonly'].includes(n.name?.text)).map(n=>c.typeToString(c.getReturnTypeOfSignature(c.getSignatureFromDeclaration(n))));
}
const declared=types(owner+wrappers);assert.deepEqual(declared,['number[]','readonly number[]']);
const truthful=owner.replace('array1: T[] | undefined, array2: T[]): T[];','array1: T[] | undefined, array2: T[]): T[] | undefined;').replace('array1: readonly T[] | undefined, array2: readonly T[]): readonly T[];','array1: readonly T[] | undefined, array2: readonly T[]): readonly T[] | undefined;');
const corrected=types(truthful+wrappers);assert.deepEqual(corrected,['number[] | undefined','readonly number[] | undefined']);
const emitted=ts.transpileModule(owner+wrappers,{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,moduleResolution:ts.ModuleResolutionKind.Bundler}}).outputText;
const result=vm.runInNewContext(emitted+'\n[mutable(undefined, []), readonly(undefined, Object.freeze([]))]');
assert.equal(result[0],undefined);assert.equal(result[1],undefined);
console.log(JSON.stringify({node:process.version,typescript:ts.version,mutable:{call:'mutable(undefined, [])',declared:declared[0],actual:'undefined'},readonly:{call:'readonly(undefined, Object.freeze([]))',declared:declared[1],actual:'undefined'},truthfulInferredReturns:corrected},null,2));
