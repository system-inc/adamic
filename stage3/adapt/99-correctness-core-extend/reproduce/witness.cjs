const fs=require('fs'),ts=require('/workspace/adamic/stage3/api/node_modules/typescript'),assert=require('assert/strict');
const base='/workspace/scratch/core-extend',payload={},symbol=Symbol('witness');
function load(side){const file=base+'/'+side+'/src/compiler/core.ts',text=fs.readFileSync(file,'utf8'),s=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true),owners=s.statements.filter(n=>ts.isFunctionDeclaration(n)&&['extend','copyProperties'].includes(n.name?.text));const out=ts.transpileModule(owners.map(n=>n.getText(s)).join('\n'),{compilerOptions:{target:ts.ScriptTarget.ES2020,module:ts.ModuleKind.CommonJS}}).outputText;return new Function('hasOwnProperty','const exports={};'+out+';return exports;')(Object.prototype.hasOwnProperty);}
const before=load('before'),after=load('after');
const classes={plain:()=>[{a:payload},{b:2,a:1}],nullPrototype:()=>[Object.assign(Object.create(null),{a:payload}),Object.assign(Object.create(null),{b:2})],inheritedEnumerable:()=>[Object.assign(Object.create({inherited:1}),{a:payload}),Object.assign(Object.create({inherited2:2}),{b:3})],accessor:()=>[{get a(){return payload;}},{get b(){return 2;}}],symbolKeys:()=>[{a:payload,[symbol]:1},{b:2,[symbol]:3}],array:()=>[[payload],[1,2]],sparseArray:()=>[Object.assign(new Array(4),{2:payload}),Object.assign(new Array(8),{6:1})]};
function snapshot(x){return Reflect.ownKeys(x).map(k=>[k,x[k]]);}
const rows=[];for(const [classification,make]of Object.entries(classes))for(const owner of ['extend','copyProperties']){
 const [a,b]=make(),[c,d]=make();let x=before[owner](a,b),y=after[owner](c,d);if(owner==='copyProperties'){assert.equal(x,undefined);assert.equal(y,undefined);x=a;y=c;}
 const xs=snapshot(x),ys=snapshot(y);assert.equal(xs.length,ys.length);xs.forEach(([k,v],i)=>{assert.equal(k,ys[i][0]);assert(Object.is(v,ys[i][1]));});
 rows.push({classification,owner,equal:true,keys:xs.map(([k])=>String(k))});
}console.log(JSON.stringify({pass:true,rows},null,2));
