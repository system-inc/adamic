'use strict';
const ts=require('/workspace/adamic/stage3/api/node_modules/typescript'),fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),cp=require('node:child_process');
const [root,output,mode='before']=process.argv.slice(2);
assert.equal(ts.version,'6.0.3');
const files=ts.sys.readDirectory(root+'/src',['.ts'],undefined,['**/*']);
const config=ts.readConfigFile(root+'/src/tsconfig-base.json',ts.sys.readFile);
const options=ts.parseJsonConfigFileContent(config.config,ts.sys,root+'/src').options;
Object.assign(options,{noEmit:true,types:['node','mocha','chai'],typeRoots:['/workspace/or-predicate-packages/node_modules/@types']});
const program=ts.createProgram(files,options),checker=program.getTypeChecker(),rows=[];
for(const sf of program.getSourceFiles())if(sf.fileName.startsWith(root+'/src/')){
 function visit(n){
  if(ts.isCallExpression(n)){
   let symbol=checker.getSymbolAtLocation(n.expression);if(symbol?.flags&ts.SymbolFlags.Alias)symbol=checker.getAliasedSymbol(symbol);
   const declarations=symbol?.declarations?.filter(d=>ts.isFunctionDeclaration(d)&&['or','orPredicate2','orPredicate3'].includes(d.name?.text)&&d.getSourceFile().fileName.endsWith('/compiler/core.ts'));
   if(declarations?.length){
    const sig=checker.getResolvedSignature(n),d=sig?.declaration,overloads=declarations.filter(d=>!d.body),owners=[];
    for(let p=n.parent;p;p=p.parent)if(ts.isFunctionLike(p)&&p.name)owners.unshift(p.name.getText(sf));
    const lc=sf.getLineAndCharacterOfPosition(n.getStart(sf));
    rows.push({file:path.relative(root,sf.fileName),line:lc.line+1,start:n.getStart(sf),end:n.end,name:symbol.name,overload:overloads.indexOf(d)+1,arguments:n.arguments.map(a=>a.getText(sf)),signature:checker.signatureToString(sig,n,ts.TypeFormatFlags.NoTruncation),declaration:d?.getText(),owners});
   }
  }
  ts.forEachChild(n,visit);
 }visit(sf);
}
const reference=JSON.parse(cp.execFileSync('git',['-C','/workspace/adamic','show','701181917b36db38528257ff37da1a4e64aaac16:stage3/census/or-calls/adapted-census.json'],{encoding:'utf8'})).rows;
const normalize=row=>JSON.stringify([row.file||row.site.split(':')[0],row.arguments.map(a=>typeof a==='string'?a:a.text)]);
const keys=rows.map(normalize),expected=reference.map(normalize);
assert.deepEqual([...keys].sort(),[...expected].sort(),'STOP: the 23-site list differs from the reference');
const used=new Map();
for(const row of rows){const key=normalize(row),occurrence=used.get(key)||0;used.set(key,occurrence+1);const r=reference.filter(r=>normalize(r)===key)[occurrence];assert(r);const predicate=r.overload<3;row.originalOverload=r.overload;
 if(mode==='before'){assert.equal(row.name,'or');assert.equal(row.overload,r.overload,'STOP: overload resolution changed');}
 else {assert.equal(row.name,predicate?'orPredicate'+(r.overload+1):'or');if(!predicate){assert.equal(row.overload,r.overload-2);assert.equal(row.declaration,r.declaration,'rest declaration changed');}assert(sigNotMissing(row));}
}
function sigNotMissing(r){return !!r.declaration&&!!r.signature;}
assert.equal(rows.filter(r=>r.originalOverload<3).length,19,'STOP: predicate site count changed');
const allDiagnostics=ts.getPreEmitDiagnostics(program);const diagnostics=allDiagnostics.map(d=>ts.flattenDiagnosticMessageText(d.messageText,'\n'));const diagnosticDetails=allDiagnostics.map(d=>{const lc=d.file&&d.start!==undefined?d.file.getLineAndCharacterOfPosition(d.start):undefined;return {code:d.code,file:d.file?path.relative(root,d.file.fileName):undefined,line:lc?lc.line+1:undefined,column:lc?lc.character+1:undefined,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')};});
fs.writeFileSync(output,JSON.stringify({version:ts.version,mode,sourceFiles:files.length,calls:rows.length,predicateCalls:19,restCalls:4,diagnosticCount:diagnostics.length,diagnostics,diagnosticDetails,rows},null,2)+'\n');
console.log(JSON.stringify({calls:rows.length,predicateCalls:19,restCalls:4,diagnosticCount:diagnostics.length}));
