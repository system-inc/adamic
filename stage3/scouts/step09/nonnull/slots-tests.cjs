const assert=require('node:assert/strict'),runtime=require('./slot-runtime.cjs');
let checks=0;
function equal(a,b){assert.deepEqual(a,b);checks++;}
{
 const r=runtime(),o={x:undefined};r.object(o,[['x','first']]);equal(r.get(o,'x','read'),undefined);equal(r.stats.first.before_write,1);
 r.initPut(o,'x',undefined,'reset');r.put(o,'x',3);equal(r.get(o,'x','read'),3);equal(r.stats.reset.after_write,1);
 r.initPut(o,'x',undefined,'reset');equal(r.get(o,'x','read'),undefined);equal(r.stats.reset.before_write,1);
 const a={},b={};r.variableInit(a,undefined,'v');r.variableInit(b,undefined,'v');r.variablePut(a,4);r.variableRead(a,4,'a');r.variableRead(b,undefined,'b');equal([r.stats.v.after_write,r.stats.v.before_write],[1,1]);
}
{
 const r=runtime(),o={x:undefined};r.object(o,[['x','alias']]);const alias=o;let coercions=0;const k={[Symbol.toPrimitive](){coercions++;return 'x';}};r.get(alias,k,'alias');equal(coercions,1);equal(r.stats.alias.before_write,1);
 r.initPut(o,'x',undefined,'getter');Object.defineProperty(o,'x',{get(){r.put(this,'y',1);return undefined;},configurable:true});r.get(o,'x','get');equal(r.stats.getter.before_write,1);
 const p={x:undefined};r.object(p,[['x','proto']]);r.get(Object.create(p),'x','inherited');equal(r.stats.proto.before_write,1);
}
{
 const r=runtime(),o={x:undefined};r.object(o,[['x','copy']]);equal({...r.source(o,'spread')},{x:undefined});equal(r.stats.copy.before_write,1);
 r.initPut(o,'x',undefined,'define');r.define(o,'x',{value:7},'define');equal(r.descriptor(o,'x','descriptor').value,7);equal(r.stats.define.after_write,1);
 r.initPut(o,'x',undefined,'json');equal(r.json(o,undefined,undefined,'json'),'{}');equal(r.stats.json.before_write,1);
 const a=[undefined];r.object(a,[['0','array']]);r.call(a,'slice','slice')();equal(r.stats.array.opaque_calls,1);
}
// Each mutation changes the observations checked above, without a compiler error.
const mutants=[];
for(const name of ['omit-read','omit-write','invert-written','getter-writes-before-read']){
 const r=runtime(),o={x:undefined};r.object(o,[['x','mutant']]);
 if(name==='omit-read'){assert.throws(()=>{equal(r.stats.mutant.before_write,1);});}
 if(name==='omit-write'){o.x=4;r.get(o,'x','read');assert.throws(()=>equal(r.stats.mutant.after_write,1));}
 if(name==='invert-written'){r.put(o,'x',4);r.get(o,'x','read');assert.throws(()=>equal(r.stats.mutant.before_write,1));}
 if(name==='getter-writes-before-read'){
  const source=runtime.toString().replace('const snapshot={id:s.id,written:s.written,read:false};s.read=true;const value=o[p];','const value=o[p];const snapshot={id:s.id,written:s.written,read:false};s.read=true;');
  const changed=eval('('+source+')')(),object={};Object.defineProperty(object,'x',{get(){changed.put(object,'x',1);return undefined;},set(){},configurable:true});changed.object(object,[['x','g']]);changed.get(object,'x','read');assert.throws(()=>equal(changed.stats.g.before_write,1));
 }
 mutants.push({name,caught:'first-read observation assertion'});
}
console.log(JSON.stringify({checks,mutants},null,2));
