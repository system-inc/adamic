// Stock pinned checker evidence for every concatenate call in the five units.
const fs=require('fs'),path=require('path');
const ts=require('/workspace/op2-base/stage3/api/node_modules/typescript');
const req=JSON.parse(fs.readFileSync(process.argv[2],'utf8'));
const config=ts.findConfigFile(path.dirname(req.entry),ts.sys.fileExists,'tsconfig.json');
let options={target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.NodeNext,moduleResolution:ts.ModuleResolutionKind.NodeNext,noEmit:true,skipLibCheck:true};
if(config){const read=ts.readConfigFile(config,ts.sys.readFile);if(read.error)throw Error(ts.flattenDiagnosticMessageText(read.error.messageText,' '));options={...ts.parseJsonConfigFileContent(read.config,ts.sys,path.dirname(config)).options,noEmit:true};}
const program=ts.createProgram({rootNames:req.sources,options}),checker=program.getTypeChecker();
function where(n){const sf=n.getSourceFile(),p=sf.getLineAndCharacterOfPosition(n.getStart(sf));return `${sf.fileName}:${p.line+1}:${p.character+1}`;}
const results=[];
for(const row of req.rows){
 const m=/^(.*):(\d+):(\d+)$/.exec(row.unit.where),sf=program.getSourceFile(m[1]);if(!sf)throw Error('missing file');
 const pos=sf.getPositionOfLineAndCharacter(+m[2]-1,+m[3]-1);
 const decl=sf.statements.find(s=>s.getStart(sf)===pos);if(!decl)throw Error('missing unit '+row.unit.where);
 const calls=[];
 function visit(n){
  if(ts.isCallExpression(n)&&((ts.isIdentifier(n.expression)&&n.expression.text==='concatenate')||(ts.isPropertyAccessExpression(n.expression)&&n.expression.name.text==='concatenate'))){
   const sig=checker.getResolvedSignature(n),d=sig?.getDeclaration();if(!d)throw Error('missing resolved declaration');
   const symbol=checker.getSymbolAtLocation(d.name),overloads=symbol.declarations.filter(x=>ts.isFunctionDeclaration(x)&&!x.body),index=overloads.indexOf(d)+1;
   if(!index)throw Error('resolved signature not an overload');
   calls.push({where:where(n),text:n.getText(sf),resolved_declaration:{where:where(d),text:d.getText(d.getSourceFile()),index,overload_count:overloads.length},all_overloads:overloads.map((x,i)=>({index:i+1,where:where(x),text:x.getText(x.getSourceFile())}))});
  }
  ts.forEachChild(n,visit);
 }
 visit(decl);if(!calls.length)throw Error('no concatenate call in '+row.unit.where);
 results.push({...row,calls});
}
const diagnostics=ts.getPreEmitDiagnostics(program).map(d=>({code:d.code,message:ts.flattenDiagnosticMessageText(d.messageText,' ')}));
fs.writeFileSync(process.argv[3],JSON.stringify({config,options,diagnostics,results},null,2)+'\n');console.log(JSON.stringify({units:results.length,calls:results.reduce((n,r)=>n+r.calls.length,0),diagnostics:diagnostics.length}));
