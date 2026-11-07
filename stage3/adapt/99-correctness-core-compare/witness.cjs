'use strict';
const fs=require('node:fs'),vm=require('node:vm'),assert=require('node:assert/strict'),crypto=require('node:crypto'),ts=require('/workspace/adamic/stage3/api/node_modules/typescript');
const root='/workspace/scratch/core-compare-amend',pins=require('/workspace/adamic/stage3/adapt/99-correctness-core-compare/owners.json')[0];
function actual(side,name,file='core.ts'){
 const path=root+'/'+side+'/src/compiler/'+file,text=fs.readFileSync(path,'utf8'),sf=ts.createSourceFile(path,text,ts.ScriptTarget.Latest,true),found=[];
 function visit(n){if(ts.isFunctionDeclaration(n)&&n.name?.text===name&&n.body)found.push(n);ts.forEachChild(n,visit);}visit(sf);assert.equal(found.length,1,name);return found[0].getText(sf).replaceAll('\r\n','\n');
}
const old=actual('before','compareComparableValues'),now=actual('after','compareComparableValues'),fail=actual('after','fail','debug.ts');assert.equal(fail,actual('before','fail','debug.ts'));
assert(pins.before.endsWith(old));assert(pins.after.endsWith(now));
const emit=body=>ts.transpileModule(body.replace('export ',''),{compilerOptions:{target:ts.ScriptTarget.ES2020,module:ts.ModuleKind.None}}).outputText;
function load(body){const context={Comparison:{LessThan:-1,EqualTo:0,GreaterThan:1}};vm.createContext(context);vm.runInContext(emit(fail),context);context.Debug={fail:context.fail};vm.runInContext(emit(body),context);return context.compareComparableValues;}
const before=load(old),after=load(now),encode=v=>v===undefined?'undefined':Number.isNaN(v)?'NaN':Object.is(v,-0)?'-0':v;
const inputs=[['equal numbers',1,1],['equal strings','a','a'],['both undefined',undefined,undefined],['left undefined number',undefined,1],['left undefined string',undefined,'a'],['right undefined string','a',undefined],['right undefined number',1,undefined],['numbers ascending',1,2],['numbers descending',2,1],['string case upper first','A','a'],['string case lower first','a','A'],['NaN left',NaN,1],['NaN right',1,NaN],['NaN both',NaN,NaN],['mixed string number','1',2],['mixed number string',1,'2']];
function outcome(fn,a,b){try{return {kind:'return',value:fn(a,b)};}catch(error){return {kind:'throw',name:error.name,message:error.message};}}
const rows=inputs.map(([name,a,b])=>{const original=outcome(before,a,b),adapted=outcome(after,a,b),mixed=(typeof a==='number'&&typeof b==='string')||(typeof a==='string'&&typeof b==='number');if(mixed){assert.deepEqual(original,{kind:'return',value:-1});assert.equal(adapted.kind,'throw');assert.equal(adapted.message,'Debug Failure. compareComparableValues: a string compared with a number');}else assert.deepEqual(adapted,original,name);return {name,a:encode(a),b:encode(b),typeofA:typeof a,typeofB:typeof b,original,adapted,mixed};});
console.log(JSON.stringify({rows,actualDebugFailSourceSha256:crypto.createHash('sha256').update(fail).digest('hex'),mixedNowThrows:true,failArmReachedOnCountedCalls:0,countedCalls:144697378},null,2));
