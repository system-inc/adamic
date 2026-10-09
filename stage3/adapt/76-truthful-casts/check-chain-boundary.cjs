'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict');
const ts=require(process.env.CENSUS_TYPESCRIPT || 'typescript');
const source=fs.readFileSync(path.join(__dirname,'same-map-chain-boundary.a'),'utf8');
function diagnostics(text) {
 const file=path.join(__dirname,'virtual-chain.ts'),options={strict:true,noUncheckedIndexedAccess:true,noEmit:true,target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext};
 const host=ts.createCompilerHost(options),read=host.getSourceFile.bind(host);
 host.getSourceFile=(f,l,...rest)=>f===file?ts.createSourceFile(f,text,l,true):read(f,l,...rest);
 return ts.getPreEmitDiagnostics(ts.createProgram([file],options,host)).map(d=>({code:d.code,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')}));
}
assert.deepEqual(diagnostics(source),[]);
const rows=[];
for(const site of [564,545,551,555]) {
 const assignment=site===564?'const publicResult: DiagnosticMessageChain[] | undefined = result564;':`const publicResult: DiagnosticMessageChain = result${site};`;
 assert(diagnostics(source+'\n'+assignment).some(d=>d.code===2322),`site ${site} hid the honest union`);
 const sf=ts.createSourceFile('model.a',source,ts.ScriptTarget.Latest,true),name=site===564?'convertOrRepopulateDiagnosticMessageChainArray':`site${site}`;
 const f=sf.statements.find(n=>ts.isFunctionDeclaration(n)&&n.name?.text===name);assert(f?.body&&f.type);
 const edits=[{start:f.type.getStart(sf),end:f.type.end,text:site===564?'DiagnosticMessageChain[] | undefined':'DiagnosticMessageChain'}];
 function visit(n){if(ts.isReturnStatement(n)&&n.expression)edits.push({start:n.expression.getStart(sf),end:n.expression.end,text:`${n.expression.getText(sf)} as unknown as ${site===564?'DiagnosticMessageChain[] | undefined':'DiagnosticMessageChain'}`});else ts.forEachChild(n,visit);}
 visit(f.body);
 let mutant=source;for(const e of edits.sort((a,b)=>b.start-a.start))mutant=mutant.slice(0,e.start)+e.text+mutant.slice(e.end);
 assert.deepEqual(diagnostics(mutant+'\n'+assignment),[],`site ${site} signature mutant was not admitted`);
 for(const removeComments of [false,true]) { const compilerOptions={target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments};assert.equal(ts.transpileModule(mutant,{compilerOptions}).outputText,ts.transpileModule(source,{compilerOptions}).outputText); }
 rows.push({site,truthful_public_assignment:'TS2322',lying_signature_mutant_caught:true,javascript_identical:true});
}
console.log(JSON.stringify({isolated_fixture_typechecks:true,sites:rows}));
