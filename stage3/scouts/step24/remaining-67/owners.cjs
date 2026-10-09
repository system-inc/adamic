'use strict';
// Observation tooling, run only against externally cached rejected TypeScript.
const fs = require('node:fs'), path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT);
const tree = path.resolve(process.argv[2]);
const unit = __dirname;
const rows = JSON.parse(fs.readFileSync(path.join(unit, 'input-comparison.json'))).rows.filter(r => r.topicProject === 'remaining');
if (rows.length !== 67) throw new Error('expected 67 pinned diagnostics');
const program = ts.createProgram(ts.sys.readDirectory(path.join(tree, 'src/compiler'), ['.ts']), {
 strict: true, exactOptionalPropertyTypes: true, noUncheckedIndexedAccess: true,
 target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext,
 moduleResolution: ts.ModuleResolutionKind.Bundler, types: [], noEmit: true,
});
const checker = program.getTypeChecker();
function at(n, pos) { let result = n; ts.forEachChild(n, c => { if (c.pos <= pos && pos < c.end) result = at(c, pos); }); return result; }
function location(n) { const f=n.getSourceFile(), p=f.getLineAndCharacterOfPosition(n.getStart(f)); return {file:path.relative(tree,f.fileName),line:p.line+1,column:p.character+1,kind:ts.SyntaxKind[n.kind],text:n.getText(f)}; }
function parts(t) {return t.isUnion() ? t.types : [t];}
for (const row of rows) {
 const file=program.getSourceFile(path.join(tree,'src/compiler',row.file));
 const pos=file.getPositionOfLineAndCharacter(row.line-1,row.column-1), node=at(file,pos);
 row.ancestors=[]; row.ownerCandidates=[];
 const name=[...row.message.matchAll(/Types of property '([^']+)' are incompatible/g)][0]?.[1];
 const seen=new Set();
 for(let n=node;n && !ts.isSourceFile(n);n=n.parent) {
  if(row.ancestors.length<6)row.ancestors.push({kind:ts.SyntaxKind[n.kind],text:n.getText(file).slice(0,600)});
  let targets=[];
  if(ts.isExpressionNode(n)){const t=checker.getContextualType(n);if(t){targets.push(t);(row.contextualTargets ||= []).push({expression:n.getText(file).slice(0,80),type:checker.typeToString(t),nullable:parts(t).some(p=>!!(p.flags & (ts.TypeFlags.Undefined|ts.TypeFlags.Null)))});}}
  if(ts.isBinaryExpression(n)&&n.operatorToken.kind===ts.SyntaxKind.EqualsToken)targets.push(checker.getTypeAtLocation(n.left));
  for(const t of targets)for(const part of parts(t)) {
   if(name && ts.isExpressionNode(n)){const actual=checker.getPropertyOfType(checker.getTypeAtLocation(n),name);if(actual){const v=checker.getTypeOfSymbolAtLocation(actual,n);(row.sourceCandidates ||= []).push({expression:n.getText(file).slice(0,100),value:checker.typeToString(v),flags:v.flags,constituents:parts(v).map(p=>({type:checker.typeToString(p),flags:p.flags})),declarations:actual.declarations?.map(location)});}}
   const prop=name ? checker.getPropertyOfType(part,name) : ts.isPropertyAccessExpression(n) ? checker.getSymbolAtLocation(n.name) : undefined;
   for(const d of prop?.declarations||[])if(!seen.has(d)){seen.add(d);row.ownerCandidates.push({...location(d),target:checker.typeToString(part),property:prop.name, symbolFlags:prop.flags, optional:!!(prop.flags & ts.SymbolFlags.Optional), annotation:d.type && checker.typeToString(checker.getTypeFromTypeNode(d.type)), annotationFlags:d.type && checker.getTypeFromTypeNode(d.type).flags});}
  }
 }
 if(row.code===2412) {
  for(let n=node;n;n=n.parent)if(ts.isBinaryExpression(n)&&ts.isPropertyAccessExpression(n.left)) {
   row.receiverType=checker.typeToString(checker.getTypeAtLocation(n.left.expression));
   for(const d of checker.getSymbolAtLocation(n.left.name)?.declarations||[])if(!seen.has(d)){seen.add(d);row.ownerCandidates.push({...location(d),property:n.left.name.text});}
   row.assignment=n.getText(file);break;
  }
 }
}
const families = {
 'checker.ts:1683': 'NodeBuilder', 'checker.ts:54261': 'ModuleSpecifierResolutionHost',
 'checker.ts:54329': 'SymbolTracker', 'checker.ts:6035': 'SymbolVisibilityResult',
 'commandLineParser.ts:2663': 'CompilerOptions', 'commandLineParser.ts:2677': 'ProjectReference',
 'moduleNameResolver.ts:2772': 'Resolved', 'moduleNameResolver.ts:286': 'ResolvedModuleFull',
 'program.ts:1876': 'Program', 'program.ts:2599': 'EmitHost', 'program.ts:472': 'CompilerHost',
 'program.ts:5107': 'ParseConfigFileHost', 'watch.ts:756': 'CompilerHost', 'watch.ts:845': 'ProgramHost',
 'transformers/classFields.ts:2745': 'PrivateEnvironmentData',
};
for (const row of rows) {
 if(row.ownerCandidates.length)continue;
 const family=families[row.file+':'+row.line];
 const name=[...row.message.matchAll(/Types of property '([^']+)' are incompatible/g)][0]?.[1];
 if(!family || !name)continue;
 for(const file of program.getSourceFiles()) {
  if(!file.fileName.startsWith(tree+path.sep+'src'+path.sep))continue;
  function visit(n) {
   if((ts.isInterfaceDeclaration(n)||ts.isTypeAliasDeclaration(n))&&n.name.text===family) {
    const type=checker.getTypeAtLocation(n),prop=checker.getPropertyOfType(type,name);
    for(const d of prop?.declarations||[])row.ownerCandidates.push({...location(d),property:name,lookup:'named diagnostic contract',family});
   }
   ts.forEachChild(n,visit);
  }
  visit(file);
 }
}
fs.writeFileSync(path.join(unit,'owners.json'),JSON.stringify(rows,null,2)+'\n');
console.log(JSON.stringify({rows:rows.length,resolved:rows.filter(r=>r.ownerCandidates.length).length}));
