const fs = require('fs');
const path = require('path');
const crypto = require('crypto');
const child = require('child_process');
const pin = process.argv[2];
const ts = require(path.join(pin, 'lib/typescript.js'));
const sha = child.execFileSync('git', ['-C', pin, 'rev-parse', 'HEAD'], {encoding:'utf8'}).trim();
if (sha !== '050880ce59e30b356b686bd3144efe24f875ebc8') throw Error('wrong TypeScript pin');
const ranks = new Set(process.argv[3] ? process.argv[3].split(',').map(Number) : [7,22,31,36,51,52]);
const pairs = JSON.parse(fs.readFileSync(path.join(__dirname,'unknown-callable-pairs-ranked.json'),'utf8')).filter(p=>ranks.has(p.rank));
if(pairs.length!==ranks.size)throw Error("missing requested candidate rank");
const rows = [];
for (const pair of pairs) {
 const declarationFile=pair.type==='DirectoryStructureHost'?'src/compiler/watchUtilities.ts':pair.type==='ModuleResolutionCache'?'src/compiler/moduleNameResolver.ts':pair.type==='HostForUseSourceOfProjectReferenceRedirect'?'src/compiler/program.ts':pair.type==='IterationTypesResolver'?'src/compiler/checker.ts':pair.type==='typeof Debug'?'src/compiler/debug.ts':pair.type==='FormatDiagnosticsHost'?'src/compiler/program.ts':pair.type==='ProgramDiagnostics'?'src/compiler/programDiagnostics.ts':pair.type==='Scanner'?'src/compiler/scanner.ts':(pair.type==='System'||pair.type==='FileWatcher')?'src/compiler/sys.ts':pair.type==='ResolutionCacheHost'?'src/compiler/resolutionCache.ts':pair.type.includes('ts.performance')?'src/compiler/performance.ts':'src/compiler/types.ts';
 const declarationBytes=fs.readFileSync(path.join(pin,declarationFile));
 const types=ts.createSourceFile(declarationFile,declarationBytes.toString('utf8'),ts.ScriptTarget.Latest,true);
 const w = pair.witness;
 const bytes = fs.readFileSync(path.join(pin,w.file));
 const source = ts.createSourceFile(w.file,bytes.toString('utf8'),ts.ScriptTarget.Latest,true);
 let read,decl;
 const interfaces=new Map();
 let inheritancePath;
 let originalHeader;
 function visit(node) {
  if (ts.isPropertyAccessExpression(node) && node.name.text === pair.field || ts.isBindingElement(node) && node.propertyName && node.propertyName.getText(source) === pair.field) {
   const lc=source.getLineAndCharacterOfPosition(node.getStart(source));
   if (lc.line+1===w.line && lc.character+1===w.column) read=node;
  }
  ts.forEachChild(node,visit);
 }
 function declaration(node) {
  if((pair.type.includes('ts.performance')||pair.type==='typeof Debug')&&ts.isFunctionDeclaration(node)&&node.name?.text===pair.field&&node.body) {
   decl=node;
   originalHeader=types.text.slice(node.getStart(types),node.body.getStart(types)).trim();
   inheritancePath=[pair.type];
  }
  if (ts.isInterfaceDeclaration(node)) {
   const name=node.name.text;
   interfaces.set(name,[...(interfaces.get(name)||[]),node]);
  }
  ts.forEachChild(node,declaration);
 }
 function member(name,chain=[]) {
  if(chain.includes(name))return undefined;
  const declarations=interfaces.get(name)||[];
  for(const declaration of declarations) {
   const found=declaration.members.find(node=>(ts.isMethodSignature(node)||ts.isPropertySignature(node))&&node.name.getText(types)===pair.field);
   if(found)return {declaration:found,path:[...chain,name]};
  }
  for(const declaration of declarations)for(const clause of declaration.heritageClauses||[])for(const base of clause.types) {
   const found=member(base.expression.getText(types),[...chain,name]);
   if(found)return found;
  }
 }
 visit(source);declaration(types);
 const resolved=member(pair.type);
 if(resolved){decl=resolved.declaration;inheritancePath=resolved.path;}
 if (!read || !decl) throw Error('missing original witness '+pair.type+'.'+pair.field);
 rows.push({rank:pair.rank,type:pair.type,field:pair.field,candidateReads:pair.reads,witness:w,read:read.getText(source),readKind:ts.isBindingElement(read)?"binding":"property",call:ts.isBindingElement(read)?read.getText(source):read.parent.getText(source),utf16Start:read.getStart(source),utf16End:read.end,fileSha256:crypto.createHash('sha256').update(bytes).digest('hex'),declaration:originalHeader?originalHeader.replace(/^export function /,"")+";":decl.getText(types),declarationFile,declarationSha256:crypto.createHash('sha256').update(declarationBytes).digest('hex'),declarationLine:types.getLineAndCharacterOfPosition(decl.getStart(types)).line+1,...(inheritancePath.length>1?{declarationOwner:inheritancePath.at(-1),inheritancePath}:{}),...(originalHeader?{declarationHeader:originalHeader,declarationKind:"exported-function"}:ts.isPropertySignature(decl)?{declarationKind:"property-function"}:{})});
}
fs.writeFileSync(path.join(__dirname,process.argv[4] || 'aggregate-original-witnesses.json'),JSON.stringify({sourceSha:sha,basis:'original declarations and read spans; reduced adjacent helpers and data carriers',members:rows},null,2)+'\n');
console.log('Verified '+rows.length+' original declarations/read spans, '+rows.reduce((n,p)=>n+p.candidateReads,0)+' conservative candidate reads.');
