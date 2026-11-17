'use strict';
// Select outside-src programs by resolved import/reference reachability, not callee spelling.
const ts=require('../../api/node_modules/typescript'),fs=require('fs'),path=require('path');
const root=path.resolve(process.argv[2]),out=process.argv[3],core=root+'/src/compiler/core.ts';
const options={target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.NodeNext,moduleResolution:ts.ModuleResolutionKind.NodeNext,strict:true,skipLibCheck:true};
const files=ts.sys.readDirectory(root,['.ts','.tsx','.mts','.cts'],['**/node_modules/**','**/.git/**','src/**']);
const cache=ts.createModuleResolutionCache(root,x=>x,options),edges=new Map(),parseFailures=[];
function dependencies(file){if(edges.has(file))return edges.get(file);const next=[];edges.set(file,next);let text;try{text=fs.readFileSync(file,'utf8');}catch{return next;}
 const info=ts.preProcessFile(text,true,true);
 for(const ref of info.importedFiles){const resolved=ts.resolveModuleName(ref.fileName,file,options,ts.sys,cache).resolvedModule?.resolvedFileName;if(resolved)next.push(path.resolve(resolved));}
 for(const ref of info.referencedFiles){let name=path.resolve(path.dirname(file),ref.fileName);if(fs.existsSync(name))next.push(name);}
 for(const ref of info.typeReferenceDirectives){const resolved=ts.resolveTypeReferenceDirective(ref.fileName,file,options,ts.sys).resolvedTypeReferenceDirective?.resolvedFileName;if(resolved)next.push(path.resolve(resolved));}
 return next;
}
const todo=[...files],expanded=new Set(),reverse=new Map();
while(todo.length){const file=todo.pop();if(expanded.has(file))continue;expanded.add(file);for(const dep of dependencies(file)){if(!reverse.has(dep))reverse.set(dep,[]);reverse.get(dep).push(file);todo.push(dep);}}
const reaches=new Set(),pending=[core];while(pending.length){const file=pending.pop();if(reaches.has(file))continue;reaches.add(file);pending.push(...(reverse.get(file)||[]));}
const candidates=files.filter(f=>reaches.has(f));
console.log(JSON.stringify({files:files.length,candidates:candidates.length,modules:edges.size}));
fs.writeFileSync(out+'.selection.json',JSON.stringify({files,candidates,edges:Object.fromEntries(edges)},null,2));
const calls=[],program=ts.createProgram(candidates,options),c=program.getTypeChecker();let callExpressions=0;
for(const file of candidates){const f=program.getSourceFile(file),todo=[f];while(todo.length){const n=todo.pop();if(ts.isCallExpression(n)){callExpressions++;const d=c.getResolvedSignature(n)?.declaration;if(d?.getSourceFile().fileName===core&&d.name?.getText()==='filter'){const lc=f.getLineAndCharacterOfPosition(n.getStart());calls.push({file,line:lc.line+1,call:n.getText(),declaration:d.getText()});}}ts.forEachChild(n,x=>{todo.push(x);});}}
fs.writeFileSync(out,JSON.stringify({options,files,candidates,callExpressions,calls,edges:Object.fromEntries(edges)},null,2));console.log(JSON.stringify({files:files.length,candidates:candidates.length,calls:calls.length,callExpressions,modules:edges.size}));
