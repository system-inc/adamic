// Audit target contracts against the pinned assertions ledger. Eligibility is an
// upper bound on lowering, never a claim that a whole compiler file lowers.
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
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
const callable=t=>parts(t).some(p=>checker.getSignaturesOfType(p,ts.SignatureKind.Call).length || checker.getSignaturesOfType(p,ts.SignatureKind.Construct).length);
const scalar=t=>parts(t).every(p=>!!(p.flags&(ts.TypeFlags.String|ts.TypeFlags.StringLiteral|ts.TypeFlags.Number|ts.TypeFlags.NumberLiteral|ts.TypeFlags.Boolean|ts.TypeFlags.BooleanLiteral))) && (parts(t).every(p=>p.flags&ts.TypeFlags.StringLike) || parts(t).every(p=>p.flags&ts.TypeFlags.NumberLike) || parts(t).every(p=>p.flags&ts.TypeFlags.BooleanLike));

function collect(t, blockers, witnesses, seen, fieldPath) {
 const add=family=>{blockers.add(family);witnesses[family]??={field:fieldPath,type:checker.typeToString(t)};};
 if(seen.has(t))return;seen.add(t);
 if(t.flags&ts.TypeFlags.Any){add('any field contract');return;}
 if(t.flags&ts.TypeFlags.Unknown){add('unknown field contract');return;}
 if(t.flags&ts.TypeFlags.TypeParameter){add('generic field contract');return;}
 if(t.flags&ts.TypeFlags.Never){add('never field contract');return;}
 if(t.flags&(ts.TypeFlags.Null|ts.TypeFlags.Undefined))return;
 if(t.isUnion()) {
  if(t.types.some(x=>x.flags&(ts.TypeFlags.Null|ts.TypeFlags.Undefined)))add('nullish members');
  const present=t.types.filter(x=>!(x.flags&(ts.TypeFlags.Null|ts.TypeFlags.Undefined)));
  if(present.some(x=>x.flags&ts.TypeFlags.Object)&&present.some(x=>!(x.flags&ts.TypeFlags.Object)))add('object plus primitive union');
  const primitiveKinds=new Set(present.filter(x=>!(x.flags&ts.TypeFlags.Object)).map(x=>x.flags&ts.TypeFlags.StringLike?'string':x.flags&ts.TypeFlags.NumberLike?'number':x.flags&ts.TypeFlags.BooleanLike?'boolean':'other'));
  if(primitiveKinds.size>1)add('mixed primitive union');
  if(present.length>1&&present.every(x=>x.flags&ts.TypeFlags.Object)) {
   const common=checker.getPropertiesOfType(t);
   const finite=p=>{const v=checker.getTypeOfSymbolAtLocation(p,p.valueDeclaration||p.declarations?.[0]);return !(p.flags&ts.SymbolFlags.Optional)&&parts(v).every(x=>x.flags&(ts.TypeFlags.StringLiteral|ts.TypeFlags.NumberLiteral|ts.TypeFlags.BooleanLiteral));};
   if(!common.some(finite))add('untagged object union');
  }
  for(const member of present)collect(member,blockers,witnesses,seen,fieldPath);return;
 }
 if(scalar(t))return;
 if(checker.isArrayType(t)||checker.isTupleType(t)){add('array or tuple contracts');return;}
 if(callable(t)){add('callable contracts');return;}
 if(t.isIntersection())add('intersection field contract');
 if(t.objectFlags&ts.ObjectFlags.Class || t.objectFlags&ts.ObjectFlags.Reference&&t.target?.objectFlags&ts.ObjectFlags.Class)add('nominal class fields');
 const indexes=checker.getIndexInfosOfType(t);if(indexes.length){add('dictionary contracts');for(const info of indexes)collect(info.type,blockers,witnesses,seen,fieldPath+'[key]');}
 const properties=checker.getPropertiesOfType(t);
 if(!properties.length&&!indexes.length){add('empty or unsupported field contract');return;}
 for(const property of properties){
  if(property.flags&ts.SymbolFlags.Optional){blockers.add('optional properties');witnesses['optional properties']??={field:fieldPath+'.'+property.name,type:checker.typeToString(checker.getTypeOfSymbolAtLocation(property,property.valueDeclaration||property.declarations?.[0]))};}
  collect(checker.getTypeOfSymbolAtLocation(property,property.valueDeclaration||property.declarations?.[0]),blockers,witnesses,seen,fieldPath+'.'+property.name);
 }
}
const lookup=new Map(selected.map(row=>[`${row.file}:${row.start}:${row.end}`,row]));const rows=[];
for(const file of program.getSourceFiles()) {
 const relative=path.relative(root,file.fileName).replaceAll('\\','/');
 function visit(node) {
  if(ts.isAsExpression(node)) {
   const key=`${relative}:${node.getStart(file)}:${node.end}`;const original=lookup.get(key);
   if(original) {
    assert.equal(node.getText(file),original.text);const target=checker.getTypeFromTypeNode(node.type);
    const blockers=new Set();const witnesses={};const seen=new Set();
    collect(target,blockers,witnesses,seen,'<target>');
    if(target.isUnion())blockers.add('union cast admission');
    rows.push({file:original.file,line:original.line,start:original.start,end:original.end,text:original.text,kind:tagged.has(key)?'tagged':'untagged',target_type:original.target_type,blockers:[...blockers].sort(),witnesses});lookup.delete(key);
   }
  }
  ts.forEachChild(node,visit);
 }
 visit(file);
}
assert.equal(lookup.size,0);assert.equal(rows.length,2936);
const counts={};for(const row of rows)for(const family of row.blockers){counts[family]??={tagged:0,untagged:0,total:0};counts[family][row.kind]++;counts[family].total++;}
const ranked=Object.entries(counts).sort((a,b)=>b[1].total-a[1].total);
const summary={typescript:ts.version,source_commit:'050880ce59e30b356b686bd3144efe24f875ebc8',sites:rows.length,counts:Object.fromEntries(ranked),limits:'Overlapping complete-target dependencies, not successful lowering or runtime frequencies. Stop at arrays/callables, owned by lane 2. Generic/intersection/dictionary boundaries are recorded; their field/index contracts are also inspected where available. Every witness is a declared checker type, never a proof of construction or initialization.'};
fs.writeFileSync(path.join(output,'blocking-families-sites.json'),JSON.stringify(rows,null,2)+'\n');fs.writeFileSync(path.join(output,'blocking-families-summary.json'),JSON.stringify(summary,null,2)+'\n');console.log(JSON.stringify(summary,null,2));
