'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),ts=require('typescript');
const {pins,guards,plan}=require('./adapt.cjs');
function instrument(text){
 const owners=require('./adapt.cjs').owners(text);const edits=[];
 for(const pin of pins){const o=owners.get(pin.owner);const source=ts.createSourceFile('debug.ts',text,ts.ScriptTarget.Latest,true);let found=[];
  function visit(n){if(ts.isArrowFunction(n)&&n.getStart(source)>=o.start&&n.end<=o.end&&n.body.getText(source).includes(pin.newMessage.slice(1,12)))found.push(n);ts.forEachChild(n,visit);}visit(source);assert.equal(found.length,1,pin.owner);const body=found[0].body;edits.push({start:body.getStart(source),end:body.end,text:`{ globalThis.__debugMessageCount(${JSON.stringify(pin.owner)}); return ${body.getText(source)}; }`});
 }
 for(const e of edits.sort((a,b)=>b.start-a.start))text=text.slice(0,e.start)+e.text+text.slice(e.end);return text;
}
function instantiate(text,count){
 const sf=ts.createSourceFile('debug.ts',text,ts.ScriptTarget.Latest,true);const imports=sf.statements.filter(ts.isImportDeclaration);const names=imports.flatMap(i=>i.importClause?.namedBindings?.elements?.filter(n=>!n.isTypeOnly).map(n=>n.name.text)||[]);const body=text.slice(imports.at(-1).end);const js=ts.transpileModule(body,{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.CommonJS}}).outputText;const exports={};
 const old=globalThis.__debugMessageCount;globalThis.__debugMessageCount=count;
 try{new Function('exports','ts',...names,js)(exports,ts,...names.map(n=>ts[n]));return exports.Debug;}finally{globalThis.__debugMessageCount=old;}
}
function observe(text,name,level,passing,counted){
 const counts=Object.fromEntries(pins.map(p=>[p.owner,0]));const hit=n=>{assert(Object.hasOwn(counts,n));counts[n]++;};const debug=instantiate(counted?instrument(text):text,hit);debug.setAssertionLevel(level);
 const node={kind:ts.SyntaxKind.Identifier};const predicate=name==='assertNotNode'?(passing?function testFalse(){return false}:function testTrue(){return true}):(passing?function testTrue(){return true}:function testFalse(){return false});
 const old=globalThis.__debugMessageCount;globalThis.__debugMessageCount=hit;
 let result;try{debug[name](name==='assertEachNode'?[node]:node,predicate);result={returned:true};}catch(e){result={threw:true,constructor:e.constructor.name,sameErrorConstructor:e.constructor===Error,prototypeIsError:Object.getPrototypeOf(e)===Error.prototype,message:e.message};}finally{globalThis.__debugMessageCount=old;}
 return {result,counts};
}
if(require.main===module){
 const [beforeTree,afterTree,output]=process.argv.slice(2);const before=fs.readFileSync(path.join(beforeTree,'src/compiler/debug.ts'),'utf8'),after=fs.readFileSync(path.join(afterTree,'src/compiler/debug.ts'),'utf8');assert.equal(plan(before).after,after,'only four exact message edits');
 const cases=[];
 for(const pin of pins)for(const level of [0,1,2,3])for(const passing of [false,true]){
  const b=observe(before,pin.owner,level,passing,true),a=observe(after,pin.owner,level,passing,true);assert.deepEqual(a.counts,b.counts);
  if(level>=1&&!passing){assert(b.result.threw&&a.result.threw&&b.result.sameErrorConstructor&&a.result.sameErrorConstructor&&b.result.prototypeIsError&&a.result.prototypeIsError);const name=pin.owner==='assertNotNode'?'testTrue':'testFalse';assert.equal(a.result.message,b.result.message.replace(` '${name}'`,''));assert.equal(a.counts[pin.owner],1);}else{assert.deepEqual(a.result,b.result);assert.equal(a.result.returned,true);assert.equal(a.counts[pin.owner],0);}
  assert.equal(Object.values(a.counts).reduce((x,y)=>x+y,0),level>=1&&!passing?1:0);cases.push({owner:pin.owner,level,passing,before:b,after:a});
 }
 fs.writeFileSync(output,JSON.stringify({cases,allPassed:true,counterControls:pins.map(p=>({owner:p.owner,before:cases.find(c=>c.owner===p.owner&&c.level===1&&!c.passing).before.counts[p.owner],after:cases.find(c=>c.owner===p.owner&&c.level===1&&!c.passing).after.counts[p.owner]}))},null,2)+'\n');console.log('32 assertion cases passed; four independent counters fire before and after');
}
module.exports={instrument,instantiate};
