const fs=require('node:fs'),path=require('node:path');
const ts=require('/tmp/family5-land/stage3/api/node_modules/typescript');
const base='/tmp/family5-base-tree',out=process.argv[2],kind=process.argv[3];
fs.mkdirSync(out,{recursive:true});
const skip=new Set(['.git','node_modules','built','tests']);
for(const name of fs.readdirSync(base))if(!skip.has(name))fs.cpSync(path.join(base,name),path.join(out,name),{recursive:true,dereference:false});
fs.symlinkSync('/tmp/family5-land/stage3/typescript/node_modules',path.join(out,'node_modules'),'dir');
fs.symlinkSync(path.join(base,'tests'),path.join(out,'tests'),'dir');
const edits=[];
if(kind==='a'||kind==='b'){
 const file='src/compiler/types.ts',p=path.join(out,file),before=fs.readFileSync(p,'utf8');
 const crlf=before.includes('\r\n'),nl=crlf?'\r\n':'\n';
 const anchor='export interface SourceFile extends Declaration, LocalsContainer {';
 const member=kind==='a'?'/** @internal */ getPositionOfLineAndCharacter?(line: number, character: number, allowEdits?: true): number;':'/** @internal */ getPositionOfLineAndCharacter?: ((line: number, character: number, allowEdits?: true) => number) | undefined;';
 if(before.split(anchor).length!==2)throw Error('SourceFile owner count');
 const after=before.replace(anchor,anchor+nl+'    '+member);fs.writeFileSync(p,after);edits.push({file,owner:'SourceFile',member,publicOwner:true,typeOnly:true});
}else if(kind==='c'){
 const sites=JSON.parse(fs.readFileSync('/workspace/family5-analysis/sites.json','utf8')).filter(x=>x.group==='G02'),files=new Map();
 for(const s of sites){const rel=s.where.split('/src/')[1],m=rel.match(/^(.*):(\d+):(\d+)$/),file='src/'+m[1];if(!files.has(file))files.set(file,[]);files.get(file).push({...s,line:+m[2],col:+m[3]});}
 for(const [file,ss] of files){const p=path.join(out,file),before=fs.readFileSync(p,'utf8'),sf=ts.createSourceFile(file,before,ts.ScriptTarget.Latest,true),changes=[];
  for(const s of ss){const start=sf.getPositionOfLineAndCharacter(s.line-1,s.col-1);let expr;
   function visit(n){if(n.getStart(sf)===start&&(ts.isIdentifier(n)||ts.isCallExpression(n)||ts.isPropertyAccessExpression(n)||ts.isElementAccessExpression(n)||ts.isParenthesizedExpression(n)||ts.isNonNullExpression(n))&&(!expr||n.end>expr.end))expr=n;ts.forEachChild(n,visit);}visit(sf);
   if(!expr)throw Error('no source expression '+s.where);
   const text=before.slice(expr.getStart(sf),expr.end);changes.push({start:expr.getStart(sf),end:expr.end,text:text+' satisfies SourceFileLike'});edits.push({id:s.id,file,line:s.line,column:s.col,before:text,after:text+' satisfies SourceFileLike',typeOnly:true,sourceTypePreservedBySatisfies:true});
  }
  if(!/\btype SourceFileLike\b/.test(before)){
   const imp=sf.statements.find(n=>ts.isImportDeclaration(n)&&n.importClause?.namedBindings&&ts.isNamedImports(n.importClause.namedBindings)&&n.moduleSpecifier.text==='./_namespaces/ts.js');if(!imp)throw Error('no namespace import '+file);
   const list=imp.importClause.namedBindings;changes.push({start:list.getStart(sf)+1,end:list.getStart(sf)+1,text:(before.includes('\r\n')?'\r\n':'\n')+'    type SourceFileLike,'});
  }
  changes.sort((a,b)=>b.start-a.start);let after=before;for(const e of changes)after=after.slice(0,e.start)+e.text+after.slice(e.end);
  const parsed=ts.createSourceFile(file,after,ts.ScriptTarget.Latest,true);if(parsed.parseDiagnostics.length)throw Error('parse diagnostics '+file);fs.writeFileSync(p,after);
 }
}else throw Error(kind);
fs.writeFileSync(path.join(out,'candidate-edits.json'),JSON.stringify(edits,null,2)+'\n');console.log(JSON.stringify({kind,edits:edits.length,tree:out}));
