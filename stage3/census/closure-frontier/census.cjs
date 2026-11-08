'use strict';
// Usage: NODE_PATH=stage3/api/node_modules node census.cjs ADAPTED_TREE OUTPUT_DIR
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const cp = require('node:child_process');
const ts = require('typescript');
if (ts.version !== '6.0.3') throw Error(`Expected TypeScript 6.0.3, got ${ts.version}`);
const [treeArg, outputArg] = process.argv.slice(2);
if (!treeArg || !outputArg) throw Error('Usage: census.cjs ADAPTED_TREE OUTPUT_DIR');
const root = path.resolve(treeArg), out = path.resolve(outputArg);
if (!fs.existsSync(path.join(root, 'patch-set.md')) && !process.env.CENSUS_FIXTURE) throw Error('Adaptation is incomplete: patch-set.md missing');
const compiler = path.join(root, 'src/compiler');
const config = ts.readConfigFile(path.join(compiler, 'tsconfig.json'), ts.sys.readFile);
if (config.error) throw Error(ts.flattenDiagnosticMessageText(config.error.messageText, '\n'));
const parsed = ts.parseJsonConfigFileContent(config.config, ts.sys, compiler);
const options = {...parsed.options, noEmit: true, composite: false, isolatedDeclarations: false,
  declaration: false, declarationMap: false, emitDeclarationOnly: false, exactOptionalPropertyTypes: true, noUncheckedIndexedAccess: true, erasableSyntaxOnly: false, types: ['node'],
  typeRoots: [path.resolve(path.dirname(require.resolve('typescript/package.json')), '../@types')]};
