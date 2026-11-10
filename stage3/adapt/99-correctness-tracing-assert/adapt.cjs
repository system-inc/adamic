'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto');
const ts=require('typescript'),pins=require('./pins.json');
function owner(text){
 const sf=ts.createSourceFile('tracing.ts',text,ts.ScriptTarget.Latest,true);assert.equal(sf.parseDiagnostics.length,0);
 const found=[];function walk(n){if(ts.isFunctionDeclaration(n)&&n.name?.text==='stopTracing'&&n.body)found.push(n);ts.forEachChild(n,walk);}walk(sf);
 assert.equal(found.length,1,'stopTracing owner drift');const n=found[0];return {text:n.getText(sf),start:n.getStart(sf),end:n.end};
}
function plan(text){
 assert.equal(ts.version,'6.0.3');const o=owner(text);
 const state=o.text===pins.before?'before':o.text===pins.after?'after':undefined;assert(state,'unreviewed or half-adapted stopTracing');
 assert.equal(crypto.createHash('sha256').update(o.text).digest('hex'),pins[state+'Sha256']);
 assert.equal(pins.before.replace('Debug.assert(tracing,','Debug.assert(!!tracing,'),pins.after,'only assertion argument may change');
 if(state==='after')return text;
 const changed=text.slice(0,o.start)+pins.after+text.slice(o.end);
 for(const removeComments of [true,false]){
  const emit=s=>ts.transpileModule(s,{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments}}).outputText;
  const before=emit(text);const expression='Debug.assert(tracing, "Tracing is not in progress")';
  assert.equal(before.split(expression).length-1,1,'emitted assertion drift');
  assert.equal(emit(changed),before.replace(expression,'Debug.assert(!!tracing, "Tracing is not in progress")'),'unreviewed JavaScript change');
 }
 for(const stripInternal of [true,false]){
  const emit=s=>ts.transpileDeclaration(s,{fileName:'tracing.ts',compilerOptions:{stripInternal}}).outputText;
  assert.equal(emit(changed),emit(text),'declarations changed');
 }
 return changed;
}
function adapt(tree){const file=path.join(tree,'src/compiler/tracing.ts'),before=fs.readFileSync(file,'utf8'),after=plan(before);assert.equal(fs.readFileSync(file,'utf8'),before);if(after!==before)fs.writeFileSync(file,after);return {edits:Number(after!==before),javascriptChanging:true};}
module.exports={owner,plan,adapt};
if(require.main===module){assert.equal(process.argv.length,3);console.log(JSON.stringify(adapt(process.argv[2])));}
