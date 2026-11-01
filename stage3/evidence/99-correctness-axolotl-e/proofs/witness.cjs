"use strict";
const assert = require('node:assert/strict');
const path = require('node:path');
const [beforePath,afterPath] = process.argv.slice(2);
const before = require(path.resolve(beforePath));
const after = require(path.resolve(afterPath));
const malformed = 'export class C extends f() extends {}';
const reduced = 'class C extends 0 extends {}';
const ordinary = 'declare function f(): new () => {}; export class C extends f() {}';
function declaration(ts,text) {
 const result=ts.transpileDeclaration(text,{fileName:'input.ts',compilerOptions:{target:ts.ScriptTarget.ESNext}});
 function diagnostic(d) { return {file:d.file?.fileName,start:d.start,length:d.length,category:d.category,code:d.code,message:ts.flattenDiagnosticMessageText(d.messageText,'\n'),related:d.relatedInformation?.map(diagnostic)}; }
 return {...result,diagnostics:result.diagnostics?.map(diagnostic)};
}
assert.throws(()=>declaration(before,malformed),/reading 'kind'/);
const result = declaration(after,malformed);
assert.equal(result.outputText,'declare const C_base: any;\nexport declare class C extends C_base {\n}\nexport {};\n');
const controlBefore = declaration(before,ordinary);
const controlAfter = declaration(after,ordinary);
assert.deepEqual(controlAfter,controlBefore);
function programEmit(ts,text) {
 const options={declaration:true,emitDeclarationOnly:true,noLib:true,target:ts.ScriptTarget.ESNext};
 const file='input.ts'; const host=ts.createCompilerHost(options); const sf=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true);
 host.getSourceFile=name=>name===file?sf:undefined; host.fileExists=name=>name===file; host.readFile=name=>name===file?text:undefined;
 let output=''; host.writeFile=(name,text)=>{if(name.endsWith('.d.ts')) output=text;};
 const p=ts.createProgram([file],options,host); const diagnostics=p.getSemanticDiagnostics().map(d=>({code:d.code,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')}));
 const emit=p.emit(); return {output,diagnostics,emitSkipped:emit.emitSkipped};
}
assert.throws(()=>programEmit(before,malformed),/reading 'kind'/);
const emitted=programEmit(after,malformed);
assert.equal(emitted.output,result.outputText);
assert.deepEqual(programEmit(before,ordinary),programEmit(after,ordinary));
assert.throws(()=>declaration(before,reduced),/reading 'kind'/);
assert.throws(()=>programEmit(before,reduced),/reading 'kind'/);
const reducedDeclaration=declaration(after,reduced);
const reducedProgram=programEmit(after,reduced);
assert.equal(reducedDeclaration.outputText,'declare const C_base: number;\ndeclare class C extends C_base {\n}\n');
assert.equal(reducedProgram.output,reducedDeclaration.outputText);
const numericControl='class C extends 0 {}';
const numericBefore=declaration(before,numericControl);
const numericAfter=declaration(after,numericControl);
assert.deepEqual(numericAfter,numericBefore);
assert.equal(numericAfter.outputText,reducedDeclaration.outputText);
assert.deepEqual(programEmit(before,numericControl),programEmit(after,numericControl));
console.log(JSON.stringify({malformed:result,program:emitted,reduced:reducedDeclaration,reducedProgram,numericControl:numericAfter,ordinary:controlAfter,ordinaryByteIdentical:true},null,2));