const entry = path.join(root, 'src/tsc/tsc.ts');
const program = ts.createProgram([...new Set([entry, ...parsed.fileNames])], options);
const checker = program.getTypeChecker();
const files = program.getSourceFiles().filter(f => f.fileName.startsWith(compiler + '/') && !f.isDeclarationFile);
const relative = f => path.relative(root, typeof f === 'string' ? f : f.fileName).split(path.sep).join('/');
const typeText = t => checker.typeToString(t, undefined, ts.TypeFormatFlags.NoTruncation);
function sym(n) {
  let s = checker.getSymbolAtLocation(ts.isPropertyAccessExpression(n) ? n.name : n);
  if (s && s.flags & ts.SymbolFlags.Alias) s = checker.getAliasedSymbol(s);
  return s;
}
function unwrap(n) {
  while (ts.isParenthesizedExpression(n) || ts.isNonNullExpression(n) || ts.isAsExpression(n) || ts.isTypeAssertionExpression(n) || ts.isSatisfiesExpression(n)) n = n.expression;
  return n;
}
// ESM static dependency traversal, source order, marking on entry to break cycles.
// Type-only edges are omitted. All file indices are 1-based completion order.
const seen = new Set(), evaluation = [];
function evaluate(f) {
  if (!f || seen.has(f.fileName) || f.isDeclarationFile) return;
  seen.add(f.fileName);
  for (const s of f.statements) {
    if (!(ts.isImportDeclaration(s) || ts.isExportDeclaration(s)) || !s.moduleSpecifier || s.isTypeOnly || s.importClause?.isTypeOnly) continue;
    const clause = s.importClause;
    if (clause && !clause.name && clause.namedBindings && ts.isNamedImports(clause.namedBindings) && clause.namedBindings.elements.length && clause.namedBindings.elements.every(e => e.isTypeOnly)) continue;
    if (s.exportClause && ts.isNamedExports(s.exportClause) && s.exportClause.elements.length && s.exportClause.elements.every(e => e.isTypeOnly)) continue;
    const resolved = ts.resolveModuleName(s.moduleSpecifier.text, f.fileName, options, ts.sys).resolvedModule;
    if (resolved) evaluate(program.getSourceFile(resolved.resolvedFileName));
    else if (s.moduleSpecifier.text.startsWith('.')) throw Error(`Unresolved ${s.moduleSpecifier.text} in ${f.fileName}`);
  }
  evaluation.push(relative(f));
}
evaluate(program.getSourceFile(entry));
const order = new Map(evaluation.map((f,i) => [f,i+1]));
const writes = new Set();
function scanWrites(n) {
  if (ts.isBinaryExpression(n) && n.operatorToken.kind >= ts.SyntaxKind.FirstAssignment && n.operatorToken.kind <= ts.SyntaxKind.LastAssignment) {
    const target = unwrap(n.left); if (ts.isIdentifier(target)) writes.add(sym(target));
  }
  if ((ts.isPrefixUnaryExpression(n) || ts.isPostfixUnaryExpression(n)) && [ts.SyntaxKind.PlusPlusToken, ts.SyntaxKind.MinusMinusToken].includes(n.operator)) writes.add(sym(n.operand));
  ts.forEachChild(n, scanWrites);
}
files.forEach(scanWrites);
function hasAssertion(n) {
  while (ts.isParenthesizedExpression(n) || ts.isNonNullExpression(n) || ts.isSatisfiesExpression(n)) n=n.expression;
  return ts.isAsExpression(n) || ts.isTypeAssertionExpression(n);
}
function calleeKind(original) {
  if (hasAssertion(original)) return {kind:calleeKind(unwrap(original)).kind || 'other', unsure:'asserted callable expression may wrap or adapt closure identity'};
  const n = unwrap(original);
  if (ts.isArrowFunction(n) || ts.isFunctionExpression(n)) return {bound:true};
  if (ts.isConditionalExpression(n) || ts.isBinaryExpression(n) && [ts.SyntaxKind.QuestionQuestionToken, ts.SyntaxKind.BarBarToken, ts.SyntaxKind.AmpersandAmpersandToken].includes(n.operatorToken.kind)) return {kind:'join'};
  if (ts.isCallExpression(n)) return {kind:'returned value'};
  const s = sym(n), ds = s?.declarations || [];
  let d = s?.valueDeclaration || ds[0];
  if (d && (ts.isPropertyAccessExpression(n) || ts.isElementAccessExpression(n)) && !ts.isFunctionDeclaration(d) && !program.isSourceFileDefaultLibrary(d.getSourceFile()) && d.getSourceFile().fileName.includes('/node_modules/')) return {kind:'property', unsure:'external library or host intrinsic dispatch'};
  if (d && (ts.isPropertyAccessExpression(n) || ts.isElementAccessExpression(n)) && ts.isParameter(d)) return {kind:'property'};
  if (d && ts.isBindingElement(d)) {
    while (d && !ts.isParameter(d) && !ts.isVariableDeclaration(d)) d = d.parent;
    if (d && ts.isParameter(d)) return {kind:'parameter'};
    if (d && ts.isVariableDeclaration(d)) return {kind: ts.isVariableDeclarationList(d.parent) && !(d.parent.flags & ts.NodeFlags.Const) ? 'let or var' : 'const indirect'};
  }
  if (d && program.isSourceFileDefaultLibrary(d.getSourceFile()) && ts.isIdentifier(n)) return {bound:true};
  if (d && ts.isFunctionDeclaration(d)) return writes.has(s) ? {kind:'other', unsure:'reassigned function declaration'} : {bound:true};
  if (d && ts.isParameter(d)) return {kind:'parameter'};
  if (d && ts.isVariableDeclaration(d)) {
    const list = d.parent;
    if (ts.isVariableDeclarationList(list) && !(list.flags & ts.NodeFlags.Const)) return {kind:'let or var'};
    if (d.initializer && hasAssertion(d.initializer)) return {kind:'const indirect', unsure:'asserted const initializer may wrap or adapt closure identity'};
    if (d.initializer && (ts.isArrowFunction(unwrap(d.initializer)) || ts.isFunctionExpression(unwrap(d.initializer)))) return {bound:true};
    return {kind:'const indirect'};
  }
  if (ts.isPropertyAccessExpression(n) || ts.isElementAccessExpression(n)) {
    if (d && ts.isMethodDeclaration(d) && (ts.isClassDeclaration(d.parent) || ts.isClassExpression(d.parent)) && !n.questionDotToken) return {bound:true};
    if (d && program.isSourceFileDefaultLibrary(d.getSourceFile())) {
      const owner = d.parent?.name?.getText() || '';
      if (/^(Array|ReadonlyArray|Map|ReadonlyMap|Set|ReadonlySet|String|Number|Boolean|RegExp|Function|CallableFunction|NewableFunction|Promise|Math|JSON|Object|Symbol)(Constructor)?$/.test(owner) || /^(?:Uint|Int|Float|BigInt|BigUint)\d+(?:Clamped)?Array/.test(owner)) return {bound:true};
      return {kind:'property', unsure:'standard library receiver dispatch not proven to be an intrinsic'};
    }
    if (d && !ts.isFunctionDeclaration(d) && !program.isSourceFileDefaultLibrary(d.getSourceFile()) && d.getSourceFile().fileName.includes('/node_modules/')) return {kind:'property', unsure:'external library or host intrinsic dispatch'};
    if (d && (ts.isPropertySignature(d) || ts.isPropertyDeclaration(d) || ts.isMethodSignature(d) || ts.isPropertyAssignment(d))) return {kind:'property'};
    return {kind:'property', unsure:'property or element dispatch not settled by declaration'};
  }
  return {kind:'other', unsure:'unclassified callable expression'};
}
const scalarFlags = ts.TypeFlags.NumberLike | ts.TypeFlags.BooleanLike | ts.TypeFlags.StringLike | ts.TypeFlags.Void | ts.TypeFlags.Undefined | ts.TypeFlags.Never;
function resultClass(t) {
  if (t.isUnion()) {
    const members = t.types.map(x => x.flags & ts.TypeFlags.Undefined ? 'scalar' : resultClass(x));
    if (members.some(x => x === 'unsure')) return 'unsure';
    if (members.every(x => x === 'scalar')) {
      const groups = new Set(t.types.filter(x => !(x.flags & (ts.TypeFlags.Undefined | ts.TypeFlags.Never))).map(x => x.flags & ts.TypeFlags.NumberLike ? 'number' : x.flags & ts.TypeFlags.BooleanLike ? 'boolean' : x.flags & ts.TypeFlags.StringLike ? 'string' : 'void'));
      return groups.size <= 1 ? 'scalar' : 'nonScalar';
    }
    return members.some(x => x === 'scalar') ? 'mixed' : 'nonScalar';
  }
  if (t.flags & ts.TypeFlags.Undefined) return 'nonScalar';
  if (t.flags & scalarFlags) return 'scalar';
  if (t.flags & (ts.TypeFlags.Any | ts.TypeFlags.BigIntLike | ts.TypeFlags.ESSymbolLike | ts.TypeFlags.TypeParameter | ts.TypeFlags.Conditional | ts.TypeFlags.IndexedAccess | ts.TypeFlags.Substitution)) return 'unsure';
  const name = t.symbol?.name || '';
  if (/^(?:Uint|Int|Float|BigInt|BigUint)\d+(?:Clamped)?Array$/.test(name)) return 'scalar';
  if (t.isIntersection()) {
    if (t.types.some(x => x.flags & scalarFlags)) return 'scalar';
    if (t.types.every(x => resultClass(x) === 'unsure')) return 'unsure';
  }
  return 'nonScalar';
}
function sourceClass(t) {
  const members = t.isUnion() ? t.types : [t];
  const sourceScalar = x => !!(x.flags & scalarFlags) || x.isIntersection() && x.types.some(sourceScalar);
  const scalar = members.map(sourceScalar);
  return scalar.every(Boolean) ? 'scalar' : scalar.some(Boolean) ? 'mixed' : 'nonScalar';
}
function owner(node) {
  for (let p=node.parent;p;p=p.parent) if (ts.isFunctionLike(p) && p.name) return p.name.getText();
  return '(anonymous)';
}
const sites = [], excluded = {};
function exclude(why) { excluded[why] = (excluded[why] || 0) + 1; }
function visit(n) {
  if (ts.isCallExpression(n)) {
    const e = unwrap(n.expression);
    let info;
    if (ts.isPropertyAccessExpression(e) && ['reduce','reduceRight'].includes(e.name.text) && (checker.isArrayType(checker.getTypeAtLocation(e.expression)) || checker.isTupleType(checker.getTypeAtLocation(e.expression)))) {
      if (!n.arguments.length) { exclude('reduce without callback'); ts.forEachChild(n,visit); return; }
      const callback = calleeKind(n.arguments[0]);
      info = callback.bound ? {bound:true} : {kind:'reduce callback', callbackKind:callback.kind, unsure:callback.unsure};
      if (!info.bound && (e.name.text === 'reduceRight' || n.arguments.length !== 2)) info.unsure = 'reduceRight is not an ArrayReduce intrinsic, or missing required initial value';
    } else info = calleeKind(n.expression);
    if (info.bound) exclude('bounded target, direct or virtual Call, or library intrinsic');
    else {
      const t = checker.getTypeAtLocation(n), result = resultClass(t);
      if (n.questionDotToken || n.flags & ts.NodeFlags.OptionalChain) info.unsure = 'optional chain: optionalCall/optionalIntrinsic can refuse or rewrite before CallClosure';
      if (result === 'unsure' && !info.unsure) info.unsure = 'result representation depends on generic instantiation, erased type, or scalar-union boxing';
      const f = n.getSourceFile(), p = f.getLineAndCharacterOfPosition(n.getStart(f));
      const targetSymbol=sym(e), targetDeclaration=targetSymbol?.valueDeclaration || targetSymbol?.declarations?.[0];
      const declarationEvidence=targetDeclaration ? {file: relative(targetDeclaration.getSourceFile()), kind:ts.SyntaxKind[targetDeclaration.kind], start:targetDeclaration.getStart()} : null;
      sites.push({file:relative(f), start:n.getStart(f), end:n.end, line:p.line+1, column:p.character+1, loadOrder:order.get(relative(f)) ?? null,
        text:n.getText(f), declarationEvidence, owner:owner(n), kind:info.kind, ...(info.callbackKind ? {callbackKind:info.callbackKind} : {}), type:typeText(t),
        sourceResult:sourceClass(t), bucket:info.unsure ? 'unsure' : result,
        ...(info.unsure ? {reason:info.unsure} : {})});
    }
  }
  ts.forEachChild(n,visit);
}
files.forEach(visit);
const empty = () => ({nonScalar:0, mixed:0, scalar:0, unsure:0, stops:0});
const totals=empty(), byKind=Object.fromEntries(['parameter','property','let or var','returned value','join','reduce callback','const indirect','other'].map(k => [k,empty()]));
const byFile = new Map(files.map(f => [relative(f), {file:relative(f), loadOrder:order.get(relative(f)) ?? null,...empty()}]));
for (const s of sites) for (const row of [totals, byKind[s.kind], byFile.get(s.file)]) { row[s.bucket]++; if (s.bucket==='nonScalar' || s.bucket==='mixed') row.stops++; }
const diagnostics = ts.getPreEmitDiagnostics(program);
const diagnosticCounts = {};
for (const d of diagnostics) diagnosticCounts[d.code]=(diagnosticCounts[d.code]||0)+1;
const manifest=files.map(f=>({file:relative(f),sha256:crypto.createHash('sha256').update(f.text).digest('hex')}));
const git = (...args) => cp.execFileSync('git',args,{encoding:'utf8',cwd:path.resolve(__dirname,'../../..')}).trim();
if (new Set(sites.map(s=>s.file+':'+s.start+':'+s.end)).size !== sites.length) throw Error('Duplicate source site');
if (Object.values(byKind).reduce((n,r)=>n+r.stops,0)!==totals.stops || dataCheckFiles()!==totals.stops) throw Error('Unbalanced totals');
function dataCheckFiles() { return [...byFile.values()].reduce((n,r)=>n+r.stops,0); }
const scalarShorthandDifferences = sites.filter(s=>s.bucket !== 'unsure' && ((s.sourceResult === 'scalar' && s.bucket !== 'scalar') || (s.sourceResult !== 'scalar' && s.bucket === 'scalar')));
const data = {scalarShorthandDifferences, typescript:ts.version, main:git('merge-base','HEAD','origin/main'), sourcePin:git('-C',root,'rev-parse','HEAD'),
  adaptedPatchSetSha256: fs.existsSync(path.join(root,'patch-set.md')) ? crypto.createHash('sha256').update(fs.readFileSync(path.join(root,'patch-set.md'))).digest('hex') : null,
  probe:'39a3cebf', parent:'266e385b', compilerFiles:files.length, compilerFilesOutsideEntry:files.filter(f=>!order.has(relative(f))).map(relative),
  options: JSON.parse(JSON.stringify(options).split(root).join('<adapted>').split(path.resolve(path.dirname(require.resolve('typescript/package.json')), '..')).join('<api-node-modules>')), diagnosticDetails: diagnostics.map(d=>({code:d.code, file:d.file ? relative(d.file):null, start:d.start, message:ts.flattenDiagnosticMessageText(d.messageText,'\n')})), diagnostics:diagnostics.length, diagnosticCounts, totals, byKind,
  byFile:[...byFile.values()].sort((a,b)=>b.stops-a.stops || b.unsure-a.unsure || a.file.localeCompare(b.file)), evaluationOrder:evaluation, excluded, manifest, sites};
