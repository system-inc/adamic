'use strict';
const fs=require('fs'),assert=require('assert/strict');
const s='/workspace/scratch/diag-read-view-1',r=s+'/repo',tree=s+'/candidate-tree',file=tree+'/src/compiler/utilities.ts';
const ts=require(r+'/stage3/api/node_modules/typescript');
const {compiler}=require(r+'/stage3/adapt/47-host-errors/adapt.cjs');
const base=compiler(tree,ts),text=fs.readFileSync(file,'utf8'),source=base.getSourceFile(file);
const allowed=new Set(['compareDiagnostics','compareDiagnosticsSkipRelatedInformation','compareRelatedInformation','compareMessageText','diagnosticsEqualityComparer','getDiagnosticFilePath','getDiagnosticCode','getDiagnosticMessage']);
const edits=[];
for(const n of source.statements)if(ts.isFunctionDeclaration(n)&&allowed.has(n.name?.text))for(const p of n.parameters)if(p.type?.getText(source)==='Diagnostic')edits.push({start:p.type.getStart(source),end:p.type.end});
let allEight=text;for(const e of edits.sort((a,b)=>b.start-a.start))allEight=allEight.slice(0,e.start)+'DiagnosticReadView'+allEight.slice(e.end);
function inspect(program){
 const sf=program.getSourceFile(file),checker=program.getTypeChecker(),owner=sf.statements.find(n=>ts.isFunctionDeclaration(n)&&n.name?.text==='compareRelatedInformation'),calls=[];
 function visit(n){if(ts.isCallExpression(n)&&ts.isIdentifier(n.expression)&&n.expression.text==='compareDiagnostics')calls.push(n);ts.forEachChild(n,visit)}visit(owner);
 assert.equal(calls.length,1);
 const call=calls[0],signature=checker.getResolvedSignature(call);
 return {
  expression:call.getText(sf),
  argumentTypes:call.arguments.map(n=>checker.typeToString(checker.getTypeAtLocation(n))),
  argumentHasReportsUnnecessary:call.arguments.map(n=>!!checker.getPropertyOfType(checker.getTypeAtLocation(n),'reportsUnnecessary')),
  targetParameterTypes:signature.parameters.map(p=>checker.typeToString(checker.getTypeOfSymbolAtLocation(p,call))),
  targetHasReportsUnnecessary:signature.parameters.map(p=>!!checker.getPropertyOfType(checker.getTypeOfSymbolAtLocation(p,call),'reportsUnnecessary')),
  comparatorParameterType:checker.typeToString(checker.getTypeAtLocation(owner.parameters[0])).replace(/\r\n/g,'\n')
 };
}
const host=ts.createCompilerHost(base.getCompilerOptions()),get=host.getSourceFile.bind(host);
host.getSourceFile=(name,version,...rest)=>name===file?ts.createSourceFile(name,allEight,version,true):get(name,version,...rest);
const virtual=ts.createProgram(base.getRootFileNames(),base.getCompilerOptions(),host);
const current=inspect(base),allAllowedParameters=inspect(virtual);
assert.deepEqual(current.argumentTypes,['DiagnosticRelatedInformation','DiagnosticRelatedInformation']);
assert.deepEqual(current.argumentTypes,allAllowedParameters.argumentTypes);
assert.deepEqual(current.targetParameterTypes,allAllowedParameters.targetParameterTypes);
assert.deepEqual(current.argumentHasReportsUnnecessary,[false,false]);
assert.deepEqual(allAllowedParameters.argumentHasReportsUnnecessary,[false,false]);
assert.deepEqual(current.targetHasReportsUnnecessary,[true,true]);
const result={scope:'read-only stock checker facts explaining the already observed full-census stop; not a replacement census',currentTwoComparatorParameters:current,virtualAllEightComparatorParameters:allAllowedParameters,explanation:'All allowed comparator parameter changes leave both recursive call arguments DiagnosticRelatedInformation and the target DiagnosticReadView. The source lacks reportsUnnecessary and the target retains it through Diagnostic. Removing these changed optional-widening findings requires modifying the field types, interface members, recursive call, or comparator overload shape outside the requested amendment.'};
fs.writeFileSync(s+'/raw/stop-type-facts.json',JSON.stringify(result,null,2)+'\n');
console.log(JSON.stringify(result,null,2));
