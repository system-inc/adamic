const fs=require('fs'),path=require('path'),crypto=require('crypto'),ts=require('/workspace/adamic/stage3/api/node_modules/typescript');
const dir='/workspace/adamic/stage3/adapt/99-correctness-core-extend';fs.mkdirSync(dir,{recursive:true});
const file='/workspace/scratch/core-extend/before/src/compiler/core.ts',text=fs.readFileSync(file,'utf8').replace(/\r\n/g,'\n'),s=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true),variants=JSON.parse(fs.readFileSync('/workspace/scratch/core-extend/stock-probes.json'))[0].changes;
const hash=t=>crypto.createHash('sha256').update(t).digest('hex');
const owners=s.statements.filter(n=>ts.isFunctionDeclaration(n)&&['extend','copyProperties'].includes(n.name?.text)).map(n=>{const before=n.getText(s);let after=variants[n.name.text].replace(/\r\n/g,'\n');if(n.name.text==='copyProperties')after=after.replace('<T1 extends T2, T2>','<T2>');return {file:'core.ts',owner:n.name.text,beforeDeclarations:1,afterDeclarations:1,before,after,ownerBeforeSha256:hash(before),ownerAfterSha256:hash(after)};});
fs.writeFileSync(dir+'/owners.json',JSON.stringify(owners,null,2)+'\n');
let adapter=fs.readFileSync('/workspace/adamic/stage3/adapt/99-correctness-axolotl-k/adapt.cjs','utf8').replace('javascriptChanging: true','javascriptChanging: false');fs.writeFileSync(dir+'/adapt.cjs',adapter);
fs.writeFileSync(dir+'/README.md','# extend and copyProperties: type-only candidate\n\nMeasurement and full proof are in progress. No JavaScript-changing adaptation is implemented.\n');
console.log(owners.map(o=>({owner:o.owner,before:o.before,after:o.after})));
