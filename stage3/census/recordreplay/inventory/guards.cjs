'use strict';
const fs=require('fs'),path=require('path'),crypto=require('crypto'),assert=require('assert/strict'),vm=require('vm');
const ts=require('/workspace/scratch/recordreplay/head/stage3/api/node_modules/typescript');
const D='/workspace/scratch/recordreplay',R=D+'/head',catalog=JSON.parse(fs.readFileSync(D+'/inventory-audit/catalog.json','utf8')),rows=[];
for(const entry of catalog.readers.filter(e=>/\.(?:cjs|mjs|js)$/.test(e.file))){
 const code=fs.readFileSync(path.join(R,entry.file),'utf8'),sf=ts.createSourceFile(entry.file,code,99,true,ts.ScriptKind.JS),defs=[],readerNames=new Set(['inventory','files','walk','manifest','snapshot']);
 for(const n of sf.statements){
  if(ts.isFunctionDeclaration(n)&&n.name)defs.push(n.getText(sf).replace(/^export\s+/,''));
  if(ts.isVariableStatement(n))for(const d of n.declarationList.declarations)if(ts.isIdentifier(d.name)&&d.initializer&&(ts.isArrowFunction(d.initializer)||ts.isFunctionExpression(d.initializer))){defs.push('const '+d.name.text+'='+d.initializer.getText(sf)+';');if(/readdirSync/.test(d.initializer.getText(sf)))readerNames.add(d.name.text);}
 }
 const scopes=new Set();
 function visit(n){
  if(ts.isCallExpression(n)&&/^assert(?:\.|$)/.test(n.expression.getText(sf))){
   let statement=n;while(statement.parent&&!ts.isSourceFile(statement.parent)&&!ts.isBlock(statement.parent))statement=statement.parent;
   const scope=statement.parent;if(!scope?.statements)return;const prior=[...scope.statements].slice(0,[...scope.statements].indexOf(statement));const variables=[],inventoryVariables=new Set();
   for(const p of prior){if(!ts.isVariableStatement(p))continue;const text=p.getText(sf);if(/writeFile|mkdir|copyFile|spawn|execFile|\.emit\(|symlink|unlink|rename/.test(text))continue;
    for(const d of p.declarationList.declarations){if(!d.initializer)continue;const init=d.initializer.getText(sf),name=d.name.getText(sf);if(/require\(/.test(init))continue;
     const reads=/readdirSync|process\.argv|path\.(?:resolve|join)|JSON\.parse/.test(init)||[...readerNames].some(k=>new RegExp('\\b'+k+'\\s*\\(').test(init))||[...inventoryVariables].some(k=>new RegExp('\\b'+k+'\\b').test(init));
     const localFn=ts.isBlock(scope)&&(ts.isArrowFunction(d.initializer)||ts.isFunctionExpression(d.initializer))&&/readdirSync/.test(init);
     if(reads||localFn){if(ts.isSourceFile(scope)&&(ts.isArrowFunction(d.initializer)||ts.isFunctionExpression(d.initializer)))continue;variables.push('const '+name+'='+init+';');if(reads||localFn)inventoryVariables.add(name);}
    }
   }
   const expr=n.getText(sf);const relevant=/inventory|source.?files|file.?names|names|roots|746/i.test(expr)||[...readerNames,...inventoryVariables].some(k=>new RegExp('\\b'+k+'\\b').test(expr));
   if(relevant&&variables.some(s=>/readdirSync|inventory|files\(|walk\(|manifest\(|snapshot\(/.test(s))&&!scopes.has(scope.pos)){
    scopes.add(scope.pos);const row={file:entry.file,line:sf.getLineAndCharacterOfPosition(n.getStart(sf)).line+1,guard:expr,method:'original first inventory guard with original read-only local declarations; historical driver not run',results:{}};
    for(const [side,after] of [['control',D+'/control'],['amended',D+'/candidate2']]){
     const readOnly=new Proxy(fs,{get(t,k){if(/^(?:write|append|mkdir|rm|rmdir|unlink|rename|copyFile|symlink|link|chmod|chown|truncate)/.test(String(k)))return()=>{throw Error('mutation forbidden in reader audit')};return t[k]}});
     const ctx={fs:readOnly,path,crypto,assert,ts,hash:x=>crypto.createHash('sha256').update(x).digest('hex'),before:D+'/control',after,reference:D+'/control',candidate:after,tree:after,root:after,__dirname:path.dirname(path.join(R,entry.file)),process:{argv:['node',entry.file,D+'/control',after,D+'/inventory-audit/unused'],env:{}},console:{log(){},error(){}}};
     try{vm.runInNewContext(defs.join('\n')+'\n'+variables.join('\n')+'\n'+expr,ctx,{timeout:20000});row.results[side]={status:'pass'};}catch(e){row.results[side]={status:e.name==='AssertionError'?'reject':'context_required',error:String(e.message).slice(0,500)};}
    }
    rows.push(row);
   }
  }
  ts.forEachChild(n,visit);
 }
 visit(sf);
}
fs.writeFileSync(D+'/inventory-audit/guards.json',JSON.stringify(rows,null,2)+'\n');console.log('original JS inventory guards exercised',rows.length);
