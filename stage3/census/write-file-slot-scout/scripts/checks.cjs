const fs=require('fs'),path=require('path'),cp=require('child_process'),crypto=require('crypto'),ts=require('/workspace/scratch/write-file-slot-scout/stage3/api/node_modules/typescript');
const r='/workspace/scratch/write-file-slot-measure', names=['main','a-construction','b-slot'];
function files(d){return fs.readdirSync(d,{withFileTypes:true}).flatMap(e=>e.isDirectory()?files(path.join(d,e.name)):[path.join(d,e.name)]);}
const result={};
for(const n of names){const root=path.join(r,n), all=files(path.join(root,'src')), roots=files(path.join(root,'src/compiler')).filter(f=>f.endsWith('.ts')).sort();
 const config=path.join(r,n+'.tsconfig.json'); fs.writeFileSync(config,JSON.stringify({extends:path.join(root,'src/compiler/tsconfig.json'),compilerOptions:{noEmit:true,typeRoots:['/workspace/scratch/write-file-slot-scout/stage3/api/node_modules/@types'],tsBuildInfoFile:path.join(r,n+'.tsbuildinfo')},files:roots,include:[],references:[]},null,2));fs.writeFileSync(path.join(r,n+'.roots.json'),JSON.stringify(roots,null,2));
 const command=['node','/workspace/scratch/write-file-slot-scout/stage3/api/node_modules/typescript/lib/tsc.js','--project',config,'--pretty','false'];
 const checked=cp.spawnSync(command[0],command.slice(1),{encoding:'utf8',maxBuffer:50e6});fs.writeFileSync(path.join(r,n+'.tsc.log'),checked.stdout+checked.stderr);
 const changed=n==='main'?[]:all.filter(f=>!fs.readFileSync(f).equals(fs.readFileSync(path.join(r,'main',path.relative(root,f))))).map(f=>path.relative(root,f));
 const sys=fs.readFileSync(path.join(root,'src/compiler/sys.ts'),'utf8');
 const src=ts.createSourceFile('sys.ts',sys,ts.ScriptTarget.Latest,true), system=src.statements.find(s=>ts.isInterfaceDeclaration(s)&&s.name.text==='System').getText(src).replace(/\r\n/g,"\n");
 for(const removeComments of [false,true]){const emit=ts.transpileModule(sys,{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,removeComments}}).outputText;fs.writeFileSync(path.join(r,n+`.comments-${!removeComments}.js`),emit);}
 const declaration=ts.transpileDeclaration(sys,{compilerOptions:{stripInternal:true}});fs.writeFileSync(path.join(r,n+'.sys.d.ts'),declaration.outputText);
 result[n]={rootCount:roots.length,command,stockTscExit:checked.status,changedSourceFiles:changed,sysSha256:crypto.createHash('sha256').update(sys).digest('hex'),systemSha256:crypto.createHash('sha256').update(system).digest('hex'),sysDeclarationSha256:crypto.createHash('sha256').update(declaration.outputText).digest('hex'),transpileDeclarationDiagnosticCount:declaration.diagnostics.length};console.log(n,result[n]);}
fs.writeFileSync(path.join(r,'checks.json'),JSON.stringify(result,null,2)+'\n');
