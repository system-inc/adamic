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
const callable=t=>parts(t).some(p=>checker.getSignaturesOfType(p,ts.SignatureKind.Call).length || checker.getSignaturesOfType(p,ts.SignatureKind.Construct).length);
const scalar=t=>parts(t).every(p=>!!(p.flags&(ts.TypeFlags.String|ts.TypeFlags.StringLiteral|ts.TypeFlags.Number|ts.TypeFlags.NumberLiteral|ts.TypeFlags.Boolean|ts.TypeFlags.BooleanLiteral)));
const primitive=t=>parts(t).every(p=>!!(p.flags&ts.TypeFlags.StringLike))?'string':parts(t).every(p=>!!(p.flags&ts.TypeFlags.NumberLike))?'number':parts(t).every(p=>!!(p.flags&ts.TypeFlags.BooleanLike))?'boolean':undefined;
const lookup=new Map(selected.map(row=>[`${row.file}:${row.start}:${row.end}`,row]));
const rows=[];
for(const file of program.getSourceFiles()) {
 const relative=path.relative(root,file.fileName).replaceAll('\\','/');
 function visit(node) {
  if(ts.isAsExpression(node)) {
   const key=`${relative}:${node.getStart(file)}:${node.end}`;const original=lookup.get(key);
   if(original) {
    assert.equal(node.getText(file),original.text);
    const target=checker.getTypeFromTypeNode(node.type);const properties=checker.getPropertiesOfType(target);
    const fields=properties.map(p=>({name:p.name,type:checker.getTypeOfSymbolAtLocation(p,node),optional:!!(p.flags&ts.SymbolFlags.Optional)}));
    const uncheckable=fields.filter(f=>callable(f.type)).map(f=>({field:f.name,type:checker.typeToString(f.type)}));
    const unsupported=fields.filter(f=>f.optional || !scalar(f.type) || !primitive(f.type)).map(f=>({field:f.name,type:checker.typeToString(f.type),reason:f.optional?'optional field':!scalar(f.type)?'object, nullish, intersection or type-parameter field':'mixed primitive union'}));
    const kind=tagged.has(key)?'tagged':'untagged';
    const reason=callable(target)?'Refused: callable target':uncheckable.length?'Refused: callable members':unsupported.length?'NotYet: broader field contracts':kind==='untagged'?'NotYet: untagged admission':'target contract eligible';
    rows.push({...original,kind,reason,uncheckable,unsupported});lookup.delete(key);
   }
  }
  ts.forEachChild(node,visit);
 }
 visit(file);
}
assert.equal(lookup.size,0);
const counters={};for(const row of rows) {const key=`${row.kind}: ${row.reason}`;counters[key]=(counters[key]||0)+1;}
const summary={typescript:ts.version,source_commit:'050880ce59e30b356b686bd3144efe24f875ebc8',ledger_sha256:crypto.createHash('sha256').update(fs.readFileSync(ledgerFile)).digest('hex'),counts:counters,limits:'Target contract audit using the outside checker, not full source lowering. Optional/accessor/conversion aliases may add NotYet reasons. Zero eligible contracts proves zero admitted sites; nonzero eligibility would require actual lowerer probes.'};
fs.writeFileSync(path.join(output,'default-contracts.json'),JSON.stringify(rows,null,2)+'\n');fs.writeFileSync(path.join(output,'default-contract-summary.json'),JSON.stringify(summary,null,2)+'\n');console.log(JSON.stringify(summary,null,2));
