'use strict';
const fs=require('node:fs'),path=require('node:path'),ts=require('/workspace/adamic/stage3/api/node_modules/typescript'),cp=require('node:child_process'),assert=require('node:assert/strict');
const s='/workspace/scratch/gov-a',source=fs.readFileSync('/workspace/scratch/gov/before/src/compiler/core.ts','utf8'),file=ts.createSourceFile('core.ts',source,ts.ScriptTarget.Latest,true);
const owner=name=>{const found=file.statements.filter(n=>ts.isFunctionDeclaration(n)&&n.name?.text===name&&n.body);assert.equal(found.length,1);return found[0].getText(file);};
const original=owner('getOwnValues'),adapted=original.replace('{\r\n    const values','{\r\n    if (isArray(collection)) return collection.filter(() => true);\r\n    const values').replace('(collection as MapLike<T>)[key]!','collection[key]!');assert.notEqual(original,adapted);
const prelude='export interface MapLike<T> { [index: string]: T; }\nconst hasOwnProperty = Object.prototype.hasOwnProperty;\n'+owner('isArray')+'\n';
const output=`console.log('filtered-length='+String(result.length)); for (const value of result) console.log(value === undefined ? 'undefined' : value);`;
const cases={sparse:`const a: string[] = []; console.log('before-write-length='+String(a.length)); a[5]='x'; console.log('after-write-length='+String(a.length)); const result=getOwnValues(a); ${output}`,presentUndefined:`const a: (string | undefined)[] = [undefined, 'x']; const result=getOwnValues(a); ${output}`,empty:`const a: string[] = []; const result=getOwnValues(a); ${output}`,record:`const a: MapLike<string> = {'1':'a', '0':'b'}; const result=getOwnValues(a); ${output}`};
const dir=s+'/witnesses';fs.mkdirSync(dir,{recursive:true});const results={originalOwner:original,adaptedOwner:adapted,isArrayOwner:owner('isArray'),cases:{}};
for(const [name,body] of Object.entries(cases)){
 let sides={};for(const [side,fn] of Object.entries({original,adapted})){
  const text=prelude+fn+'\n'+body+'\n',stem=path.join(dir,name+'-'+side);fs.writeFileSync(stem+'.ts',text);
  const js=ts.transpileModule(text,{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext}}).outputText;fs.writeFileSync(stem+'.mjs',js);
  const run=cp.spawnSync(process.execPath,[stem+'.mjs'],{encoding:'utf8'});fs.writeFileSync(stem+'.node.stdout',run.stdout);fs.writeFileSync(stem+'.node.stderr',run.stderr);sides[side]={exit:run.status,stdout:run.stdout,stderr:run.stderr};
 }assert.deepEqual(sides.original,sides.adapted);assert.equal(sides.original.exit,0);results.cases[name]=sides;
}
const controls={sparseWrite:`const a: string[]=[];console.log('before-write-length='+String(a.length));a[5]='x';console.log('after-write-length='+String(a.length));const result=a.filter(()=>true);${output}`,arrayArm:`function arrayArm<T>(collection:T[]):T[]{return collection.filter(()=>true);}const result=arrayArm(['x']);${output}`,recordArm:`function recordArm<T>(collection:MapLike<T>):T[]{const values:T[]=[];for(const key in collection){if(hasOwnProperty.call(collection,key))values.push(collection[key]!);}return values;}const result=recordArm<string>({'1':'a','0':'b'});${output}`};
for(const [name,body] of Object.entries(controls)){
 const text=prelude+body+'\n',stem=path.join(dir,name);fs.writeFileSync(stem+'.ts',text);fs.writeFileSync(stem+'.mjs',ts.transpileModule(text,{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext}}).outputText);
 const r=cp.spawnSync(process.execPath,[stem+'.mjs'],{encoding:'utf8'});assert.equal(r.status,0);fs.writeFileSync(stem+'.node.stdout',r.stdout);fs.writeFileSync(stem+'.node.stderr',r.stderr);
}
fs.writeFileSync(s+'/raw/node-witnesses.json',JSON.stringify(results,null,2)+'\n');
