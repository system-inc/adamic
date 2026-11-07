'use strict';
// b5w3ycg, Nov 7 19:16Z, road (b): omit predicate function names only in lazy failure text.
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto');
const ts=require('typescript');
const {pins,guards}=require('./pins.json');
const hash=text=>crypto.createHash('sha256').update(text).digest('hex');
function owners(text){
 const source=ts.createSourceFile('debug.ts',text,ts.ScriptTarget.Latest,true);assert.equal(source.parseDiagnostics.length,0,'debug.ts parse error');
 const ns=source.statements.filter(n=>ts.isModuleDeclaration(n)&&n.name.text==='Debug');assert.equal(ns.length,1,'namespace Debug changed');
 const result=new Map();for(const node of ns[0].body.statements)if(ts.isFunctionDeclaration(node)&&node.body&&node.name){assert(!result.has(node.name.text),'duplicate implementation');result.set(node.name.text,{node,text:node.getText(source),start:node.getStart(source),end:node.end});}
 return result;
}
function inspect(text){
 const found=owners(text);
 for(const guard of guards){const owner=found.get(guard.owner);assert(owner,`missing ${guard.owner}`);assert.equal(owner.text,guard.text,`unchanged helper moved: ${guard.owner}`);assert.equal(hash(owner.text),guard.sha256);}
 return pins.map(pin=>{const owner=found.get(pin.owner);assert(owner,`missing ${pin.owner}`);const state=owner.text===pin.before?'before':owner.text===pin.after?'after':undefined;assert(state,`unreviewed owner: ${pin.owner}`);assert.equal(hash(owner.text),pin[state+'Sha256']);return {pin,owner,state};});
}
function javascript(text,removeComments){return ts.transpileModule(text,{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments}}).outputText;}
function plan(text){
 assert.equal(ts.version,'6.0.3');const sites=inspect(text);assert(sites.every(s=>s.state===sites[0].state),'half-adapted failure messages');
 if(sites[0].state==='after')return {after:text,edits:0};
 let after=text;
 for(const {pin,owner} of [...sites].sort((a,b)=>b.owner.start-a.owner.start)){
  assert.equal(owner.text.split(pin.oldMessage).length-1,1,`message count: ${pin.owner}`);
  assert.equal(owner.text.replace(pin.oldMessage,pin.newMessage),pin.after,'only message expression may change');
  after=after.slice(0,owner.start)+pin.after+after.slice(owner.end);
 }
 assert(inspect(after).every(s=>s.state==='after'));
 for(const removeComments of [false,true]){
  let expected=javascript(text,removeComments);
  const message=expression=>{const js=ts.transpileModule('const message = '+expression+';',{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext}}).outputText;const source=ts.createSourceFile('message.js',js,ts.ScriptTarget.Latest,true);return source.statements.find(ts.isVariableStatement).declarationList.declarations[0].initializer.getText(source);};
  const messages=new Map();
  for(const pin of pins){const old=message(pin.oldMessage),replacement=message(pin.newMessage);const entry=messages.get(old)||{count:0,replacement};assert.equal(entry.replacement,replacement);entry.count++;messages.set(old,entry);}
  for(const [old,entry] of messages){assert.equal(expected.split(old).length-1,entry.count,'emitted message count');expected=expected.replaceAll(old,entry.replacement);}
  assert.equal(javascript(after,removeComments),expected,'unreviewed JavaScript change');
 }
 for(const stripInternal of [false,true]){
  const emit=text=>ts.transpileDeclaration(text,{fileName:'debug.ts',compilerOptions:{stripInternal}}).outputText;
  assert.equal(emit(after),emit(text),'declarations changed');
 }
 return {after,edits:4};
}
function adapt(tree){const file=path.join(tree,'src/compiler/debug.ts'),before=fs.readFileSync(file,'utf8');const result=plan(before);assert.equal(fs.readFileSync(file,'utf8'),before,'concurrent edit');if(result.edits)fs.writeFileSync(file,result.after);return {files:result.edits?1:0,edits:result.edits,javascriptChanging:true};}
module.exports={adapt,plan,owners,inspect,pins,guards};
if(require.main===module){assert.equal(process.argv.length,3,'usage: node adapt.cjs TREE');console.log(JSON.stringify(adapt(path.resolve(process.argv[2])),null,2));}
