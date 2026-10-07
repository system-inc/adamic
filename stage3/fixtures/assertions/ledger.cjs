// Stock TypeScript is the outside type oracle. Usage: NODE_PATH=<6.0.3 node_modules>
// node ledger.cjs <TypeScript checkout> <census data directory> [output directory]
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const assert = require('node:assert/strict');
const ts = require('typescript');
assert.equal(ts.version, '6.0.3');
const root = path.resolve(process.argv[2]);
const census = process.argv[3];
const output = process.argv[4] || __dirname;
const configPath = path.join(root, 'src/compiler/tsconfig.json');
const read = ts.readConfigFile(configPath, ts.sys.readFile);
assert.equal(read.error, undefined);
const config = ts.parseJsonConfigFileContent(read.config, ts.sys, path.dirname(configPath), undefined, configPath);
assert.equal(config.errors.length, 0);
const program = ts.createProgram(config.fileNames, config.options);
const checker = program.getTypeChecker();
const diagnostics = ts.getPreEmitDiagnostics(program);
assert.equal(diagnostics.length, 0, diagnostics.map(d => ts.flattenDiagnosticMessageText(d.messageText, '\n')).join('\n'));
const inventory = JSON.parse(fs.readFileSync(path.join(census, 'sites.json')));
const hashes = JSON.parse(fs.readFileSync(path.join(census, 'files.json')));
const expected = inventory.filter(s => s.reason === 'type assertion sites' && !s.file.includes('.generated.'));
const actual = [];
const nonnull = [];
const counts = {};
const nonnullCounts = {};
function parts(t) { return t.isUnion() ? t.types : [t]; }
function literals(t) {
 const values = [];
 for (const p of parts(t)) {
  if (p.flags & ts.TypeFlags.StringLiteral) values.push({type:'string', value:p.value});
  else if (p.flags & ts.TypeFlags.NumberLiteral) values.push({type:'number', value:p.value});
  else if (p.flags & ts.TypeFlags.BooleanLiteral) values.push({type:'boolean', value:p.intrinsicName === 'true'});
  else return undefined;
 }
 return values;
}
function interfaces(t) {
 return parts(t).every(p => p.symbol?.declarations?.some(ts.isInterfaceDeclaration));
}
function tags(source, target, node) {
 const result = [];
 for (const property of checker.getPropertiesOfType(target)) {
  const name = property.name;
  const sourceProperty = checker.getPropertyOfType(source, name);
  if (!sourceProperty || property.flags & ts.SymbolFlags.Optional || sourceProperty.flags & ts.SymbolFlags.Optional) continue;
  const narrowed = checker.getTypeOfSymbolAtLocation(property, node);
  const broad = checker.getTypeOfSymbolAtLocation(sourceProperty, node);
  const values = literals(narrowed);
  if (!values || !checker.isTypeAssignableTo(narrowed, broad) || checker.isTypeAssignableTo(broad, narrowed)) continue;
  // A union tag must partition all members. A shared literal with extra structure is not enough.
  let partition = false;
  if (source.isUnion()) {
   partition = source.types.every(member => {
    const p = checker.getPropertyOfType(member, name);
    if (!p || p.flags & ts.SymbolFlags.Optional) return false;
    const mt = checker.getTypeOfSymbolAtLocation(p, node);
    const mv = literals(mt);
    if (!mv) return false;
    const overlaps = mv.some(v => values.some(w => v.type === w.type && v.value === w.value));
    return !overlaps || checker.isTypeAssignableTo(member, target);
   });
  }
  result.push({name,values,union_partition:partition,source_type:checker.typeToString(broad),target_type:checker.typeToString(narrowed)});
 }
 return result;
}
function strip(n) { while (ts.isParenthesizedExpression(n)) n = n.expression; return n; }
function unknownChain(n) {
 const inner = strip(n.expression);
 if ((ts.isAsExpression(inner) || ts.isTypeAssertionExpression(inner)) && inner.type.kind === ts.SyntaxKind.UnknownKeyword) return 'outer';
 const parent = stripParent(n);
 if (n.type.kind === ts.SyntaxKind.UnknownKeyword && (ts.isAsExpression(parent) || ts.isTypeAssertionExpression(parent)) && strip(parent.expression) === n) return 'inner';
 return undefined;
}
function stripParent(n) { let p = n.parent; while (p && ts.isParenthesizedExpression(p)) p = p.parent; return p; }
function location(file, node) {
 const start = node.getStart(file);
 const p = file.getLineAndCharacterOfPosition(start);
 return {file:path.relative(root,file.fileName),line:p.line+1,column:p.character+1,start,end:node.end,text:node.getText(file)};
}
for (const file of program.getSourceFiles().filter(f => f.fileName.startsWith(path.join(root,'src/compiler') + '/') && !f.fileName.includes('.generated.')).sort((a,b) => a.fileName.localeCompare(b.fileName))) {
 const relative = path.relative(root,file.fileName);
 const hash = hashes.find(f => f.file === relative);
 assert.ok(hash, relative);
 assert.equal(crypto.createHash('sha256').update(fs.readFileSync(file.fileName)).digest('hex'), hash.sha256, relative);
 function visit(node) {
  if (ts.isNonNullExpression(node)) {
   const expression = strip(node.expression);
   const form = ts.isPropertyAccessExpression(expression) ? 'field' : ts.isElementAccessExpression(expression) ? 'indexed read' : ts.isCallExpression(expression) ? 'call' : expression.kind === ts.SyntaxKind.Identifier && expression.text === 'undefined' ? 'literal undefined' : 'other';
   nonnull.push({...location(file,node),form});
   nonnullCounts[form] = (nonnullCounts[form] || 0) + 1;
  }
  if (ts.isAsExpression(node) || ts.isTypeAssertionExpression(node)) {
   const source = checker.getTypeAtLocation(node.expression);
   const target = node.type.getText(file) === 'const' ? checker.getTypeAtLocation(node) : checker.getTypeFromTypeNode(node.type);
   const chain = unknownChain(node);
   const tagCandidates = tags(source,target,node);
   const forward = checker.isTypeAssignableTo(source,target);
   const reverse = checker.isTypeAssignableTo(target,source);
   const unsafe = !!((source.flags | target.flags) & ts.TypeFlags.Any);
   let category, detail;
   if (node.type.getText(file) === 'const') category = 'as const';
   else if (chain) category = 'as unknown as';
   else if (!unsafe && forward) category = 'upcast';
   else if (!unsafe && reverse && source.isUnion() && tagCandidates.some(t => t.union_partition)) category = 'tagged union downcast';
   else if (!unsafe && reverse && interfaces(source) && interfaces(target) && tagCandidates.length === 0) category = 'structural interface downcast without tag';
   else {
    category = 'other';
    detail = unsafe ? 'any involved' : reverse && tagCandidates.length ? 'tagged narrowing outside a partitioned union' : reverse ? 'other narrowing' : 'non-assignable assertion';
   }
   counts[category] = (counts[category] || 0) + 1;
   actual.push({...location(file,node),syntax:ts.isAsExpression(node)?'as':'angle bracket',category,source_type:checker.typeToString(source),target_type:checker.typeToString(target),source_assignable_to_target:forward,target_assignable_to_source:reverse,tags:tagCandidates,...(chain?{chain_role:chain}:{}),...(detail?{detail}:{})});
  }
  ts.forEachChild(node,visit);
 }
 visit(file);
}
const keys = rows => rows.map(s => `${s.file}:${s.start}:${s.end}`).sort();
assert.deepEqual(keys(actual), keys(expected));
assert.equal(actual.length,4101);
assert.deepEqual(keys(nonnull),keys(inventory.filter(s => s.reason === 'the non-null assertion !' && !s.file.includes('.generated.'))));
assert.equal(nonnull.length,1123);
const summary = {typescript:ts.version,source_commit:'050880ce59e30b356b686bd3144efe24f875ebc8',assertions:actual.length,counts,other_details:actual.filter(s=>s.category==='other').reduce((a,s)=>(a[s.detail]=(a[s.detail]||0)+1,a),{}),syntax_counts:actual.reduce((a,s)=>(a[s.syntax]=(a[s.syntax]||0)+1,a),{}),nonnull:nonnull.length,nonnull_counts:nonnullCounts,diagnostics:diagnostics.length};
fs.mkdirSync(output,{recursive:true});
for (const [name,data] of Object.entries({'ledger.json':actual,'nonnull-ledger.json':nonnull,'ledger-summary.json':summary})) fs.writeFileSync(path.join(output,name),JSON.stringify(data,null,2)+'\n');
const locations = ['file\tline\tcolumn\tstart\tend\tcategory\tdiscriminants'];
for (const row of actual) locations.push([row.file,row.line,row.column,row.start,row.end,row.category,row.tags.map(t => t.name).join(',') || '-'].join('\t'));
fs.writeFileSync(path.join(output,'locations.tsv'),locations.join('\n')+'\n');
console.log(JSON.stringify(summary,null,2));
