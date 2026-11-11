const fs=require('fs'),path=require('path'),crypto=require('crypto'),ts=require('../../../api/node_modules/typescript');
const out=__dirname,normalize=text=>{
 const sf=ts.createSourceFile('typescript.d.ts',text,ts.ScriptTarget.Latest,true);
 const transformed=ts.transform(sf,[context=>{const visit=n=>{
  if(ts.isFunctionDeclaration(n)&&n.name?.text==='getLineAndCharacterOfPosition'){
   const p=n.parameters[0],first=ts.factory.updateParameterDeclaration(p,p.modifiers,p.dotDotDotToken,p.name,p.questionToken,ts.factory.createTypeReferenceNode('SourceFileLike'),p.initializer);
   return ts.factory.updateFunctionDeclaration(n,n.modifiers,n.asteriskToken,n.name,n.typeParameters,[first,...n.parameters.slice(1)],n.type,n.body);
  }
  return ts.visitEachChild(n,visit,context);
 };return n=>ts.visitNode(n,visit);}]);
 const printed=ts.createPrinter({newLine:ts.NewLineKind.LineFeed}).printFile(transformed.transformed[0]);transformed.dispose();return printed;
};
const base=normalize(fs.readFileSync(path.join(out,'original/typescript.d.ts'),'utf8')),results=[];
for(const label of ['original','a','b']){
 const dir=path.join(out,label),text=fs.readFileSync(path.join(dir,'typescript.d.ts'),'utf8');
 const consumer=path.join(dir,'public-sourcefile-consumer.ts');fs.writeFileSync(consumer,'import ts = require("./typescript");\ndeclare const file: ts.SourceFile;\nts.getLineAndCharacterOfPosition(file, 0);\n');
 const p=ts.createProgram([consumer],{strict:true,noEmit:true,target:ts.ScriptTarget.ES2020,module:ts.ModuleKind.CommonJS,types:[],skipLibCheck:true});
 const diagnostics=ts.getPreEmitDiagnostics(p).map(d=>{const pos=d.file&&d.start!=null?d.file.getLineAndCharacterOfPosition(d.start):null;return {code:d.code,file:d.file?path.basename(d.file.fileName):null,line:pos?pos.line+1:null,column:pos?pos.character+1:null,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')};});
 const sf=p.getSourceFile(consumer),checker=p.getTypeChecker();let fileType;function find(n){if(ts.isVariableDeclaration(n)&&n.name.getText(sf)==='file')fileType=checker.getTypeAtLocation(n.name);n.forEachChild(find);}find(sf);
 const signatures=JSON.parse(fs.readFileSync(path.join(dir,'public-signatures.json'),'utf8'));
 const result={label,only_getLineAndCharacterOfPosition_parameter_differs:normalize(text)===base,free_getLineStarts_declarations:signatures.functions.filter(f=>f.name==='getLineStarts').length,public_SourceFile_has_lineMap:!!checker.getPropertyOfType(fileType,'lineMap'),public_SourceFile_consumer_diagnostics:diagnostics};
 fs.writeFileSync(path.join(dir,'public-consumer-and-diff.json'),JSON.stringify(result,null,2)+'\n');results.push(result);
 if(label!=='original'){
  const cp=require('child_process'),log=fs.openSync(path.join(dir,'public-diff.log'),'w');const child=cp.spawnSync('diff',['-u',path.join(out,'original/typescript.d.ts'),path.join(dir,'typescript.d.ts')],{encoding:'utf8'});fs.closeSync(log);fs.writeFileSync(path.join(dir,'typescript-public.diff'),child.stdout||'');fs.writeFileSync(path.join(dir,'public-diff.log'),'diff child exit='+child.status+'\n'+(child.stderr||''));if(child.status!==1&&child.status!==0)throw Error('diff failed '+child.status);
 }
}
fs.writeFileSync(path.join(out,'public-comparison.json'),JSON.stringify(results,null,2)+'\n');console.log(JSON.stringify(results,null,2));
