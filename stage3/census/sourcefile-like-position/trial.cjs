'use strict';
const fs=require('node:fs'), path=require('node:path'), assert=require('node:assert/strict'), crypto=require('node:crypto');
const ts=require('/workspace/scratch/sourcefile-like-position/stage3/api/node_modules/typescript');
const [before,after]=process.argv.slice(2);
const owner='src/compiler/types.ts';
const file=path.join(after,owner), raw=fs.readFileSync(file,'utf8');
const old='getPositionOfLineAndCharacter?(line: number, character: number, allowEdits?: true): number;';
const replacement='getPositionOfLineAndCharacter?: ((line: number, character: number, allowEdits?: true) => number) | undefined;';
assert.equal(raw.split(old).length,2);
fs.writeFileSync(file,raw.replace(old,replacement));
const all=(root,sub='src')=>fs.readdirSync(path.join(root,sub),{withFileTypes:true}).flatMap(e=>e.isDirectory()?all(root,path.join(sub,e.name)):[path.join(sub,e.name)]);
const names=all(before); assert.deepEqual(names,all(after));
let count=0, emitted=0, changed=[];
for(const name of names){
 const a=fs.readFileSync(path.join(before,name)),b=fs.readFileSync(path.join(after,name));
 if(!a.equals(b))changed.push(name);
 if(!/\.(tsx?|mts|cts)$/.test(name)||/\.d\.(ts|mts|cts)$/.test(name))continue;
 for(const removeComments of [false,true]){
  const emit=text=>ts.transpileModule(text,{fileName:name,compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments}}).outputText;
  const x=emit(a.toString()),y=a.equals(b)?x:emit(b.toString());
  assert.equal(x,y,`${name}: emitted JS changed (removeComments=${removeComments})`); emitted++;
 }
 count++;
}
assert.deepEqual(changed,[owner]);
console.log(JSON.stringify({status:'pass',typescript:ts.version,files:count,emits:emitted,changed,method:old,property:replacement},null,2));
