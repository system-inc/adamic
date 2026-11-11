'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto');
const ts=require('typescript'),pins=require('./pins.json');
function owners(text){
 const sf=ts.createSourceFile('tracing.ts',text,ts.ScriptTarget.Latest,true);assert.equal(sf.parseDiagnostics.length,0);
 const result=[];for(const pin of pins){const found=[];function walk(n){if(ts.isFunctionDeclaration(n)&&n.name?.text===pin.owner&&n.body)found.push(n);ts.forEachChild(n,walk);}walk(sf);
 assert.equal(found.length,1,pin.owner+' owner drift');const n=found[0];result.push({pin,text:n.getText(sf),start:n.getStart(sf),end:n.end});}return result;
}
function plan(text){
 assert.equal(ts.version,'6.0.3');const selected=owners(text);const states=selected.map(o=>o.text===o.pin.before?'before':o.text===o.pin.after?'after':undefined);
 assert(states.every(Boolean),'unreviewed pop or popAll');assert(states.every(s=>s===states[0]),'half-adapted owners');
 for(const [i,o] of selected.entries())assert.equal(crypto.createHash('sha256').update(o.text).digest('hex'),o.pin[states[i]+'Sha256']);
 assert.equal(pins[0].after,pins[0].before.replace('eventStack.length--;','eventStack.splice(eventStack.length - 1);'));
 assert.equal(pins[1].after,pins[1].before.replace('eventStack.length = 0;','eventStack.splice(0);'));
 if(states[0]==='after')return text;
 let changed=text;for(const o of selected.sort((a,b)=>b.start-a.start))changed=changed.slice(0,o.start)+o.pin.after+changed.slice(o.end);
 for(const removeComments of [true,false]){
  const emit=s=>ts.transpileModule(s,{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments}}).outputText;
  const before=emit(text);assert.equal(before.split('eventStack.length--;').length-1,1);assert.equal(before.split('eventStack.length = 0;').length-1,1);
  assert.equal(emit(changed),before.replace('eventStack.length--;','eventStack.splice(eventStack.length - 1);').replace('eventStack.length = 0;','eventStack.splice(0);'),'unreviewed JavaScript change');
 }
 for(const stripInternal of [true,false]){
  const emit=s=>ts.transpileDeclaration(s,{fileName:'tracing.ts',compilerOptions:{stripInternal}}).outputText;
  assert.equal(emit(changed),emit(text),'declarations changed');
 }return changed;
}
function adapt(tree){const file=path.join(tree,'src/compiler/tracing.ts'),before=fs.readFileSync(file,'utf8'),after=plan(before);assert.equal(fs.readFileSync(file,'utf8'),before);if(after!==before)fs.writeFileSync(file,after);return {edits:Number(after!==before)*2,javascriptChanging:true};}
module.exports={owners,plan,adapt};
if(require.main===module){assert.equal(process.argv.length,3);console.log(JSON.stringify(adapt(process.argv[2])));}
