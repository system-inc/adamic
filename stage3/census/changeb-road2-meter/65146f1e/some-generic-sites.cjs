// Stock TypeScript API inventory: resolved core.some identity, first argument's type.
const fs = require('fs');
const path = require('path');
const ts = require('typescript');
if (ts.version !== '6.0.3') throw new Error(`expected 6.0.3, got ${ts.version}`);
const root = '/workspace/scratch/tree';
const out = path.resolve(__dirname);
const config = path.join(root, 'src/compiler/tsconfig.json');
const raw = ts.readConfigFile(config, ts.sys.readFile);
if (raw.error) throw new Error(ts.flattenDiagnosticMessageText(raw.error.messageText, '\n'));
const parsed = ts.parseJsonConfigFileContent(raw.config, ts.sys, path.dirname(config), {
  noEmit: true,
  typeRoots: ['/workspace/adamic/stage3/api/node_modules/@types'],
}, config);
if (parsed.errors.length) throw new Error(JSON.stringify(parsed.errors));
const program = ts.createProgram(parsed.fileNames, parsed.options);
const checker = program.getTypeChecker();
const sourceFiles = program.getSourceFiles().filter(s => s.fileName.startsWith(path.join(root, 'src/compiler/') ) && !s.isDeclarationFile);
if (sourceFiles.length !== 79) throw new Error(`expected 79 files, got ${sourceFiles.length}`);
const core = program.getSourceFile(path.join(root, 'src/compiler/core.ts'));
const decls = core.statements.filter(n => ts.isFunctionDeclaration(n) && n.name?.text === 'some');
if (decls.length !== 3) throw new Error(`expected 3 core.some declarations, got ${decls.length}`);
const some = checker.getSymbolAtLocation(decls[0].name);
function unalias(s) {
  const seen = new Set();
  while (s && (s.flags & ts.SymbolFlags.Alias) && !seen.has(s)) {seen.add(s);s=checker.getAliasedSymbol(s);}
  return s;
}
function where(n) {
  const sf = n.getSourceFile();
  const lc = sf.getLineAndCharacterOfPosition(n.getStart(sf));
  return `${path.relative(root,sf.fileName)}:${lc.line+1}:${lc.character+1}`;
}
const fmt = ts.TypeFormatFlags.NoTruncation | ts.TypeFormatFlags.UseAliasDefinedOutsideCurrentScope;
function text(t, n) {return checker.typeToString(t,n,fmt);}
// Match the compiler's mentionsTypeParameter traversal: instantiable type flags,
// union/intersection members, alias arguments and reference arguments. In
// particular readonly T[], NodeArray<T> and optional unions retain their T.
function genericComponents(t, at, seen = new Set(), route = 'argument') {
  if (seen.has(t)) return [];
  seen.add(t);
  const found=[];
  if (t.flags & ts.TypeFlags.Instantiable) found.push({route,type:text(t,at),flags:t.flags,typeParameter:!!(t.flags & ts.TypeFlags.TypeParameter)});
  const children=[];
  if (t.isUnionOrIntersection()) t.types.forEach((x,i)=>children.push([x,`member[${i}]`]));
  (t.aliasTypeArguments || []).forEach((x,i)=>children.push([x,`aliasArgument[${i}]`]));
  if ((t.flags & ts.TypeFlags.Object) && (t.objectFlags & ts.ObjectFlags.Reference)) {
    checker.getTypeArguments(t).forEach((x,i)=>children.push([x,`typeArgument[${i}]`]));
  }
  for (const [x,label] of children) found.push(...genericComponents(x,at,seen,`${route}.${label}`));
  return found;
}
const all=[];const unresolvedSome=[];
for (const sf of sourceFiles) {
  function visit(n) {
    if (ts.isCallExpression(n)) {
      const sig = checker.getResolvedSignature(n);
      const d = sig?.getDeclaration();
      const resolvedFromCore = d && decls.includes(d);
      const sym = unalias(checker.getSymbolAtLocation(ts.isPropertyAccessExpression(n.expression) ? n.expression.name : n.expression));
      const symbolFromCore = sym === some;
      if (resolvedFromCore || symbolFromCore) {
        if (!sig || !resolvedFromCore) unresolvedSome.push(where(n));
        const arg=n.arguments[0];
        const type=arg && checker.getTypeAtLocation(arg);
        const generic=type ? genericComponents(type,arg) : [];
        const predicate=sig && checker.getTypePredicateOfSignature(sig);
        all.push({where:where(n),file:path.relative(root,sf.fileName),line:sf.getLineAndCharacterOfPosition(n.getStart(sf)).line+1,call:n.getText(sf),argument:arg?.getText(sf) || null,argumentType:type ? text(type,arg):null,argumentCount:n.arguments.length,generic:generic.length>0,genericComponents:generic,resolvedDeclaration:d ? where(d):null,resolvedSignature:sig ? checker.signatureToString(sig,n,fmt):null,predicate:predicate ? {kind:predicate.kind,parameterIndex:predicate.parameterIndex,type:predicate.type ? text(predicate.type,n):null}:null,identity:{resolvedFromCore:!!resolvedFromCore,symbolFromCore}});
      }
    }
    ts.forEachChild(n,visit);
  }
  visit(sf);
}
all.sort((a,b)=>a.file.localeCompare(b.file)||a.line-b.line||a.where.localeCompare(b.where));
if (unresolvedSome.length) throw new Error(`unresolved core.some calls: ${unresolvedSome.join(', ')}`);
const generic=all.filter(r=>r.generic);
const result={typescriptVersion:ts.version,sourceRoot:root,config,options:parsed.options,files:sourceFiles.length,definition:'Calls resolving to core.ts some; generic first-argument type contains an instantiable type through union/intersection, alias arguments or reference arguments, matching mentionsTypeParameter. Includes one-argument predicate and two-argument boolean overloads separately. This is a source-site inventory, independent of census eligibility and first-error truncation; guard exposure is not proof the call stops.',coreDeclarations:decls.map(where),allSomeCalls:all.length,genericArgumentCalls:generic.length,genericPredicateCalls:generic.filter(r=>r.predicate).length,genericBooleanCalls:generic.filter(r=>!r.predicate).length,diagnostics:{options:program.getOptionsDiagnostics().map(d=>ts.flattenDiagnosticMessageText(d.messageText,'\n')),syntactic:program.getSyntacticDiagnostics().length,semantic:program.getSemanticDiagnostics().length, semanticDetails:program.getSemanticDiagnostics().map(d=>({code:d.code,where:d.file && d.start !== undefined ? (()=>{const p=d.file.getLineAndCharacterOfPosition(d.start);return `${path.relative(root,d.file.fileName)}:${p.line+1}:${p.character+1}`;})():null,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')}))},sites:generic};
fs.writeFileSync(path.join(out,'some-generic-sites.json'),JSON.stringify(result,null,2)+'\n');
fs.writeFileSync(path.join(out,'some-all-sites.json'),JSON.stringify(all,null,2)+'\n');
const esc=s=>String(s).replaceAll('|','&#124;').replaceAll('\n',' ');
const lines=['| Location | Argument | Stock argument type | Overload | Generic components |','| --- | --- | --- | --- | --- |'];
for (const r of generic) lines.push('| '+[r.where,r.argument,r.argumentType,r.predicate?'type predicate':'boolean (two arguments)',r.genericComponents.map(x=>x.type).join(', ')].map(esc).join(' | ')+' |');
fs.writeFileSync(path.join(out,'some-generic-sites.md'),lines.join('\n')+'\n');
console.log(JSON.stringify({...result,options:undefined,sites:undefined},null,2));
