#!/usr/bin/env node
// Measurement only. Broad source upper bound, not the checked-view admission rule.
const fs = require('node:fs'), path = require('node:path'), zlib = require('node:zlib');
const ts = require('../../api/node_modules/typescript');
if (ts.version !== '6.0.3') throw Error(`Expected 6.0.3, got ${ts.version}`);
const [treeArg, out, fixture] = process.argv.slice(2);
if (!treeArg || !out) throw Error('Usage: node count.cjs TREE OUTPUT_PREFIX [fixture]');
const tree = path.resolve(treeArg);
const options = {strict:true, strictBindCallApply:true, useUnknownInCatchVariables:true,
 noUncheckedIndexedAccess:true, exactOptionalPropertyTypes:true, erasableSyntaxOnly:false,
 verbatimModuleSyntax:true, allowImportingTsExtensions:true, noEmit:true, emitDeclarationOnly:false,
 module:ts.ModuleKind.ESNext,moduleDetection:ts.ModuleDetectionKind.Force,
 moduleResolution:ts.ModuleResolutionKind.Bundler,target:ts.ScriptTarget.ES2024,
 lib:['lib.es2024.d.ts'],types:[]};
let roots;
let compilerOptions=options;
if (fixture) roots = [path.join(tree,'fixture.ts')];
else {
 const configPath = path.join(tree,'src/compiler/tsconfig.json');
 const config = ts.readConfigFile(configPath,ts.sys.readFile);
 if (config.error) throw Error(ts.flattenDiagnosticMessageText(config.error.messageText,'\n'));
 const parsed = ts.parseJsonConfigFileContent(config.config,ts.sys,path.dirname(configPath),options,configPath);
 if (parsed.errors.length) throw Error(JSON.stringify(parsed.errors));
 roots = [path.join(tree,'src/tsc/tsc.ts')];
 compilerOptions=parsed.options;
}
const program = ts.createProgram({rootNames:roots,options:compilerOptions}), checker = program.getTypeChecker();
// Broad source inventory, deliberately independent of checked-view admission.
function arrayShape(t, seen = new Set()) {
 if (seen.has(t)) return {array:false,readonly:false};
 seen = new Set(seen); seen.add(t);
 if (checker.isTupleType(t)) return {array:true,readonly:!!t.target.readonly};
 if (checker.isArrayType(t)) return {array:true,readonly:t.target?.symbol?.name==='ReadonlyArray'};
 let parts=[];
 if(t.isUnionOrIntersection()) parts=t.types;
 else if(t.flags & ts.TypeFlags.TypeParameter) {
  const c=checker.getBaseConstraintOfType(t); if(c) parts=[c];
 } else if(t.flags & ts.TypeFlags.Object) {
  try {parts=checker.getBaseTypes(t)||[];} catch {}
 }
 const shapes=parts.map(x=>arrayShape(x,seen));
 return {array:shapes.some(x=>x.array),readonly:shapes.some(x=>x.readonly)};
}
const roads = new Set(['at','values','map','forEach','some','every','filter','find','findIndex','findLast','findLastIndex','reduce','reduceRight','slice','concat','join','indexOf','includes','pop','shift','splice','reverse','sort','copyWithin','parallelMap']);
function unparen(n) {while(n.parent && ts.isParenthesizedExpression(n.parent)) n=n.parent; return n;}
function read(n) {
 const site=unparen(n),p=site.parent;
 if(p && ts.isBinaryExpression(p) && p.left===site && p.operatorToken.kind===ts.SyntaxKind.EqualsToken) return false;
 // Compound assignments and increments read and write the element.
 return true;
}
const rows=[];
for(const source of program.getSourceFiles().filter(f=>fixture?roots.includes(f.fileName):f.fileName.startsWith(path.join(tree,'src')+path.sep) && !f.isDeclarationFile).sort((a,b)=>a.fileName.localeCompare(b.fileName))) {
 function add(node,receiver,form) {
  if(process.env.V3_MUTANT_SKIP_INDEX==='1' && form==='index') return;
  const type=checker.getTypeAtLocation(receiver), a=arrayShape(type);
  if(!a.array) return;
  const p=source.getLineAndCharacterOfPosition(node.getStart(source));
  rows.push({file:path.relative(tree,source.fileName),line:p.line+1,column:p.character+1,form,
   readonly:a.readonly,
   receiver:receiver.getText(source),type:checker.typeToString(type,receiver,ts.TypeFormatFlags.NoTruncation)});
 }
 function visit(n) {
  if(ts.isTypeNode(n)) return;
  if(ts.isElementAccessExpression(n)&&read(n)) add(n,n.expression,'index');
  if(ts.isForOfStatement(n)) {
   // values() has its own row; do not also count its iterator as an array receiver.
   add(n,n.expression,'for-of');
  }
  if(ts.isSpreadElement(n)) add(n,n.expression,'spread');
  if(ts.isCallExpression(n)) {
   const e=n.expression;
   if(ts.isPropertyAccessExpression(e)&&roads.has(e.name.text)) add(n,e.expression,e.name.text);
   if(ts.isPropertyAccessExpression(e)&&ts.isIdentifier(e.expression)&&e.expression.text==='Array'&&e.name.text==='from'&&n.arguments[0]) add(n,n.arguments[0],'Array.from');
  }
  ts.forEachChild(n,visit);
 }
 visit(source);
}
fs.writeFileSync(out+'.jsonl.gz',zlib.gzipSync(rows.map(r=>JSON.stringify(r)).join('\n')+(rows.length?'\n':''),{mtime:0}));
function group(key) {
 const m={};
 if(key==='form') for(const k of ['index','at','for-of','spread','slice','map','forEach']) m[k]={upper_bound:0,readonly_subset:0};
 for(const r of rows) {const k=r[key];m[k]||={upper_bound:0,readonly_subset:0};m[k].upper_bound++;m[k].readonly_subset+=+r.readonly;}
 return Object.fromEntries(Object.entries(m).sort((a,b)=>b[1].upper_bound-a[1].upper_bound||a[0].localeCompare(b[0])));
}
const seven=new Set(['index','at','for-of','spread','slice','map','forEach']);
const summary={typescript:ts.version,seven_form_upper_bound:rows.filter(r=>seven.has(r.form)).length,seven_form_readonly_subset:rows.filter(r=>seven.has(r.form)&&r.readonly).length,label:'source upper bound on what the head can check; not admission',
 readonly_label:'narrower subset containing a readonly array or readonly tuple receiver arm; not admission',
 upper_bound:rows.length,readonly_subset:rows.filter(r=>r.readonly).length,forms:group('form'),files:group('file')};
fs.writeFileSync(out+'.json',JSON.stringify(summary,null,2)+'\n');
console.log(JSON.stringify(summary));
