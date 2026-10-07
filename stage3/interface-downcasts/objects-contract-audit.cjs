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

const family=process.argv[6]||'objects';assert.ok(['objects','interfaces','object-unions'].includes(family));
const parts=t=>t.isUnion()?t.types:[t];
const callable=t=>parts(t).some(p=>checker.getSignaturesOfType(p,ts.SignatureKind.Call).length || checker.getSignaturesOfType(p,ts.SignatureKind.Construct).length);
const scalar=t=>parts(t).every(p=>!!(p.flags&(ts.TypeFlags.String|ts.TypeFlags.StringLiteral|ts.TypeFlags.Number|ts.TypeFlags.NumberLiteral|ts.TypeFlags.Boolean|ts.TypeFlags.BooleanLiteral))) && (parts(t).every(p=>p.flags&ts.TypeFlags.StringLike) || parts(t).every(p=>p.flags&ts.TypeFlags.NumberLike) || parts(t).every(p=>p.flags&ts.TypeFlags.BooleanLike));
const iface=t=>!!(t.objectFlags&ts.ObjectFlags.Interface) || !!(t.objectFlags&ts.ObjectFlags.Reference && t.target?.objectFlags&ts.ObjectFlags.Interface);
function contract(t,seen=new Set(),descendant=false) {
 if(family==='object-unions' && t.isUnion() && !scalar(t)) {
  if(seen.has(t))return undefined;seen.add(t);
  if(!t.types.every(p=>p.flags&ts.TypeFlags.Object))return 'NotYet: nullish or mixed union';
  const common=checker.getPropertiesOfType(t);
  const finite=p=>{const v=checker.getTypeOfSymbolAtLocation(p,p.valueDeclaration||p.declarations?.[0]);return !(p.flags&ts.SymbolFlags.Optional) && parts(v).every(x=>x.flags&(ts.TypeFlags.StringLiteral|ts.TypeFlags.NumberLiteral|ts.TypeFlags.BooleanLiteral));};
  if(!common.some(finite))return 'NotYet: object union without finite tag';
  for(const member of t.types){const failure=contract(member,seen,true);if(failure)return failure;}
  return undefined;
 }
 if(seen.has(t))return undefined;seen.add(t);
 if(family==='objects' && descendant && iface(t))return 'NotYet: interface descendant';
 if(t.objectFlags & ts.ObjectFlags.Class || t.objectFlags & ts.ObjectFlags.Reference && t.target?.objectFlags & ts.ObjectFlags.Class)return 'NotYet: nominal class field';
 if(checker.getIndexInfosOfType(t).length)return 'NotYet: dictionary view';
 const properties=checker.getPropertiesOfType(t);
 for(const p of properties) {
  const field=checker.getTypeOfSymbolAtLocation(p,p.valueDeclaration||p.declarations?.[0]);
  if(p.flags & ts.SymbolFlags.Optional)return 'NotYet: optional field';
  if(callable(field))return 'Refused: callable members';
  if(scalar(field))continue;
  if(!(field.flags&ts.TypeFlags.Object)) {
   if(family==='object-unions' && field.isUnion()) {const failure=contract(field,seen,true);if(failure)return failure;continue;}
   return 'NotYet: union, nullish, intersection or generic field';
  }
  if(checker.isArrayType(field)||checker.isTupleType(field))return 'NotYet: array field';
  if(!checker.getPropertiesOfType(field).length)return 'NotYet: empty structural object field';
  const failure=contract(field,seen,true);if(failure)return failure;
 }
 return undefined;
}
const lookup=new Map(selected.map(row=>[`${row.file}:${row.start}:${row.end}`,row]));const rows=[];
for(const file of program.getSourceFiles()) {
 const relative=path.relative(root,file.fileName).replaceAll('\\','/');
 function visit(node) {
  if(ts.isAsExpression(node)) {
   const key=`${relative}:${node.getStart(file)}:${node.end}`;const original=lookup.get(key);
   if(original) {
    assert.equal(node.getText(file),original.text);const target=checker.getTypeFromTypeNode(node.type);
    const hasCallable=callable(target)||checker.getPropertiesOfType(target).some(p=>callable(checker.getTypeOfSymbolAtLocation(p,node)));
    const reason=family==='object-unions' && target.isUnion()?'NotYet: union cast target':hasCallable?'Refused: callable members':checker.isArrayType(target)||checker.isTupleType(target)?'NotYet: array view':contract(target)||'target contract eligible';
    rows.push({file:original.file,line:original.line,start:original.start,end:original.end,text:original.text,kind:tagged.has(key)?'tagged':'untagged',target_type:original.target_type,reason});lookup.delete(key);
   }
  }
  ts.forEachChild(node,visit);
 }
 visit(file);
}
assert.equal(lookup.size,0);const counts={};for(const row of rows){const key=row.kind+': '+row.reason;counts[key]=(counts[key]||0)+1;}
const summary={typescript:ts.version,source_commit:'050880ce59e30b356b686bd3144efe24f875ebc8',counts,limits:'Required scalar and '+family+' contract rejection upper bound, not full-source lowering. Arrays, optional/mixed unions and nominal fields remain unsupported; objects-only also excludes interface descendants. Source representation, aliases and writable-slot checks may add reasons. Zero eligible targets proves zero complete sites unlocked; positive eligibility requires actual lowerer probes.'};
fs.writeFileSync(path.join(output,family+'-contract-sites.json'),JSON.stringify(rows,null,2)+'\n');fs.writeFileSync(path.join(output,family+'-contract-summary.json'),JSON.stringify(summary,null,2)+'\n');console.log(JSON.stringify(summary,null,2));