fs.mkdirSync(out,{recursive:true});
fs.writeFileSync(path.join(out,'counts.json'),JSON.stringify(data,null,2).replace(/[\u2013\u2014]/g, c => c === '\u2013' ? '\\u2013' : '\\u2014')+'\n');
console.log(JSON.stringify({totals,byKind,topTen:data.byFile.slice(0,10),diagnostics:data.diagnosticCounts},null,2));
if (process.env.CENSUS_FIXTURE) process.exit(0);
const rows = data.byFile;
const cell = s => String(s).replace(/[\u2013\u2014]/g,'-').replace(/\|/g,'\\|').replace(/\r?\n/g,' ').replace(/`/g,'&#96;');
const inline = value => { const s=String(value).replace(/[\u2013\u2014]/g,'-').replace(/\r?\n/g,' '); const ticks='`'.repeat(Math.max(0,...(s.match(/`+/g)||[]).map(x=>x.length))+1); return ticks+' '+s+' '+ticks; };
const table = (headers, values) => '| '+headers.join(' | ')+' |\n|'+headers.map(()=> '---').join('|')+'|\n'+values.map(r=>'| '+r.map(cell).join(' | ')+' |').join('\n')+'\n';
const summaryJSON = {totals, byKind, compilerFiles:files.length, compilerFilesOutsideEntry:data.compilerFilesOutsideEntry,
  scalarShorthandDifferences:scalarShorthandDifferences.length};
const selected=[];
function pick(predicate) { const s=sites.find(s=>!selected.includes(s) && predicate(s)); if(s) selected.push(s); }
for(const helper of ['forEach','forEachRight','map','filter','sameMap','flatMap']) pick(s=>s.file==='src/compiler/core.ts' && s.owner===helper && s.kind==='parameter');
for(const kind of Object.keys(byKind)) pick(s=>s.kind===kind && ['nonScalar','mixed'].includes(s.bucket) && s.text.length<220);
for(const bucket of ['unsure','mixed','nonScalar']) pick(s=>s.bucket===bucket && s.text.length<160 && s.file!=='src/compiler/core.ts');
for(const s of sites) if(selected.length<20 && !selected.includes(s) && ['nonScalar','mixed'].includes(s.bucket) && s.text.length<160 && !selected.some(x=>x.file===s.file && x.kind===s.kind)) selected.push(s);
for(const s of sites) if(selected.length<20 && !selected.includes(s) && s.text.length<160) selected.push(s);
if(selected.length!==20) throw Error('Expected twenty samples');
const unsureReasons={}; for(const s of sites.filter(s=>s.bucket==='unsure')) unsureReasons[s.reason]=(unsureReasons[s.reason]||0)+1;
const evidenceBase='https://github.com/system-inc/adamic/blob/39a3cebf/';
let report=`# Function-value result frontier census\n\n${totals.stops} source sites satisfy the recovered unknown-target, non-scalar result rule: ${totals.nonScalar} non-scalar-only results and ${totals.mixed} mixed scalar/non-scalar unions. ${totals.unsure} additional sites are unsure and are excluded from both stop and scalar totals. ${totals.scalar} unknown-target sites have scalar results.\n\nThis is a static source-site census, conditional on each site reaching the relevant IR node. It is not a count of executed calls, generic instantiations, generated adapter nodes, or successive compiler failures. Unrelated earlier lowering refusals and reachability are not simulated. A helper body is counted once per syntactic call expression, irrespective of how many callers use it.\n\n`;
report+=`## Inputs and rule confirmation\n\n- Main at fetch time: \`${data.main}\`. Only this directory is changed.\n- TypeScript source: \`${data.sourcePin}\`, v6.0.3, adapted by main's \`stage3/apply.sh\`. Patch table SHA-256: \`${data.adaptedPatchSetSha256}\`.\n- Checker: TypeScript ${ts.version} from \`stage3/api\`; compiler tsconfig roots plus \`src/tsc/tsc.ts\`, with strict, exactOptionalPropertyTypes and noUncheckedIndexedAccess enabled. Effective options and all source hashes are in [counts.json](counts.json).\n- Probe: [39a3cebf versus 266e385b](https://github.com/system-inc/adamic/compare/266e385b...39a3cebf). The changed \`graphFlowRules\` entries stop exactly \`CallClosure\` and \`ArrayReduce\` when \`ClosureTargets(...).Unknown && !graphFlowScalar(expression.Type())\`. They do not add the same stop to ArrayMap, ArrayVisit, ArrayFrom, or MapForEach.\n\nThe proposed source reading needs these corrections:\n\n1. [ClosureTargets](${evidenceBase}internal/ir/call_targets.go) also bounds \`CallClosure.Direct\`, \`BoundClosure\`, and \`Effects\` whose result unwraps to a bounded value. Plain const aliases are not transitively followed. [locals.go](${evidenceBase}internal/lower/locals.go) sets ConstantClosure only for an initializer that lowers directly to MakeClosure. Indirect consts have their own extra row because the six requested kinds are not exhaustive.\n2. [callOrMethod](${evidenceBase}internal/lower/class.go) routes a class MethodDeclaration to a direct or virtual Call, but structural MethodSignature declarations fall back to [call/callClosure](${evidenceBase}internal/lower/expression.go). Consequently interface methods, as well as function-valued properties, belong in the property row. Namespace references resolving to declared functions are excluded. External host dispatch and unresolved element/property dispatch are unsure.\n3. [graphFlowScalar](${evidenceBase}internal/lower/graph_flow_classes.go) accepts IR void, number, boolean, string, MaybeNumber, MaybeBoolean, and typed arrays. The [representation function](${evidenceBase}internal/lower/expression.go) boxes differently represented scalar unions such as string | number as ir.Union, so those stop too. Standalone undefined is ir.Object; void and never call results are separately lowered to zero. The suggested scalar shorthand disagrees at ${scalarShorthandDifferences.length} classified sites, listed in counts.json's scalarShorthandDifferences. Mixed means the checker union contains both a source scalar member and a non-scalar member; object | undefined is mixed.\n4. Only \`.reduce(callback, initial)\` is recognized by [arrayReduce](${evidenceBase}internal/lower/object.go); reduceRight is absent from the intrinsic set. Unknown-callback reduceRight and reduce without the required initial value are unsure. Closure literals and direct const closure callbacks are bounded. Intrinsic callback representation adapters can introduce a bounded outer closure plus an unknown inner CallClosure; generated sites have no separate source call and are outside this source-site census.\n5. \`optionalCall\` generally refuses \`f?.()\` before CallClosure. \`optionalIntrinsic\` can rewrite chains and \`object?.method()\` has exceptions. Every optional-chain candidate is therefore kept in unsure rather than presumed to hit the probe. Generic result parameters, any, conditional/indexed-access results, asserted callable expressions/initializers, and unresolved callable expressions are also unsure. Arrays or objects containing a generic element still have a known non-scalar outer representation.\n\n`;
report+=`## Counts as JSON\n\n\`\`\`json\n${JSON.stringify(summaryJSON,null,2)}\n\`\`\`\n\nStops = nonScalar + mixed. Unsure is disjoint from every other column. The other row holds cases that cannot safely be assigned a requested kind. The const indirect row includes immutable aliases, destructured const function values and for-of const function bindings. Destructured parameter function values are parameters.\n\n`;
report+='## By callee kind\n\n'+table(['Kind','Stops','Non-scalar only','Mixed','Scalar','Unsure'],Object.entries(byKind).map(([k,r])=>[k,r.stops,r.nonScalar,r.mixed,r.scalar,r.unsure]))+'\n';
report+=`## By file, descending stop count\n\nLoad order is a 1-based, dependency-first completion index from src/tsc/tsc.ts. The walk follows static imports and re-exports in source order, skips explicit type-only edges, and marks modules on entry to break cycles. Namespace barrel and supporting modules occupy indices too. This is module initialization order, not the time a function body executes. Ties sort by unsure count then filename. All ${files.length} compiler files are listed, including zeros; an absent entry-program index is shown as outside entry. The full order is in counts.json's evaluationOrder.\n\n`;
report+=table(['File','Load index','Stops','Non-scalar only','Mixed','Scalar','Unsure'],rows.map(r=>[r.file,r.loadOrder??'outside entry',r.stops,r.nonScalar,r.mixed,r.scalar,r.unsure]))+'\n';
report+='## Twenty sample sites\n\nScalar and unsure core helper controls are included deliberately. Parameter callbacks behind map, filter and forEach-style helpers are individual source sites, even when many compiler callers pass through them. Exact unabridged call text and checker type strings for every candidate are in counts.json.\n\n';
for(const [i,s] of selected.entries()) report+=`${i+1}. **${s.file}:${s.line}:${s.column}**; ${s.kind}; ${s.bucket}; helper/owner ${inline(s.owner)}.\n\n   Call: ${inline(s.text)}\n\n   Checker result: ${inline(s.type)}\n\n`;
report+='## Unsure and validation\n\n'+table(['Reason','Sites'],Object.entries(unsureReasons).sort((a,b)=>b[1]-a[1]))+'\n';
report+=`The compiler program reports ${diagnostics.length} diagnostics. Their codes, locations and messages are retained in counts.json; checker errors are not silently suppressed. The census does not claim successful Adamic lowering of the whole program.\n\nValidation used a separate small program covering parameter/destructured parameter calls, mixed and scalar returns, unknown returns, typed arrays, structural method signatures, function properties, mutable bindings, indirect consts, returned functions, joins, reduce callbacks, optional calls and generic results. Fifteen classification assertions and bounded/direct/virtual-call exclusion assertions passed.\n\nReproduce from the main base above:\n\n\`\`\`bash\nsource /workspace/adamic-tools/env.sh\nbash cloud/setup.sh > /tmp/closure-setup.log 2>&1\nrc=$?; printf 'setup exit=%s\\n' "$rc"; [ "$rc" -eq 0 ] || exit "$rc"\nnpm ci --prefix stage3/api > /tmp/closure-npm.log 2>&1\nrc=$?; printf 'npm exit=%s\\n' "$rc"; [ "$rc" -eq 0 ] || exit "$rc"\nexport ADAMIC_TYPESCRIPT_SOURCE=/path/to/TypeScript-at-050880ce59e30b356b686bd3144efe24f875ebc8\nbash stage3/apply.sh /tmp/closure-adapted > /tmp/closure-apply.log 2>&1\nrc=$?; printf 'apply exit=%s\\n' "$rc"; [ "$rc" -eq 0 ] || exit "$rc"\nNODE_PATH=stage3/api/node_modules node stage3/census/closure-frontier/census.cjs /tmp/closure-adapted stage3/census/closure-frontier > /tmp/closure-census.log 2>&1\nrc=$?; printf 'census exit=%s\\n' "$rc"; exit "$rc"\n\`\`\`\n`;
fs.writeFileSync(path.join(out,'REPORT.md'),report);
