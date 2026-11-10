const fs=require('fs'),path=require('path');
const ts=require('/workspace/adamic/stage3/api/node_modules/typescript');
const root='/workspace/scratch/abrq61p-adapted';
const files=[];
function walk(dir){for(const e of fs.readdirSync(dir,{withFileTypes:true})){const p=path.join(dir,e.name);if(e.isDirectory())walk(p);else if(/\.tsx?$/.test(p))files.push(p);}}
walk(root+'/src');files.sort();
const options={target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.NodeNext,moduleResolution:ts.ModuleResolutionKind.NodeNext,skipLibCheck:true,allowJs:false};
const host={getScriptFileNames:()=>files,getScriptVersion:()=> '0',getScriptSnapshot:f=>fs.existsSync(f)?ts.ScriptSnapshot.fromString(fs.readFileSync(f,'utf8')):undefined,getCurrentDirectory:()=>root,getCompilationSettings:()=>options,getDefaultLibFileName:o=>ts.getDefaultLibFilePath(o),fileExists:ts.sys.fileExists,readFile:ts.sys.readFile,readDirectory:ts.sys.readDirectory,directoryExists:ts.sys.directoryExists,getDirectories:ts.sys.getDirectories};
const service=ts.createLanguageService(host);const program=service.getProgram();
const where=(f,pos)=>{const s=program.getSourceFile(f);const l=s.getLineAndCharacterOfPosition(pos);return path.relative(root,f)+':'+(l.line+1)+':'+(l.character+1)};
const result=[];
for(const f of files.filter(f=>f.startsWith(root+'/src/compiler/'))){const source=program.getSourceFile(f);function visit(n){if(ts.isFunctionDeclaration(n)&&n.name&&n.typeParameters?.length){
 const row={where:where(f,n.getStart(source)),name:n.name.text,references:[]};
 const groups=service.findReferences(f,n.name.getStart(source))||[];
 for(const group of groups)for(const r of group.references){if(r.isDefinition)continue;
 const sf=program.getSourceFile(r.fileName);if(!sf||!r.fileName.startsWith(root+'/src/'))continue;
 let node=ts.getTokenAtPosition(sf,r.textSpan.start),a=node,unit=node;
 while(a.parent&&a.parent!==sf){a=a.parent;unit=a;}
 let call=node;while(call.parent&&(ts.isPropertyAccessExpression(call.parent)||ts.isElementAccessExpression(call.parent)||ts.isParenthesizedExpression(call.parent)))call=call.parent;
 let target=call;while(ts.isParenthesizedExpression(target))target=target.expression;
 const direct=call.parent&&ts.isCallExpression(call.parent)&&call.parent.expression===call&&(target===node||(ts.isPropertyAccessExpression(target)&&target.name===node)||(ts.isElementAccessExpression(target)&&ts.isStringLiteralLike(node)&&target.argumentExpression===node));
 row.references.push({where:where(r.fileName,r.textSpan.start),unit:where(r.fileName,unit.getStart(sf)),call:!!direct,definition:!!r.isDefinition});
 }
 row.references.sort((a,b)=>a.where.localeCompare(b.where)); result.push(row);
 console.log(row.where,row.references.length);
}ts.forEachChild(n,visit);}visit(source);}
fs.writeFileSync('/workspace/scratch/abrq61p/references.json',JSON.stringify({typescript:ts.version,source_files:files.length,units:result},null,2)+'\n');
service.dispose();
