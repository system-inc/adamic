const fs=require('fs'),path=require('path'),crypto=require('crypto'),ts=require('/workspace/scratch/scout-ty5myjt/adamic/stage3/api/node_modules/typescript');
const root='/workspace/scratch/scout-ty5myjt',out=root+'/adamic/stage3/census/scout-ty5myjt';
const wanted=new Set(['sortAndDeduplicateDiagnostics','sortAndDeduplicate','toSorted','deduplicateSorted','compareDiagnostics','compareDiagnosticsSkipRelatedInformation','compareRelatedInformation','compareMessageText','compareMessageChain','compareMessageChainSize','compareMessageChainContent','diagnosticsEqualityComparer','getDiagnosticFilePath','getDiagnosticCode','getDiagnosticMessage','messageTextEqualityComparer']);
const bodies=[];
for(const file of ['utilitiesPublic.ts','utilities.ts','core.ts']){
 const sf=ts.createSourceFile(file,fs.readFileSync(root+'/adapted/src/compiler/'+file,'utf8'),ts.ScriptTarget.Latest,true);
 function visit(n){if(ts.isFunctionDeclaration(n)&&wanted.has(n.name?.text))bodies.push({file,name:n.name.text,line:sf.getLineAndCharacterOfPosition(n.getStart(sf)).line+1,text:n.getText(sf)});ts.forEachChild(n,visit);}visit(sf);
}
fs.writeFileSync(out+'/sorter-call-chain.json',JSON.stringify(bodies,null,2)+'\n');
const rows=[];
for(const file of ['program.ts','utilitiesPublic.ts']){
 const emit=tree=>ts.transpileModule(fs.readFileSync(root+'/'+tree+'/src/compiler/'+file,'utf8'),{compilerOptions:{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.CommonJS,removeComments:false}}).outputText;
 const a=emit('adapted'),b=emit('diag-readonly-sorter'),hash=s=>crypto.createHash('sha256').update(s).digest('hex');
 rows.push({candidate:'diag-readonly-sorter',file,identical:a===b,before_bytes:Buffer.byteLength(a),after_bytes:Buffer.byteLength(b),before_sha256:hash(a),after_sha256:hash(b)});
}
fs.writeFileSync(out+'/diag-readonly-sorter-javascript.json',JSON.stringify(rows,null,2)+'\n');if(!rows.every(x=>x.identical))throw Error('JS changed');
