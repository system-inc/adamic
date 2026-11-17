'use strict';
const fs=require('node:fs'),path=require('node:path'),crypto=require('node:crypto');
const ts=require(process.env.TYPESCRIPT_API);
const tree=path.resolve(process.argv[2]),output=process.argv[3],core=tree+'/src/compiler/core.ts';
const options={strict:true,target:ts.ScriptTarget.ES2020,module:ts.ModuleKind.NodeNext,moduleResolution:ts.ModuleResolutionKind.NodeNext,skipLibCheck:true,types:['node'],typeRoots:[path.resolve(process.env.ADAMIC_TYPESCRIPT_SOURCE,'node_modules/@types')]};
const result={typescript:ts.version,tree,programs:{}};
for(const [mode,roots] of [['entry',[tree+'/src/tsc/tsc.ts']],['allSrc',ts.sys.readDirectory(tree+'/src',['.ts','.tsx'],undefined,['**/*'])]]){
 const p=ts.createProgram(roots,options),checker=p.getTypeChecker(),owner=p.getSourceFile(core),selected=new Map(),overloads={};
 for(const n of owner.statements)if(ts.isFunctionDeclaration(n)&&!n.body&&['arrayFrom','arrayToMap'].includes(n.name?.text)){const name=n.name.text,ordinal=(overloads[name]??=[]).length+1;overloads[name].push({ordinal,line:owner.getLineAndCharacterOfPosition(n.getStart(owner)).line+1,text:n.getText(owner)});selected.set(n,{name,ordinal});}
 const calls=[],sources=[],hashes={};let callExpressions=0;
 for(const sf of p.getSourceFiles()){
  if(!sf.fileName.startsWith(tree+'/')||sf.isDeclarationFile||sf.fileName.includes('/node_modules/'))continue;
  const relative=path.relative(tree,sf.fileName);sources.push(relative);hashes[relative]=crypto.createHash('sha256').update(fs.readFileSync(sf.fileName)).digest('hex');
  const stack=[sf];while(stack.length){const n=stack.pop();
   if(ts.isCallExpression(n)||ts.isNewExpression(n)||ts.isTaggedTemplateExpression(n)){
    callExpressions++;const signature=checker.getResolvedSignature(n),d=signature?.declaration,chosen=selected.get(d);
    if(chosen){const location=sf.getLineAndCharacterOfPosition(n.getStart(sf)),ancestors=[];for(let a=n.parent;a;a=a.parent)if(ts.isFunctionLike(a)){const loc=sf.getLineAndCharacterOfPosition(a.getStart(sf));ancestors.push({kind:ts.SyntaxKind[a.kind],name:a.name?.getText(sf),line:loc.line+1,column:loc.character+1,start:a.getStart(sf),end:a.end});}
     calls.push({set:chosen.name,overload:chosen.ordinal,file:relative,line:location.line+1,column:location.character+1,start:n.getStart(sf),end:n.end,call:n.getText(sf),signature:checker.signatureToString(signature,undefined,ts.TypeFormatFlags.NoTruncation),result:checker.typeToString(checker.getTypeAtLocation(n),undefined,ts.TypeFormatFlags.NoTruncation),mapped:(n.arguments?.length??0)===(chosen.name==='arrayFrom'?2:3),arguments:(n.arguments??[]).map(a=>({text:a.getText(sf),type:checker.typeToString(checker.getTypeAtLocation(a),undefined,ts.TypeFormatFlags.NoTruncation)})),ancestors});
    }
   }
   ts.forEachChild(n,c=>{stack.push(c);});
  }
 }
 const files=[...new Set(calls.map(c=>c.file))].sort(),summary={sources:sources.length,callExpressions,calls:calls.length,mapped:calls.filter(c=>c.mapped).length,files:files.length,perSet:Object.fromEntries(['arrayFrom','arrayToMap'].map(n=>[n,{calls:calls.filter(c=>c.set===n).length,mapped:calls.filter(c=>c.set===n&&c.mapped).length}]))};
 result.programs[mode]={roots:roots.map(f=>path.relative(tree,f)),summary,sources:sources.sort(),sourceHashes:hashes,files,overloads,calls};
 fs.writeFileSync(output+'.partial',JSON.stringify(result,null,2)+'\n');console.log(mode,JSON.stringify(summary));
}
const entry=new Set(result.programs.entry.sources);result.fullCensusFiles=[...new Set(['src/compiler/core.ts',...result.programs.allSrc.files])].sort().map(file=>({file,entryReach:entry.has(file),root:entry.has(file)?'src/tsc/tsc.ts':file}));fs.writeFileSync(output,JSON.stringify(result,null,2)+'\n');
