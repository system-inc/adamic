// Audit checked-view contracts against the existing ledger; this is not a runtime benchmark.
// Usage: NODE_PATH=<pinned API> node view-inventory.cjs <checkout> <ledger> <other TSV> <output>
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const crypto = require('node:crypto');
const {execFileSync} = require('node:child_process');
const ts = require('typescript');
const [root, ledgerFile, refinementFile, output] = process.argv.slice(2);
assert.equal(ts.version, '6.0.3');
assert.equal(execFileSync('git',['-C',root,'rev-parse','HEAD'],{encoding:'utf8'}).trim(),'050880ce59e30b356b686bd3144efe24f875ebc8');
const ledger = JSON.parse(fs.readFileSync(ledgerFile));
const tagged = new Set(fs.readFileSync(refinementFile,'utf8').trim().split('\n').slice(1).map(line => line.split('\t')).filter(row => row[5] === 'tagged declared base-interface downcast').map(row => `${row[0]}:${row[3]}:${row[4]}`));
assert.equal(tagged.size,1758);
const selected = ledger.filter(row => row.category === 'structural interface downcast without tag' || tagged.has(`${row.file}:${row.start}:${row.end}`));
assert.equal(selected.length,2936);
const configPath = path.join(root,'src/compiler/tsconfig.json');
const read = ts.readConfigFile(configPath,ts.sys.readFile);
assert.equal(read.error,undefined);
const config = ts.parseJsonConfigFileContent(read.config,ts.sys,path.dirname(configPath),undefined,configPath);
assert.equal(config.errors.length,0);
const program = ts.createProgram(config.fileNames,config.options);
const checker = program.getTypeChecker();
assert.equal(ts.getPreEmitDiagnostics(program).length,0);
const parts = type => type.isUnion() ? type.types : [type];
const callable = type => parts(type).some(part => checker.getSignaturesOfType(part,ts.SignatureKind.Call).length > 0 || checker.getSignaturesOfType(part,ts.SignatureKind.Construct).length > 0);
const counters = {};
const rows = [];
const byFile = new Map();
for (const row of selected) {
 if (!byFile.has(row.file)) byFile.set(row.file,[]);
 byFile.get(row.file).push(row);
}
const visited = new Set();
for (const [relative, sites] of byFile) {
 const file = program.getSourceFile(path.join(root,relative));
 const lookup = new Map(sites.map(row => [`${row.start}:${row.end}`,row]));
 function visit(node) {
  if (ts.isAsExpression(node)) {
   const key = `${node.getStart(file)}:${node.end}`;
   const original = lookup.get(key);
   if (original) {
    assert.equal(node.getText(file),original.text);
    const target = checker.getTypeFromTypeNode(node.type);
    const source = checker.getTypeAtLocation(node.expression);
    const kind = tagged.has(`${relative}:${key}`) ? 'tagged' : 'untagged';
    const fields = [];
    for (const property of checker.getPropertiesOfType(target)) {
     const type = checker.getTypeOfSymbolAtLocation(property,node);
     const before = checker.getPropertyOfType(source,property.name);
     const beforeType = before && checker.getTypeOfSymbolAtLocation(before,node);
     if (callable(type)) fields.push({member:property.name,type:checker.typeToString(type),existing_equal_contract:!!beforeType && checker.isTypeAssignableTo(beforeType,type) && checker.isTypeAssignableTo(type,beforeType)});
    }
    const rootCallable = callable(target);
    const newCallable = fields.filter(field => !field.existing_equal_contract);
    const reasons = [];
    if (rootCallable) reasons.push('callable target');
    if (fields.length) reasons.push('direct callable members');
    if (newCallable.length) reasons.push('introduced or refined callable members');
    for (const reason of reasons) {
     const category = `${kind}: ${reason}`;
     counters[category] = (counters[category] || 0) + 1;
    }
    rows.push({file:relative,line:original.line,column:original.column,start:original.start,end:original.end,kind,source_type:original.source_type,target_type:original.target_type,root_callable:rootCallable,callable_members:fields,introduced_callable_members:newCallable});
    visited.add(`${relative}:${key}`);
   }
  }
  ts.forEachChild(node,visit);
 }
 visit(file);
}
assert.equal(visited.size,2936);
const untagged = rows.filter(row => row.kind === 'untagged');
const countBy = (items,key) => items.reduce((counts,item)=>(counts[key(item)]=(counts[key(item)]||0)+1,counts),{});
const descending = counts => Object.entries(counts).sort((a,b)=>b[1]-a[1] || a[0].localeCompare(b[0]));
const summary = {typescript:ts.version,source_commit:'050880ce59e30b356b686bd3144efe24f875ebc8',ledger_commit:'5b173f3920ab2c5b7058f0a9fe4b8e4a91f52523',ledger_sha256:crypto.createHash('sha256').update(fs.readFileSync(ledgerFile)).digest('hex'),tagged_sites:1758,untagged_sites:1178,diagnostics:0,contract_categories:counters,untagged_targets:descending(countBy(untagged,row=>row.target_type)),untagged_files:descending(countBy(untagged,row=>row.file)),limits:'Static exposure only, not execution heat, actual compiler refusals, erasure rates or timings. Direct callable contracts overlap categories. An equal callable contract already supplied by the source is separated from newly asserted callable behavior. Nested objects require transitive views, not eager recursive whole-shape validation. Actual refused operations require implementation proof accounting.'};
fs.mkdirSync(output,{recursive:true});
fs.writeFileSync(path.join(output,'view-contracts.json'),JSON.stringify(rows,null,2)+'\n');
fs.writeFileSync(path.join(output,'view-inventory.json'),JSON.stringify(summary,null,2)+'\n');
console.log(JSON.stringify(summary,null,2));
