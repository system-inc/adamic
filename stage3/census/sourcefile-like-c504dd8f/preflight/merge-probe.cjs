'use strict';
const fs=require('node:fs'),path=require('node:path'),ts=require(process.env.CENSUS_TYPESCRIPT || 'typescript');
const tree=path.resolve(process.argv[2]),file=path.join(tree,'src/compiler/types.ts');
const text=fs.readFileSync(file,'utf8'),source=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true);
const declarations=source.statements.filter(n=>ts.isInterfaceDeclaration(n)&&n.name.text==='SourceFile'&&n.members.length);
if(declarations.length!==1)throw Error('unexpected SourceFile owner');
const declaration=declarations[0],anchor=declaration.members.find(n=>n.name?.getText(source)==='text');
const member='getPositionOfLineAndCharacter?(line: number, character: number, allowEdits?: true): number;';
const like=source.statements.find(n=>ts.isInterfaceDeclaration(n)&&n.name.text==='SourceFileLike');
if(like.members.find(n=>n.name?.getText(source)==='getPositionOfLineAndCharacter').getText(source)!==member)throw Error('SourceFileLike drift');
const changed=text.slice(0,anchor.end)+'\r\n    '+member+text.slice(anchor.end);
const options={target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.ESNext,moduleResolution:ts.ModuleResolutionKind.Bundler,strict:true,noEmit:true,skipLibCheck:true,types:[]};
for(const adapted of [false,true]){
 const host=ts.createCompilerHost(options),read=host.readFile.bind(host);
 host.readFile=name=>path.resolve(name)===file?(adapted?changed:text):read(name);
 const program=ts.createProgram([path.join(tree,'src/services/services.ts')],options,host);
 const owner=program.getSourceFile(path.join(tree,'src/services/types.ts'));
 if(!owner)throw Error('services augmentation not loaded');
 const diagnostics=program.getSemanticDiagnostics(owner);
 const result=diagnostics.filter(d=>d.code===2386).map(d=>({code:d.code,file:path.relative(tree,d.file.fileName),line:d.file.getLineAndCharacterOfPosition(d.start).line+1,column:d.file.getLineAndCharacterOfPosition(d.start).character+1,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')}));
 console.log(JSON.stringify({typescript:ts.version,adapted,member,diagnostics:result}));
 if(result.length!==(adapted?1:0))throw Error('unexpected merge diagnostic count');
}
