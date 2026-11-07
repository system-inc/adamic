'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto'),ts=require('typescript');
const {adapt,pins}=require('./adapt.cjs');
const [before,after,scratch,output]=process.argv.slice(2),hash=b=>crypto.createHash('sha256').update(b).digest('hex');
function files(root){const result=[];function walk(dir){for(const entry of fs.readdirSync(dir,{withFileTypes:true})){const name=path.join(dir,entry.name);if(entry.isDirectory())walk(name);else if(entry.isFile())result.push(path.relative(root,name));}}walk(root);return result.sort();}
for(const file of [pins.core.file,pins.harness.file]){const dest=path.join(scratch,file);fs.mkdirSync(path.dirname(dest),{recursive:true});fs.copyFileSync(path.join(before,file),dest);}
assert.equal(adapt(scratch).files,2);
const sourceFiles=files(path.join(before,'src'));assert.deepEqual(sourceFiles,files(path.join(after,'src')));const changed=[];
for(const file of sourceFiles){const a=fs.readFileSync(path.join(before,'src',file)),b=fs.readFileSync(path.join(after,'src',file));if(!a.equals(b)){changed.push('src/'+file);assert([pins.core.file,pins.harness.file].includes('src/'+file),'unexpected source delta '+file);assert(b.equals(fs.readFileSync(path.join(scratch,'src',file))),'full apply differs from isolated adaptation '+file);}}
assert.deepEqual(changed.sort(),[pins.core.file,pins.harness.file].sort());
const declarations=[];
for(const file of ['typescript.d.ts','tsserverlibrary.d.ts','typescript.internal.d.ts']){
 const a=fs.readFileSync(path.join(before,'built/local',file),'utf8'),b=fs.readFileSync(path.join(after,'built/local',file),'utf8');
 const removed='    /** @internal */\n    function getAllKeys(obj: object): string[];\n';
 if(file==='typescript.internal.d.ts'){assert.equal(a.split(removed).length,2,'missing unique declaration');assert.equal(b,a.replace(removed,''),'unexpected internal declaration delta');}else assert.equal(b,a,'public declarations changed');
 declarations.push({file,before:hash(a),after:hash(b),identical:a===b});
}
const javascript=files(path.join(before,'built/local')).filter(f=>f.endsWith('.js'));assert.deepEqual(javascript,files(path.join(after,'built/local')).filter(f=>f.endsWith('.js')));
function parse(file,text){const sf=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true,ts.ScriptKind.JS);assert.equal(sf.parseDiagnostics.length,0,'cannot parse emitted JS');return sf;}
function nodes(root,pred){const found=[];function visit(n){if(pred(n))found.push(n);ts.forEachChild(n,visit);}visit(root);return found;}
function functions(sf){return nodes(sf,n=>ts.isFunctionDeclaration(n)&&n.name?.text==='getAllKeys');}
let originalBody;
const emitted=javascript.map(file=>{
 const a=fs.readFileSync(path.join(before,'built/local',file),'utf8'),b=fs.readFileSync(path.join(after,'built/local',file),'utf8');const x=parse(file,a),y=parse(file,b),old=functions(x),now=functions(y);
 if(file==='typescript.js'){assert.equal(old.length,1);assert.equal(now.length,0);originalBody=old[0].body.getText(x);}
 if(file==='run.js'){assert.equal(old.length,1);assert.equal(now.length,1);assert.equal(old[0].body.getText(x),now[0].body.getText(y),'emitted function body changed');}
 if(file!=='run.js')assert.equal(now.length,0,'getAllKeys remains outside harness '+file);
 // Remove only getAllKeys's complete declaration and its named export entries.
 // All remaining runtime tokens must be identical. A changed token elsewhere fails.
 function rest(sf, adapted){const spans=functions(sf).map(n=>[n.getStart(sf),n.end]);
  for(const n of nodes(sf,n=>ts.isPropertyAssignment(n)&&n.name?.getText(sf)==='getAllKeys')){
   assert.equal(n.initializer.getText(sf),'() => getAllKeys','unreviewed export entry');let end=n.end;if(sf.text[end]===',')end++;spans.push([n.getStart(sf),end]);
  }
  for(const n of nodes(sf,n=>ts.isShorthandPropertyAssignment(n)&&n.name.text==='getAllKeys')){
   const assignment=n.parent.parent,paren=assignment.parent,guard=paren.parent;
   assert(ts.isBinaryExpression(assignment)&&assignment.operatorToken.kind===ts.SyntaxKind.EqualsToken&&assignment.left.getText(sf)==='module.exports','shorthand outside export annotation');
   assert(ts.isParenthesizedExpression(paren)&&ts.isBinaryExpression(guard)&&guard.operatorToken.kind===ts.SyntaxKind.AmpersandAmpersandToken&&guard.left.getText(sf)==='0','export annotation is not dead');
   let end=n.end;if(sf.text[end]===',')end++;spans.push([n.getStart(sf),end]);
  }
  if(file==='run.js'){
   const owners=nodes(sf,n=>ts.isVariableDeclaration(n)&&n.name.getText(sf)==='init_fourslashImpl');assert.equal(owners.length,1,'harness initializer owner');
   assert(ts.isCallExpression(owners[0].initializer)&&owners[0].initializer.expression.getText(sf)==='__esm','harness initializer is not __esm');
   const helper=nodes(sf,n=>ts.isVariableDeclaration(n)&&n.name.getText(sf)==='__esm');assert.equal(helper.length,1,'unique __esm helper');
   assert.equal(helper[0].initializer.getText(sf),'(fn, res) => function __init() {\n  return fn && (res = (0, fn[__getOwnPropNames(fn)[0]])(fn = 0)), res;\n}','__esm once-only body changed');
   const methods=nodes(owners[0],n=>ts.isMethodDeclaration(n)&&ts.isStringLiteral(n.name)&&n.name.text==='src/harness/fourslashImpl.ts');assert.equal(methods.length,1,'harness initializer method');
   const calls=methods[0].body.statements.filter(n=>ts.isExpressionStatement(n)&&n.getText(sf)==='init_ts8();');
   assert.equal(calls.length,adapted?2:1,'namespace initializer call count');
   if(adapted){const statements=methods[0].body.statements;assert.equal(statements.indexOf(calls[1]),statements.indexOf(calls[0])+1,'namespace calls not adjacent');spans.push([calls[1].getStart(sf),calls[1].end]);}
  }
  let text=sf.text;for(const [start,end] of spans.sort((a,b)=>b[0]-a[0]))text=text.slice(0,start)+text.slice(end);
  // The parser supplies template and regexp context; a raw scanner cannot rescan these tokens correctly.
  const cleaned=parse(file,text),tokens=[];
  function visit(n){if(n.kind>=ts.SyntaxKind.FirstJSDocNode&&n.kind<=ts.SyntaxKind.LastJSDocNode)return;const children=n.getChildren(cleaned);if(children.length)children.forEach(visit);else if(n.kind!==ts.SyntaxKind.EndOfFileToken&&n.kind!==ts.SyntaxKind.SyntaxList)tokens.push([n.kind,n.getText(cleaned)]);}
  visit(cleaned);return tokens;
 }
 const left=rest(x,false),right=rest(y,true);assert.equal(left.length,right.length,'unowned JavaScript token count changed '+file);for(let i=0;i<left.length;i++)assert.deepEqual(left[i],right[i],'unowned JavaScript token changed '+file+' at '+i);
 return {file,before:hash(a),after:hash(b),identical:a===b,beforeFunctions:old.length,afterFunctions:now.length};
});
assert(originalBody,'no baseline compiler body');const run=parse('run.js',fs.readFileSync(path.join(after,'built/local/run.js'),'utf8'));assert.equal(functions(run)[0].body.getText(run),originalBody,'compiler body differs from local harness body');
fs.writeFileSync(output,JSON.stringify({passed:true,sourceFiles:sourceFiles.length,changedSources:changed,declarations,javascript:emitted,emittedBodyByteIdentical:true,unownedRuntimeTokensIdentical:true,namedImportAddsOneAdjacentIdempotentInitializer:true},null,2)+'\n');console.log(JSON.stringify({passed:true,sourceFiles:sourceFiles.length,changedSources:changed,javascriptFiles:javascript.length,changedJavaScript:emitted.filter(r=>!r.identical).map(r=>r.file)}));
