'use strict';
const fs = require('node:fs');
const path = require('node:path');
const cp = require('node:child_process');
const crypto = require('node:crypto');
if (process.argv.length < 6) throw Error('usage: node census.cjs REPOSITORY ADAPTED_TREE SCRATCH OUTPUT');
const repo = path.resolve(process.argv[2]);
const tree = path.resolve(process.argv[3]);
const scratch = path.resolve(process.argv[4]);
fs.mkdirSync(scratch, {recursive:true});
const tsPath = path.join(repo, 'stage3/api/node_modules/typescript');
const ts = require(tsPath);
if (ts.version !== '6.0.3') throw Error('wrong TypeScript version');
const hash = text => crypto.createHash('sha256').update(text).digest('hex');
const walk = dir => fs.readdirSync(dir,{withFileTypes:true}).flatMap(e => e.isDirectory() ? walk(path.join(dir,e.name)) : [path.join(dir,e.name)]);
const roots = walk(path.join(tree,'src/compiler')).filter(p=>p.endsWith('.ts')).sort();
const options = {strict:true, noUncheckedIndexedAccess:true, exactOptionalPropertyTypes:true, erasableSyntaxOnly:false, verbatimModuleSyntax:true, allowImportingTsExtensions:true, noEmit:true, module:'esnext', moduleDetection:'force', moduleResolution:'bundler', target:'es2024', lib:['es2024'], types:[], skipLibCheck:false};
const original = fs.readFileSync(path.join(repo,'internal/load/prelude.d.ts'),'utf8');
let removed = 0;
const filtered = original.replace(/(interface (?:MapIterator|SetIterator)<T> \{\r?\n)([\t ]*next\(\): AdamicCollectionIteratorResult<T>;)/g, (_,start,line) => {removed++; return start + ' '.repeat(line.length);});
if(removed!==2) throw Error('expected exactly two iterator next lines');
const prelude = path.join(scratch,'prelude-without-collection-next.d.ts');
fs.writeFileSync(prelude,filtered);
const originalPrelude = path.join(scratch,'prelude-original.d.ts');
fs.writeFileSync(originalPrelude,original);
const libraryEdits = {
 'lib.es5.d.ts': [
  ['replacer: (substring: string, ...args: any[]) => string','replacer: (substring: string, ...args: any[]) => unknown'],
  ['interface RegExpExecArray extends Array<string>','interface RegExpExecArray extends Array<string | undefined>'],
  ['interface RegExpMatchArray extends Array<string>','interface RegExpMatchArray extends Array<string | undefined>'],
  ['split(separator: string | RegExp, limit?: number): string[];','split(separator: string, limit?: number): string[];\n    split(separator: RegExp, limit?: number): (string | undefined)[];']],
 'lib.es2021.string.d.ts': [['replacer: (substring: string, ...args: any[]) => string','replacer: (substring: string, ...args: any[]) => unknown']],
 'lib.es2018.regexp.d.ts': [['[key: string]: string','[key: string]: string | undefined']],
 'lib.es2015.symbol.wellknown.d.ts': [
  ['replacer: (substring: string, ...args: any[]) => string','replacer: (substring: string, ...args: any[]) => unknown'],
  ['[Symbol.split](string: string, limit?: number): string[]','[Symbol.split](string: string, limit?: number): (string | undefined)[]'],
  ['split(splitter: { [Symbol.split](string: string, limit?: number): (string | undefined)[]; }, limit?: number): string[];','split(splitter: { [Symbol.split](string: string, limit?: number): (string | undefined)[]; }, limit?: number): (string | undefined)[];']]
};
const patched = path.join(scratch,'loader-typescript');
fs.cpSync(path.join(tsPath,'lib'),path.join(patched,'lib'),{recursive:true});
for(const [name,edits] of Object.entries(libraryEdits)) {
 const file=path.join(patched,'lib',name); let text=fs.readFileSync(file,'utf8');
 for(const [before,after] of edits) {if(!text.includes(before))throw Error('library edit drift '+name);text=text.split(before).join(after);}
 fs.writeFileSync(file,text);
}
function label(p) {
 if(p.startsWith(tree+'/'))return path.relative(tree,p);
 if(p===prelude)return '<prelude-without-collection-next.d.ts>';
 if(p===originalPrelude)return '<prelude-original.d.ts>';
 if(p.startsWith(patched+'/lib/'))return '<loader-lib>/'+path.basename(p);
 if(p.startsWith(tsPath+'/lib/'))return '<stock-lib>/'+path.basename(p);
 return p;
}
function diagnostic(d) {
 const loc=d.file && d.start!==undefined ? d.file.getLineAndCharacterOfPosition(d.start):null;
 return {file:d.file?label(d.file.fileName):null,line:loc?loc.line+1:null,column:loc?loc.character+1:null,code:'TS'+d.code,message:ts.flattenDiagnosticMessageText(d.messageText,'\n'),...(d.relatedInformation?{relatedInformation:d.relatedInformation.map(diagnostic)}:{})};
}
// Program construction follows checker-census.cjs at 21f6268d.
if(roots.length!==79)throw Error('expected 79 roots, got '+roots.length);
const parsed=ts.parseJsonConfigFileContent({compilerOptions:options,files:roots.concat(prelude)},ts.sys,scratch);
if(parsed.errors.length)throw Error('invalid options');
const host=ts.createCompilerHost(parsed.options);
const libDir=path.join(patched,'lib');
host.getDefaultLibLocation=()=>libDir;
host.getDefaultLibFileName=opts=>path.join(libDir,ts.getDefaultLibFileName(opts));
const program=ts.createProgram({rootNames:roots.concat(prelude),options:parsed.options,host});
const baselineDiagnostics=ts.getPreEmitDiagnostics(program);
const checker=program.getTypeChecker();
function discarded(n) {
 const p=n.parent;
 if(ts.isExpressionStatement(p)||ts.isVoidExpression(p))return true;
 if(ts.isForStatement(p)&&(p.initializer===n||p.incrementor===n))return true;
 if(ts.isBinaryExpression(p)&&p.operatorToken.kind===ts.SyntaxKind.CommaToken)return p.left===n||discarded(p);
 // Parentheses carry no read of their own.
 if(ts.isParenthesizedExpression(p))return discarded(p);
 return false;
}
function form(n){
 const p=n.parent,k=ts.SyntaxKind;
 if(ts.isVariableDeclaration(p)&&p.initializer===n)return 'variable initializer';
 if((ts.isPropertyDeclaration(p)||ts.isPropertyAssignment(p))&&p.initializer===n)return 'property initializer';
 if(ts.isParameter(p)&&p.initializer===n)return 'other: parameter initializer';
 if(ts.isReturnStatement(p))return 'return statement';
 if(ts.isArrowFunction(p)&&p.body===n)return 'concise arrow body';
 if((ts.isCallExpression(p)||ts.isNewExpression(p))&&p.arguments?.includes(n))return 'call or new argument';
 if(ts.isBinaryExpression(p)){
  const op=p.operatorToken.kind;
  if([k.EqualsEqualsEqualsToken,k.ExclamationEqualsEqualsToken,k.EqualsEqualsToken,k.ExclamationEqualsToken].includes(op))return 'comparison operand';
  if([k.AmpersandAmpersandToken,k.BarBarToken,k.QuestionQuestionToken].includes(op))return 'condition or truthiness';
  if(op===k.CommaToken&&p.right===n)return 'comma right side whose comma is read';
  if(ts.isAssignmentOperator(op)&&p.right===n)return 'assignment right side';
  if(op===k.PlusToken&&(checker.getTypeAtLocation(p).flags&ts.TypeFlags.StringLike))return 'template or string concatenation';
 }
 if(((ts.isIfStatement(p)||ts.isWhileStatement(p)||ts.isDoStatement(p))&&p.expression===n)||(ts.isForStatement(p)&&p.condition===n))return 'condition or truthiness';
 if(ts.isPrefixUnaryExpression(p)&&p.operator===k.ExclamationToken)return 'condition or truthiness';
 if(ts.isConditionalExpression(p))return p.condition===n?'condition or truthiness':'conditional-expression branch';
 if(ts.isTemplateSpan(p))return 'template or string concatenation';
 if((ts.isPropertyAccessExpression(p)||ts.isElementAccessExpression(p))&&p.expression===n)return 'property access on it';
 return 'other: '+k[p.kind];
}
const sites=[],unionWithVoid=[];
let expressionCount=0,discardedVoid=0;
for(const file of roots){
 const sf=program.getSourceFile(file);if(!sf)throw Error('missing root '+file);
 function visit(n){
  if(ts.isTypeNode(n))return;
  if(ts.isExpressionNode(n)){
   expressionCount++;
   const type=checker.getTypeAtLocation(n);
   const pure=!!(type.flags&ts.TypeFlags.Void);
   const union=type.isUnion()&&type.types.some(t=>t.flags&ts.TypeFlags.Void);
   if(pure||union){
    if(discarded(n)){if(pure)discardedVoid++;}
    else {
     const start=n.getStart(sf),loc=sf.getLineAndCharacterOfPosition(start);
     const record={file:label(file),line:loc.line+1,column:loc.character+1,sourceLine:sf.text.split(/\r?\n/)[loc.line],expression:n.getText(sf),callee:ts.isCallExpression(n)?n.expression.getText(sf):null,syntaxKind:ts.SyntaxKind[n.kind],parentKind:ts.SyntaxKind[n.parent.kind],form:form(n),type:checker.typeToString(type),start,end:n.end};
     if(pure)record.judgment={kind:record.form==='concise arrow body'?'concise arrow body':'unclassified',reason:record.form==='concise arrow body'?'Reserved for the language lead.':'No checked byte-identical adaptation or proof of impossibility is established for this site. See runtime observations and the resolved declaration.'};
     if(ts.isCallExpression(n)){
      const signature=checker.getResolvedSignature(n),decl=signature?.declaration;
      if(decl){const source=decl.getSourceFile(),where=source.getLineAndCharacterOfPosition(decl.getStart(source));record.resolvedDeclaration={file:label(source.fileName),line:where.line+1,text:decl.getText(source)};}
     }
     (pure?sites:unionWithVoid).push(record);
    }
   }
  }
  ts.forEachChild(n,visit);
 }
 visit(sf);
}
const count=(rows,key)=>Object.fromEntries([...new Set(rows.map(key))].sort().map(k=>[k,rows.filter(r=>key(r)===k).length]));
const result={date:'2026-10-19',mainSha:cp.execFileSync('git',['merge-base','HEAD','origin/main'],{cwd:repo,encoding:'utf8'}).trim(),referenceCensusSha:'21f6268d6dad5d39c985558b8469473d4f68a303',typescriptVersion:ts.version,libraryProfile:'loader',options,prelude:{removedNextDeclarations:removed,originalSha256:hash(original),filteredSha256:hash(filtered)},roots:roots.map(f=>({file:label(f),sha256:hash(fs.readFileSync(f))})),expressionCount,discardedVoid,countsByForm:count(sites,r=>r.form),countsByJudgment:count(sites,r=>r.judgment.kind),sites,unionWithVoid};
fs.writeFileSync(process.argv[5],JSON.stringify(result,null,2)+'\n');
console.log(JSON.stringify({sites:sites.length,unionWithVoid:unionWithVoid.length,expressionCount,countsByForm:result.countsByForm}));
if(process.argv[6]==='--adaptations'){
 const edits=new Map(),candidates=new Map();
 for(const site of sites){
  if(site.judgment.kind==='concise arrow body')continue;
  const sf=program.getSourceFile(path.join(tree,site.file));
  let expression;
  function find(n){if(n.getStart(sf)===site.start&&n.end===site.end&&ts.isCallExpression(n))expression=n;else if(n.pos<=site.start&&n.end>=site.end)ts.forEachChild(n,find);}
  find(sf);if(!expression)continue;
  if(site.file==='src/compiler/checker.ts'&&site.expression==='originalHandler!(onlyUnreliable)'){
   const declaration='var outofbandVarianceMarkerHandler: ((onlyUnreliable: boolean) => void) | undefined;';
   const position=sf.text.indexOf(declaration);if(position<0)throw Error('marker handler declaration drift');
   const start=position+declaration.indexOf('void'),end=start+4,id=site.file+':'+start;
   const candidate={id,file:sf.fileName,start,end,replacement:'boolean',before:'void',declarationLine:sf.getLineAndCharacterOfPosition(position).line+1,expectedFlags:ts.TypeFlags.Boolean,reason:'All assignments to the marker handler either return true from assignments to unreliable/unmeasurable, or forward to the previous handler. Refine its callback return type to boolean.',implementations:[{file:site.file,line:25043,text:'outofbandVarianceMarkerHandler = onlyUnreliable => onlyUnreliable ? unreliable = true : unmeasurable = true;'},{file:site.file,line:23399,text:'outofbandVarianceMarkerHandler = onlyUnreliable => { propagatingVarianceFlags |= onlyUnreliable ? RelationComparisonResult.ReportsUnreliable : RelationComparisonResult.ReportsUnmeasurable; return originalHandler!(onlyUnreliable); };'}]};
   candidates.set(site,candidate);if(!edits.has(sf.fileName))edits.set(sf.fileName,new Map());edits.get(sf.fileName).set(id,{start,end,replacement:'boolean'});continue;
  }
  const decl=checker.getResolvedSignature(expression)?.declaration;
  if(!decl?.body||!ts.isBlock(decl.body)||!roots.includes(decl.getSourceFile().fileName))continue;
  let hasReturnValue=false;
  function returns(n){if(n!==decl.body&&ts.isFunctionLike(n))return;if(ts.isReturnStatement(n)&&n.expression)hasReturnValue=true;ts.forEachChild(n,returns);}
  returns(decl.body);if(hasReturnValue)continue;
  const file=decl.getSourceFile();
  let start,end,replacement;
  if(decl.type){if(decl.type.kind!==ts.SyntaxKind.VoidKeyword)continue;start=decl.type.getStart(file);end=decl.type.end;replacement='undefined';}
  else {const close=decl.getChildren(file).find(n=>n.kind===ts.SyntaxKind.CloseParenToken);if(!close)continue;start=end=close.end;replacement=': undefined';}
  const id=label(file.fileName)+':'+start;
  candidates.set(site,{id,file:file.fileName,start,end,replacement,before:decl.type?.getText(file)||'<inferred void>',declarationLine:file.getLineAndCharacterOfPosition(decl.getStart(file)).line+1});
  if(!edits.has(file.fileName))edits.set(file.fileName,new Map());
  edits.get(file.fileName).set(id,{start,end,replacement});
 }
 const overlay=new Map(),proofs={};
 for(const [file,changes] of edits){
  const original=program.getSourceFile(file).text;
  let changed=original;
  for(const edit of [...changes.values()].sort((a,b)=>b.start-a.start))changed=changed.slice(0,edit.start)+edit.replacement+changed.slice(edit.end);
  const emit=text=>ts.transpileModule(text,{fileName:file,compilerOptions:parsed.options}).outputText;
  const before=emit(original),after=emit(changed);
  if(before!==after)throw Error('annotation changed JavaScript '+file);
  overlay.set(file,changed);
  proofs[label(file)]={beforeSha256:hash(before),afterSha256:hash(after),byteIdentical:true,edits:[...changes.values()]};
 }
 const changedHost=ts.createCompilerHost(parsed.options),read=changedHost.readFile;
 changedHost.readFile=file=>overlay.get(file)??read(file);
 changedHost.getDefaultLibLocation=()=>libDir;
 changedHost.getDefaultLibFileName=opts=>path.join(libDir,ts.getDefaultLibFileName(opts));
 const changedProgram=ts.createProgram({rootNames:roots.concat(prelude),options:parsed.options,host:changedHost});
 const changedChecker=changedProgram.getTypeChecker();
 const diagnosticKey=d=>JSON.stringify([label(d.file?.fileName||''),d.code,ts.flattenDiagnosticMessageText(d.messageText,'\n')]);
 const oldDiagnostics=baselineDiagnostics,newDiagnostics=ts.getPreEmitDiagnostics(changedProgram);
 const oldCounts=new Map();for(const d of oldDiagnostics){const k=diagnosticKey(d);oldCounts.set(k,(oldCounts.get(k)||0)+1);}
 const introduced=[];for(const d of newDiagnostics){const k=diagnosticKey(d),count=oldCounts.get(k)||0;if(count)oldCounts.set(k,count-1);else introduced.push(diagnostic(d));}
 // An annotation is only evidence when it adds no checker errors.
 if(introduced.length===0){
  for(const [site,candidate] of candidates){
   const sf=changedProgram.getSourceFile(path.join(tree,site.file));
   const offset=[...(edits.get(sf.fileName)?.values()||[])].filter(e=>e.end<=site.start).reduce((sum,e)=>sum+e.replacement.length-(e.end-e.start),0);
   let found;
   function find(n){if(n.getStart(sf)===site.start+offset&&ts.isCallExpression(n))found=n;else if(n.pos<=site.start+offset&&n.end>=site.start+offset)ts.forEachChild(n,find);}
   find(sf);
   if(found&&changedChecker.getTypeAtLocation(found).flags&(candidate.expectedFlags||ts.TypeFlags.Undefined)){
    site.judgment={kind:'adaptation',reason:candidate.reason||'The resolved implementation has no value-returning statement. Annotate its return as undefined; the checker then types this call as undefined.',...(candidate.implementations?{implementations:candidate.implementations}:{}),rewrite:{file:label(candidate.file),declarationLine:candidate.declarationLine,start:candidate.start,end:candidate.end,before:candidate.before,after:candidate.replacement},proof:{file:label(candidate.file),beforeSha256:proofs[label(candidate.file)].beforeSha256,afterSha256:proofs[label(candidate.file)].afterSha256,byteIdentical:true},rewrittenType:changedChecker.typeToString(changedChecker.getTypeAtLocation(found))};
   }
  }
 }
 result.adaptationProof={method:'Batch annotation-only overlay; no source files repaired.',baselineDiagnostics:oldDiagnostics.length,rewrittenDiagnostics:newDiagnostics.length,introducedDiagnostics:introduced,files:proofs};
 result.countsByJudgment=count(sites,r=>r.judgment.kind);
 fs.writeFileSync(process.argv[5],JSON.stringify(result,null,2)+'\n');
 console.log(JSON.stringify(result.countsByJudgment));
}
