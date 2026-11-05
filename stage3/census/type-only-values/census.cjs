const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const ts = require(process.env.CENSUS_TYPESCRIPT || path.resolve(__dirname, '../../api/node_modules/typescript'));
if (ts.version !== '6.0.3') throw Error(ts.version);
if (process.argv.length !== 4) throw Error('usage: node census.cjs ADAPTED_TREE OUTPUT_JSON');
const tree = path.resolve(process.argv[2]);
const roots = ts.sys.readDirectory(path.join(tree,'src/compiler'), ['.ts'], undefined, ['**/*']).sort();
const options = {strict:true, exactOptionalPropertyTypes:true, noUncheckedIndexedAccess:true, verbatimModuleSyntax:true, erasableSyntaxOnly:false, allowImportingTsExtensions:true, noEmit:true, module:ts.ModuleKind.ESNext, moduleDetection:ts.ModuleDetectionKind.Force, moduleResolution:ts.ModuleResolutionKind.Bundler, target:ts.ScriptTarget.ES2024, lib:['lib.es2024.d.ts'], types:[]};
const program = ts.createProgram(roots, options), checker = program.getTypeChecker();
if(program.getSyntacticDiagnostics().length) throw Error('parse errors');
const rel = p => path.relative(tree,p).replaceAll('\\','/');
function typePosition(n) {
  while ((ts.isQualifiedName(n.parent) && n.parent.left === n) || (ts.isPropertyAccessExpression(n.parent) && n.parent.expression === n)) n = n.parent;
  if (ts.isTypeQueryNode(n.parent)) return true;
  if (ts.isExportSpecifier(n.parent)) return n.parent.isTypeOnly || n.parent.parent.parent.isTypeOnly;
  return ts.isPartOfTypeNode(n);
}
function kind(s) {
 const f=s.flags, F=ts.SymbolFlags;
 if(f & F.ConstEnum) return 'const enum';
 if(f & F.RegularEnum) return 'enum';
 if(f & F.Class) return 'class';
 if(f & F.Function) return 'function';
 if(f & F.ValueModule) return 'namespace with values';
 if(f & F.Variable) return 'variable';
 return 'other value';
}
const rows=[], declarations=[], unused=[], unresolved=[], forms={};
const edits=new Map();
for(const file of roots) {
 const sf=program.getSourceFile(file), imports=new Map(), decls=[];
 for(const d of sf.statements) {
  if(!ts.isImportDeclaration(d) || !d.importClause || d.importClause.isTypeOnly) continue;
  const c=d.importClause, specs=[];
  if(c.name) specs.push({name:c.name,isTypeOnly:false,form:'default'});
  if(c.namedBindings) {
   if(ts.isNamedImports(c.namedBindings)) specs.push(...c.namedBindings.elements);
   else specs.push({name:c.namedBindings.name,isTypeOnly:false,form:'namespace'});
  }
  const ds={d, specs:[]}; decls.push(ds);
  for(const sp of specs) {
   if(sp.isTypeOnly) continue;
   const alias=checker.getSymbolAtLocation(sp.name);
   if(!alias) throw Error('no alias '+file);
   const target=checker.getAliasedSymbol(alias);
   const rec={sp,alias,target,refs:[],ds}; ds.specs.push(rec); imports.set(alias,rec);
   forms[sp.form || 'named']=(forms[sp.form || 'named']||0)+1;
  }
 }
 function visit(n) {
  if(ts.isImportDeclaration(n)) return;
  if(ts.isIdentifier(n)) {
   let symbol=checker.getSymbolAtLocation(n);
   if(ts.isShorthandPropertyAssignment(n.parent) && n.parent.name===n) symbol=checker.getShorthandAssignmentValueSymbol(n.parent) || symbol;
   if(ts.isExportSpecifier(n.parent) && !n.parent.parent.parent.moduleSpecifier) symbol=checker.getExportSpecifierLocalTargetSymbol(n.parent) || symbol;
   const r=imports.get(symbol);
   if(r) r.refs.push({line:sf.getLineAndCharacterOfPosition(n.getStart(sf)).line+1, type:typePosition(n), context:n.parent.getText(sf).slice(0,240)});
  }
  ts.forEachChild(n,visit);
 }
 visit(sf);
 for(const ds of decls) {
  const candidates=ds.specs.filter(r=>(r.target.flags & ts.SymbolFlags.Value) && r.refs.length && r.refs.every(x=>x.type));
  if(!candidates.length) continue;
  const remaining=ds.specs.filter(r=>!candidates.includes(r));
  const whole=remaining.length===0;
  const declLine=sf.getLineAndCharacterOfPosition(ds.d.getStart(sf)).line+1;
  const edge={file:rel(file),line:declLine,module:ds.d.moduleSpecifier.text,resolvedModule:rel(ts.resolveModuleName(ds.d.moduleSpecifier.text,file,options,ts.sys).resolvedModule.resolvedFileName),wholeDeclaration:whole,remaining:remaining.map(r=>({name:r.sp.name.text,valueReferences:r.refs.filter(x=>!x.type).length,references:r.refs.length}))};
  declarations.push(edge);
  for(const r of candidates) {
   rows.push({file:rel(file),line:sf.getLineAndCharacterOfPosition(r.sp.name.getStart(sf)).line+1,importedName:(r.sp.propertyName||r.sp.name).text,localName:r.sp.name.text,kind:kind(r.target),symbolFlags:r.target.flags,module:edge.module,declarationLine:declLine,wholeDeclaration:whole,references:r.refs});
  }
  const positions=edits.get(file)||[];
  for(const r of candidates) positions.push(r.sp.getStart(sf));
  edits.set(file,positions);
 }
 for(const r of imports.values()) if(r.target.name==='unknown') unresolved.push({file:rel(file),name:r.sp.name.text,references:r.refs});
 for(const r of imports.values()) if(!r.refs.length) unused.push({file:rel(file),name:r.sp.name.text,kind:kind(r.target)});
}
// Check the proposed inline type annotations only in memory; catch missed value uses.
const overlay=new Map();
for(const [file,positions] of edits) {
 let text=program.getSourceFile(file).text;
 for(const pos of positions.sort((a,b)=>b-a)) text=text.slice(0,pos)+'type '+text.slice(pos);
 overlay.set(file,text);
}
const host=ts.createCompilerHost(options), read=host.readFile.bind(host);
host.readFile=f=>overlay.get(f)??read(f);
const checked=ts.createProgram(roots,options,host);
for(const [file,text] of overlay) if(checked.getSourceFile(file).text !== text) throw Error('overlay not applied: '+file);
const before=program.getSemanticDiagnostics(), after=checked.getSemanticDiagnostics();
const diagKey=d=>JSON.stringify([d.file && rel(d.file.fileName),d.code,ts.flattenDiagnosticMessageText(d.messageText,'\n')]);
const baseKeys=new Set(before.map(diagKey));
const newDiagnostics=after.filter(d=>!baseKeys.has(diagKey(d)));
if(newDiagnostics.length) throw Error('new diagnostics: '+newDiagnostics.map(diagKey).join('\n'));
const prohibited=after.filter(d=>[1361,1362].includes(d.code));
if(prohibited.length) throw Error(ts.formatDiagnosticsWithColorAndContext(prohibited,{getCanonicalFileName:x=>x,getCurrentDirectory:()=>tree,getNewLine:()=> '\n'}));
const byKind={}; for(const r of rows) byKind[r.kind]=(byKind[r.kind]||0)+1;
const result={typescript:ts.version,roots:roots.map(rel),options,files:roots.map(f=>({file:rel(f),sha256:crypto.createHash('sha256').update(fs.readFileSync(f)).digest('hex')})),summary:{total:rows.length,byKind,wholeDeclarations:declarations.filter(x=>x.wholeDeclaration).length,mixedDeclarations:declarations.filter(x=>!x.wholeDeclaration).length,wholeSpecifiers:rows.filter(x=>x.wholeDeclaration).length,mixedSpecifiers:rows.filter(x=>!x.wholeDeclaration).length},forms,unresolved,unused,validation:{beforeDiagnostics:before.length,afterInlineTypeDiagnostics:after.length,TS1361or1362:prohibited.length,newDiagnosticMessages:newDiagnostics.length},declarations,rows};
fs.writeFileSync(process.argv[3],JSON.stringify(result,null,2)+'\n');
console.log(JSON.stringify({summary:result.summary,whole:declarations.filter(x=>x.wholeDeclaration),forms,unresolved,unused,validation:result.validation},null,2));
