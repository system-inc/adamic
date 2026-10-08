// Audit target contracts against the pinned assertions ledger. Eligibility is an
// upper bound on lowering, never a claim that a whole compiler file lowers.
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const crypto = require('node:crypto');
const {execFileSync} = require('node:child_process');
const ts = require('typescript');
const [root, ledgerFile, refinementFile, output] = process.argv.slice(2);
assert.equal(ts.version,'6.0.3');
assert.equal(execFileSync('git',['-C',root,'rev-parse','HEAD'],{encoding:'utf8'}).trim(),'050880ce59e30b356b686bd3144efe24f875ebc8');
const ledger = JSON.parse(fs.readFileSync(ledgerFile));
assert.equal(ledger.length,4101);
const tagged = new Set(fs.readFileSync(refinementFile,'utf8').trim().split('\n').slice(1).map(line=>line.split('\t')).filter(row=>row[5]==='tagged declared base-interface downcast').map(row=>`${row[0]}:${row[3]}:${row[4]}`));
assert.equal(tagged.size,1758);
const selected = ledger.filter(row=>tagged.has(`${row.file}:${row.start}:${row.end}`) || row.category==='structural interface downcast without tag');
assert.equal(selected.length,2936);
const configPath=path.join(root,'src/compiler/tsconfig.json');
const read=ts.readConfigFile(configPath,ts.sys.readFile);assert.equal(read.error,undefined);
const config=ts.parseJsonConfigFileContent(read.config,ts.sys,path.dirname(configPath),undefined,configPath);assert.equal(config.errors.length,0);
const program=ts.createProgram(config.fileNames,config.options);
const checker=program.getTypeChecker();assert.equal(ts.getPreEmitDiagnostics(program).length,0);

