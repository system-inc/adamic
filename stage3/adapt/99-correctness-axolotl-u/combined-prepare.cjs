'use strict';
const fs=require('fs'),cp=require('child_process'),path=require('path'),assert=require('assert/strict'),ts=require('/workspace/adamic/stage3/api/node_modules/typescript');
const adapter=require('/workspace/adamic/stage3/adapt/99-correctness-axolotl-u/adapt.cjs');const source='/workspace/scratch/99u-final31';
for(const name of ['99u-final31-main','99u-final31-counted']){
 const target='/workspace/scratch/'+name;cp.execFileSync('git',['clone','--shared',source,target],{stdio:'inherit'});const patch=cp.execFileSync('git',['-C',source,'diff','HEAD','--binary'],{maxBuffer:64*1024*1024});cp.execFileSync('git',['-C',target,'apply','--binary','-'],{input:patch});fs.cpSync(source+'/src',target+'/src',{recursive:true});
}
const baseline='/workspace/scratch/99u-final31-main';
for(const file of [...new Set(adapter.pins.map(p=>p.file))]){
 const full=baseline+'/src/compiler/'+file;let text=fs.readFileSync(full,'utf8');const found=adapter.owners(text,file);
 for(const pin of adapter.pins.filter(p=>p.file===file).sort((a,b)=>found.get(b.owner).start-found.get(a.owner).start)){const owner=found.get(pin.owner);assert.equal(adapter.hash(owner.text),pin.ownerAfterSha256);text=text.slice(0,owner.start)+pin.before.replace(/\n/g,'\r\n')+text.slice(owner.end);}
 if(file===adapter.imports.file)for(const name of adapter.imports.names){const exact='    '+name+',\r\n';assert.equal(text.split(exact).length,2);text=text.replace(exact,'');}
 fs.writeFileSync(full,text);
}
const texts=new Map([...new Set(adapter.pins.map(p=>p.file))].map(file=>[file,fs.readFileSync(baseline+'/src/compiler/'+file,'utf8')]));assert(adapter.inspect(texts).every(site=>site.state==='before'));
const manifest=JSON.parse(fs.readFileSync('/workspace/scratch/99u-guard-manifest.json','utf8'));const counted='/workspace/scratch/99u-final31-counted';
for(const file of [...new Set(manifest.map(p=>p.file))]){
 const full=counted+'/src/compiler/'+file;let text=fs.readFileSync(full,'utf8');
 for(const row of manifest.filter(p=>p.file===file)){
  const expression=`(globalThis as any).__debugTypeClaimCensus(${JSON.stringify(row.id)}, ${row.condition});`;
  if(row.id!=='G11'){
   const statement=`Debug.assert(${row.condition});`;assert.equal(text.split(statement).length,2,row.id);text=text.replace(statement,expression+'\r\n'+(row.id==='G01'?'            ':'                ')+statement);
  }else{const statement='return createExpressionError(node as Expression);';assert.equal(text.split(statement).length,2);text=text.replace(statement,expression+'\r\n                '+statement);text=text.replace('    isAsExpression,','    isExpression,\r\n    isAsExpression,');}
 }
 const parsed=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true);assert.equal(parsed.parseDiagnostics.length,0);fs.writeFileSync(full,text);
}
fs.writeFileSync('/workspace/scratch/99u-final31-preparation.json',JSON.stringify({base:'31d7a244808fa8092643926d21ab9d310a54cce3',source,baseline,counted,baselineMethod:'invert only all five exact pinned owners and their required imports from the completed full apply; all other main changes retained',countedMethod:'actual accepted assertions stay active; nonthrowing counters precede each assertion; rejected default proposal is counted without asserting'},null,2)+'\n');
