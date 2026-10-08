// Classify the exact 69 cumulative latent findings using the pinned upstream checker.
// This audits the baseline ledger, not this branch's admitted programs or runtime.
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const crypto = require('node:crypto');
const {execFileSync} = require('node:child_process');
const ts = require('typescript');
const [root, reportFile, output] = process.argv.slice(2);
assert.equal(ts.version, '6.0.3');
assert.equal(execFileSync('git', ['-C', root, 'rev-parse', 'HEAD'], {encoding:'utf8'}).trim(), '050880ce59e30b356b686bd3144efe24f875ebc8');
const report = JSON.parse(fs.readFileSync(reportFile));
const run = report.runs.find(run => run.name === 'cumulative');
assert.equal(run.head, '15eb079bc200cb1f3f8137fa762fc53bdbc45f22');
const findings = run.findings.filter(f => f.kind === 'Refused' && f.reason === "a cast the runtime can't check");
assert.equal(findings.length, 69);
const configPath = path.join(root, 'src/compiler/tsconfig.json');
const configRead = ts.readConfigFile(configPath, ts.sys.readFile);
assert.equal(configRead.error, undefined);
const config = ts.parseJsonConfigFileContent(configRead.config, ts.sys, path.dirname(configPath), undefined, configPath);
assert.equal(config.errors.length, 0);
const program = ts.createProgram(config.fileNames, config.options);
assert.equal(ts.getPreEmitDiagnostics(program).length, 0);
const checker = program.getTypeChecker();
const signatures = type => checker.getSignaturesOfType(type, ts.SignatureKind.Call).length + checker.getSignaturesOfType(type, ts.SignatureKind.Construct).length;
const members = type => type.isUnion() ? type.types : [type];
const rows = findings.map(finding => {
 const match = /^(.*):(\d+):(\d+)$/.exec(finding.where);
 assert(match);
 const file = program.getSourceFile(path.join(root, match[1]));
 assert(file);
 const position = file.getPositionOfLineAndCharacter(Number(match[2])-1, Number(match[3])-1);
 let cast;
 function visit(node) {
  if (node.getStart(file) <= position && position < node.end) {
   if (ts.isAsExpression(node) || ts.isTypeAssertionExpression(node)) cast = node;
   ts.forEachChild(node, visit);
  }
 }
 visit(file);
 assert(cast, `No cast at ${finding.where}; adapted locations must be mapped before classification`);
 const target = checker.getTypeFromTypeNode(cast.type);
 const source = checker.getTypeAtLocation(cast.expression);
 const callableMembers = checker.getPropertiesOfType(target).filter(p => members(checker.getTypeOfSymbolAtLocation(p,cast)).some(t => signatures(t))).map(p => p.name);
 const rootCallable = members(target).some(t => signatures(t));
 const indexKinds = checker.getIndexInfosOfType(target).map(i => checker.typeToString(i.keyType));
 const flags = target.flags;
 const shape = checker.isTupleType(target) ? 'tuple' : checker.isArrayType(target) ? 'array' : flags & ts.TypeFlags.TypeParameter ? 'type parameter' : flags & ts.TypeFlags.Intersection ? 'intersection' : flags & ts.TypeFlags.Union ? 'union' : !(flags & ts.TypeFlags.Object) ? 'scalar or other' : rootCallable ? 'callable target' : callableMembers.length ? 'object with direct callable members' : indexKinds.length ? 'indexed object' : 'object without direct callable members';
 return {where:finding.where, text:cast.getText(file), source_type:checker.typeToString(source), target_type:checker.typeToString(target), shape, root_callable:rootCallable, callable_members:callableMembers, index_key_types:indexKinds};
});
const counts = rows.reduce((counts,row) => (counts[row.shape]=(counts[row.shape]||0)+1,counts), {});
const result = {baseline_commit:'70456b7',baseline_run:run.head,source_commit:'050880ce59e30b356b686bd3144efe24f875ebc8',report_sha256:crypto.createHash('sha256').update(fs.readFileSync(reportFile)).digest('hex'),measurement:run.measurement,baseline_refusals:69,diagnostics:0,shape_counts:counts,limits:'Baseline target inventory only. These are first encountered refusals in eligible units on a checker-rejected corpus. They are not the final refused shapes after checked views, nor exhaustive cast counts. Nested callable fields need source proof or a deferred refusal at invocation.',rows};
fs.writeFileSync(output, JSON.stringify(result,null,2)+'\n');
console.log(JSON.stringify(counts));