const parts=t=>t.isUnion()?t.types:[t];
const location=node=>{const file=node.getSourceFile(); const pos=file.getLineAndCharacterOfPosition(node.getStart(file));return {file:path.relative(root,file.fileName).replaceAll('\\','/'),line:pos.line+1,column:pos.character+1};};
function origin(node, seen=new Set(), depth=0) {
 if(depth>12 || seen.has(node)) return {kind:'cyclic or deep origin',...location(node)};
 seen.add(node);
 if(ts.isParenthesizedExpression(node) || ts.isNonNullExpression(node)) return origin(node.expression,seen,depth+1);
 if(ts.isAsExpression(node) || ts.isTypeAssertionExpression(node)) return {kind:'prior assertion is not a shape certificate',...location(node)};
 if(ts.isObjectLiteralExpression(node)) return {kind:node.properties.some(ts.isSpreadAssignment)?'literal with dynamic spread':'direct record literal',...location(node),node};
 if(ts.isArrayLiteralExpression(node)) return {kind:node.elements.some(ts.isSpreadElement)?'array with dynamic spread':'direct array literal',...location(node),node};
 if(ts.isNewExpression(node)) return {kind:'constructor needs initialized typed-shape certificate',...location(node)};
 if(ts.isIdentifier(node)) {
  const symbol=checker.getSymbolAtLocation(node);
  const declarations=symbol?.declarations||[];
  if(declarations.length===1) {
   const declaration=declarations[0];
   if(ts.isParameter(declaration)) return {kind:'parameter shape-flow set not proven',...location(declaration)};
   if(ts.isVariableDeclaration(declaration) && declaration.initializer) {
    if(ts.isVariableDeclarationList(declaration.parent) && declaration.parent.flags & ts.NodeFlags.Const) return origin(declaration.initializer,seen,depth+1);
    return {kind:'mutable binding needs assignment/points-to analysis',...location(declaration)};
   }
   if(ts.isBindingElement(declaration)) return {kind:'destructured value needs shape-flow analysis',...location(declaration)};
  }
  return {kind:'binding has no allocation certificate',...location(node)};
 }
 if(ts.isPropertyAccessExpression(node) || ts.isElementAccessExpression(node)) return {kind:'property/element shape-flow set not proven',...location(node)};
 if(ts.isCallExpression(node)) {
  const signature=checker.getResolvedSignature(node);const declaration=signature?.declaration;
  return {kind:declaration?.body?'factory/call needs construction and effects summary':'opaque/declared call needs boundary certificate',...location(node)};
 }
 if(ts.isConditionalExpression(node)) return {kind:'conditional needs joined allocation sets',...location(node)};
 return {kind:'expression needs allocation/shape-flow analysis',expression_kind:ts.SyntaxKind[node.kind],...location(node)};
}
function directLiteral(node,target) {
 if(!node || !ts.isObjectLiteralExpression(node)) return undefined;
 // Only complete, direct data initializers are candidates. No asserted initializer,
 // method body, spread, getter, alias or staged initializer is treated as a proof.
 if(node.properties.some(p=>!ts.isPropertyAssignment(p))) return undefined;
 const actual=checker.getTypeAtLocation(node); const fields=[];
 for(const property of checker.getPropertiesOfType(target)) {
  const own=checker.getPropertyOfType(actual,property.name);
  if(!own) {if(!(property.flags & ts.SymbolFlags.Optional)) return {outcome:'never',missing:property.name};continue;}
  const declared=checker.getTypeOfSymbolAtLocation(own,node);
  const expected=checker.getTypeOfSymbolAtLocation(property,node);
  const declaration=own.valueDeclaration;
  if(!declaration || !ts.isPropertyAssignment(declaration) || ts.isAsExpression(declaration.initializer) || ts.isNonNullExpression(declaration.initializer) || declared.flags & (ts.TypeFlags.Any|ts.TypeFlags.Unknown) || checker.getSignaturesOfType(declared,ts.SignatureKind.Call).length) return undefined;
  if(!checker.isTypeAssignableTo(declared,expected)) fields.push({field:property.name,declared_type:checker.typeToString(declared),expected_type:checker.typeToString(expected)});
 }
 // Distinguish a syntactic candidate from a compiler-proven initialized allocation.
 return {outcome:fields.length?'conforms_if':'conforms',fields,requires:'Adamic construction validation, typed shape id, and alias/effect proof'};
}
const lookup=new Map(selected.map(row=>[`${row.file}:${row.start}:${row.end}`,row]));
const rows=[];const boundaries=[];
for(const file of program.getSourceFiles()) {
 const relative=path.relative(root,file.fileName).replaceAll('\\','/');
 if(!relative.startsWith('src/compiler/')) continue;
 function visit(node) {
  if(ts.isCallExpression(node)) {
   const text=node.expression.getText(file);
   if(text==='JSON.parse' || text==='require' || text==='Object.create' || text==='Object.assign') boundaries.push({...location(node),operation:text,text:node.getText(file),result_type:checker.typeToString(checker.getTypeAtLocation(node))});
  }
  if(ts.isAsExpression(node)) {
   const key=`${relative}:${node.getStart(file)}:${node.end}`;const original=lookup.get(key);
   if(original) {
    assert.equal(node.getText(file),original.text);
    const target=checker.getTypeFromTypeNode(node.type);const provenance=origin(node.expression);const candidate=directLiteral(provenance.node,target);delete provenance.node;
    const required=checker.getPropertiesOfType(target).filter(p=>!(p.flags & ts.SymbolFlags.Optional));
    const declaredSource=checker.getTypeAtLocation(node.expression);
    const declarationPreview={required_target_properties:required.length,missing_on_declared_source:[],wider_on_declared_source:[]};
    for(const property of required) {
     const field=checker.getPropertyOfType(declaredSource,property.name);
     if(!field) declarationPreview.missing_on_declared_source.push(property.name);
     else {const actual=checker.getTypeOfSymbolAtLocation(field,node);const expected=checker.getTypeOfSymbolAtLocation(property,node);if(!checker.isTypeAssignableTo(actual,expected)) declarationPreview.wider_on_declared_source.push({field:property.name,source_type:checker.typeToString(actual),target_type:checker.typeToString(expected)});}
    }
    const brands=required.filter(p=>/brand$/i.test(p.name)).map(p=>p.name);
    const anyFields=required.filter(p=>checker.getTypeOfSymbolAtLocation(p,node).flags & ts.TypeFlags.Any).map(p=>p.name);
    const targetArrays=parts(target).some(t=>checker.isArrayType(t)||checker.isTupleType(t));
    rows.push({file:original.file,line:original.line,column:original.column,start:original.start,end:original.end,text:original.text,kind:tagged.has(key)?'tagged':'untagged',source_type:original.source_type,target_type:original.target_type,outcome:'undecidable',reason:candidate?'literal candidate needs compiler typed-shape/construction certificate':provenance.kind,source_origin:provenance,known_flow_shapes:null,declaration_only_preview:declarationPreview,fields_needing_value_checks:null,maximum_check_depth:null,alias_read_checks:'not erased: no per-field initialization and effects certificate',literal_candidate:candidate||null,required_brand_fields:brands,required_any_fields:anyFields,target_array_or_tuple:targetArrays});lookup.delete(key);
   }
  }
  ts.forEachChild(node,visit);
 }
 visit(file);
}
assert.equal(lookup.size,0);assert.equal(rows.length,2936);
const counts={}; const reasons={};
for(const kind of ['tagged','untagged']) {
 const group=rows.filter(r=>r.kind===kind);
 counts[kind]={total:group.length,conforms:0,conforms_if:0,never:0,undecidable:group.length};
 reasons[kind]={};for(const row of group) reasons[kind][row.reason]=(reasons[kind][row.reason]||0)+1;
}
const boundaryCounts={};for(const row of boundaries)boundaryCounts[row.operation]=(boundaryCounts[row.operation]||0)+1;
const summary={typescript:ts.version,source_commit:'050880ce59e30b356b686bd3144efe24f875ebc8',ledger_sha256:crypto.createHash('sha256').update(fs.readFileSync(ledgerFile)).digest('hex'),counts,reasons,boundary_syntax_counts:boundaryCounts,literal_candidates:rows.filter(r=>r.literal_candidate).length,required_brand_sites:{tagged:rows.filter(r=>r.kind==='tagged'&&r.required_brand_fields.length).length,untagged:rows.filter(r=>r.kind==='untagged'&&r.required_brand_fields.length).length},limits:'Conservative evidence audit. No whole-program points-to, typed-layout, initialization, method-body or alias-effect certificate exists for this corpus. Declared interfaces are not treated as actual runtime shapes. Null check-field/depth entries mean unknown, not zero. Candidate literals are not admitted cast proofs.'};
fs.writeFileSync(path.join(output,'shape-conformance-sites.json'),JSON.stringify(rows,null,2)+'\n');
fs.writeFileSync(path.join(output,'shape-conformance-summary.json'),JSON.stringify(summary,null,2)+'\n');
fs.writeFileSync(path.join(output,'shape-conformance-boundaries.json'),JSON.stringify(boundaries,null,2)+'\n');
console.log(JSON.stringify(summary,null,2));
