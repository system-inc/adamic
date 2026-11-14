'use strict';
const fs=require('fs'),cp=require('child_process'),path=require('path'),crypto=require('crypto'),assert=require('assert/strict');
const R='/workspace/scratch/recordreplay/head',D='/workspace/scratch/recordreplay',ts=require(R+'/stage3/api/node_modules/typescript');
const hash=s=>crypto.createHash('sha256').update(s).digest('hex');
function digest(text,name){
 const f=ts.createSourceFile('sys.ts',text,ts.ScriptTarget.Latest,true);let owners=[];function visit(n){if(ts.isFunctionDeclaration(n)&&n.name?.text==='getNodeSystem')owners.push(n);ts.forEachChild(n,visit);}visit(f);assert.equal(owners.length,1);
 const imports=f.statements.filter(ts.isImportDeclaration).filter(n=>n.moduleSpecifier.text.startsWith('node:'));
 if(name.startsWith('93'))imports.unshift(...f.statements.filter(ts.isImportDeclaration).filter(n=>n.moduleSpecifier.text==='./hostErrors.js'));
 return {ownerSha256:hash(owners[0].getText(f)),importsSha256:hash(imports.map(n=>n.getText(f)).join('\n'))};
}
const result={};
for(const name of ['93-builtin-imports','99-wasm-crypto-import']){
 const pin=JSON.parse(fs.readFileSync(path.join(R,'stage3/adapt',name,'proofs/targets.json'),'utf8'))[0];result[name]={pinsGetNodeSystemWhole:true,observations:{}};
 for(const side of ['before','after']){
  const text=name.startsWith('99')&&side==='after'?fs.readFileSync(D+'/candidate2/src/compiler/sys.ts','utf8'):cp.execFileSync('git',['-C',D+'/control','show',`refs/stage3/${name}-${side}:src/compiler/sys.ts`],{encoding:'utf8'});
  const d=digest(text,name);assert.equal(d.ownerSha256,pin['owner'+side[0].toUpperCase()+side.slice(1)+'Sha256']);assert.equal(d.importsSha256,pin['imports'+side[0].toUpperCase()+side.slice(1)+'Sha256']);result[name].observations[side]=d;
 }
}
const order=fs.readdirSync(R+'/stage3/adapt',{withFileTypes:true}).filter(e=>e.isDirectory()).map(e=>e.name).sort();assert(order.indexOf('93-builtin-imports')<order.indexOf('99-wasm-crypto-import'));assert.equal(order.at(-1),'99-wasm-crypto-import');result.order=order;result.noLaterAdapter=true;
fs.writeFileSync(process.argv[2],JSON.stringify(result,null,2)+'\n');console.log('93 and99 whole-owner pins verified;99 is last');
