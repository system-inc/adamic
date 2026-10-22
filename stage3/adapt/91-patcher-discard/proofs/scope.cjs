'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto');
const ts=require('typescript');
const {targets}=require('../adapt.cjs');
assert.equal(ts.version,'6.0.3');
const [before,after,output]=process.argv.slice(2).map(p=>path.resolve(p));
const walk=(root,rel='src')=>fs.readdirSync(path.join(root,rel),{withFileTypes:true}).flatMap(e=>e.isDirectory()?walk(root,path.join(rel,e.name)):[path.join(rel,e.name)]).sort();
const files=walk(before);assert.deepEqual(walk(after),files);
const changed=files.filter(f=>!fs.readFileSync(path.join(before,f)).equals(fs.readFileSync(path.join(after,f))));
const wanted=targets.filter(t=>fs.readFileSync(path.join(before,t.file),'utf8')!==fs.readFileSync(path.join(after,t.file),'utf8'));
assert.deepEqual(changed,wanted.map(t=>t.file).sort());
const sites=[];
for(const t of wanted){
 const a=fs.readFileSync(path.join(before,t.file),'utf8'),b=fs.readFileSync(path.join(after,t.file),'utf8');
 const old=`fn => fn(${t.argument})`,replacement=`fn => { fn(${t.argument}); }`;
 assert.equal(a.split(old).length,2,'site uniqueness');assert.equal(b,a.replace(old,replacement),'change exceeds arrow body');
 const emits=[];
 for(const removeComments of [false,true]){
  const options={fileName:t.file,compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments}};
  const x=ts.transpileModule(a,options).outputText,y=ts.transpileModule(b,options).outputText;
  assert.notEqual(x,y);
  function mask(text){
   const sf=ts.createSourceFile(t.file+'.js',text,ts.ScriptTarget.Latest,true,ts.ScriptKind.JS);let body;
   function visit(n){if(ts.isCallExpression(n)&&n.expression.getText(sf)==='forEach'&&n.arguments[0]?.getText(sf)===t.registry){assert(!body);body=n.arguments[1].body;}ts.forEachChild(n,visit);}visit(sf);
   assert(body);return {text:text.slice(0,body.getStart(sf))+'BODY'+text.slice(body.end),body:body.getText(sf)};
  }
  const p=mask(x),q=mask(y);assert.equal(p.text,q.text,'emitted change exceeds arrow body');
  assert.equal(p.body,`fn(${t.argument})`);assert.equal(q.body,`{ fn(${t.argument}); }`);
  emits.push({removeComments,before:p.body,after:q.body});
  if(!removeComments){fs.writeFileSync(output+'.'+path.basename(t.file)+'.before.js.txt',x);fs.writeFileSync(output+'.'+path.basename(t.file)+'.after.js.txt',y);}
 }
 sites.push({file:t.file,before:`forEach(${t.registry}, ${old});`,after:`forEach(${t.registry}, ${replacement});`,emits});
}
fs.writeFileSync(output,JSON.stringify({typescript:ts.version,sourceFiles:files.length,changed,sites},null,2)+'\n');
console.log(JSON.stringify({sourceFiles:files.length,changed}));
