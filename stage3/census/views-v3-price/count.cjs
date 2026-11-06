#!/usr/bin/env node
// Measurement only. Fail closed on admission cases not yet ported from Go.
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
if (fixture) roots = [path.join(tree,'fixture.ts')];
else {
 const configPath = path.join(tree,'src/compiler/tsconfig.json');
 const config = ts.readConfigFile(configPath,ts.sys.readFile);
 if (config.error) throw Error(ts.flattenDiagnosticMessageText(config.error.messageText,'\n'));
 const parsed = ts.parseJsonConfigFileContent(config.config,ts.sys,path.dirname(configPath),options,configPath);
 if (parsed.errors.length) throw Error(JSON.stringify(parsed.errors));
 roots = parsed.fileNames;
}
const program = ts.createProgram({rootNames:roots,options}), checker = program.getTypeChecker();
const F=ts.TypeFlags;
function scalar(t) { return !!(t.flags & (F.StringLike|F.NumberLike|F.BooleanLike|F.EnumLike)); }
// 1 admitted, 0 rejected, -1 requires compiler representation/producer evidence.
function element(t, seen = new Set()) {
 if(seen.has(t)) return 1;
 seen = new Set(seen); seen.add(t);
 if(t.flags & F.Unknown) return 1;
 if(t.flags & (F.Any|F.Never|F.TypeParameter|F.BigIntLike|F.ESSymbolLike)) return 0;
 if(t.flags & F.Undefined || scalar(t)) return 1;
 if(t.flags & F.Null) return -1; // ViewNull requires the enclosing representation to be ir.Union.
 if(t.isUnion()) {
  const parts=t.types.map(x=>element(x,seen));
  return parts.includes(0)?0:parts.includes(-1)?-1:1;
 }
 if(checker.isArrayType(t)) return readonly(t)?element(checker.getTypeArguments(t)[0],seen):0;
 if(checker.isTupleType(t) || t.isIntersection()) return -1;
 if(t.flags & F.NonPrimitive) return 1;
 if(!(t.flags & F.Object)) return -1;
 if(checker.getSignaturesOfType(t,ts.SignatureKind.Call).length || checker.getSignaturesOfType(t,ts.SignatureKind.Construct).length) return -1;
 if(t.symbol?.flags & ts.SymbolFlags.Class) return 0;
 if(checker.getIndexInfosOfType(t).length) return 0;
 const properties=checker.getPropertiesOfType(t);
 const fields=properties.map(s=>({s,t:checker.getTypeOfSymbolAtLocation(s,s.valueDeclaration||s.declarations?.[0])}));
 // elementContractCheckable accepts a required scalar kind or literal discriminator.
 if(fields.some(({s,t})=>!(s.flags & ts.SymbolFlags.Optional) && scalar(t) && (s.name==='kind'||!!(t.flags & (F.StringLiteral|F.NumberLiteral|F.BooleanLiteral))))) return 1;
 const parts=fields.map(({t})=>t.flags & F.Unknown?0:element(t,seen));
 return parts.includes(0)?0:parts.includes(-1)?-1:1;
}
function readonly(t) { return t.target?.symbol?.name==='ReadonlyArray' && t.target.symbol.declarations?.some(d=>program.isSourceFileDefaultLibrary(d.getSourceFile())); }
function admission(t) {
 t=checker.getNonNullableType(t);
 if(!checker.isArrayType(t)) return {decision:0,reason:'not an array type'};
 if(!readonly(t)) return {decision:0,reason:'mutable array'};
 return {decision:element(checker.getTypeArguments(t)[0]),reason:'element contract'};
}
const roads = new Set(['at','values','map','forEach','some','every','filter','find','findIndex','findLast','findLastIndex','reduce','reduceRight','slice','concat','join','indexOf','includes','pop','shift','splice','reverse','sort','fill','copyWithin','parallelMap']);
function unparen(n) {while(n.parent && ts.isParenthesizedExpression(n.parent)) n=n.parent; return n;}
function read(n) {
 const site=unparen(n),p=site.parent;
 if(p && ts.isBinaryExpression(p) && p.left===site && p.operatorToken.kind===ts.SyntaxKind.EqualsToken) return false;
 // Compound assignments and increments read and write the element.
 return true;
}
const rows=[];
for(const source of program.getSourceFiles().filter(f=>fixture?roots.includes(f.fileName):f.fileName.startsWith(path.join(tree,'src/compiler')+path.sep)).sort((a,b)=>a.fileName.localeCompare(b.fileName))) {
 function add(node,receiver,form) {
  if(process.env.V3_MUTANT_SKIP_INDEX==='1' && form==='index') return;
  const type=checker.getTypeAtLocation(receiver), a=admission(type);
  if(a.decision===0) return;
  const p=source.getLineAndCharacterOfPosition(node.getStart(source));
  rows.push({file:path.relative(tree,source.fileName),line:p.line+1,column:p.character+1,form,
   decision:a.decision===1?'candidate':'unresolved',reason:a.reason,
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
function group(key) {const m={}; for(const r of rows) {const k=r[key];m[k]=(m[k]||0)+1;} return Object.fromEntries(Object.entries(m).sort((a,b)=>b[1]-a[1]||a[0].localeCompare(b[0])));}
const summary={typescript:ts.version,status:'provisional: not a certified ceiling',candidates:rows.filter(r=>r.decision==='candidate').length,unresolved:rows.filter(r=>r.decision==='unresolved').length,forms:group('form'),files:group('file')};
fs.writeFileSync(out+'.json',JSON.stringify(summary,null,2)+'\n');
console.log(JSON.stringify(summary));
