const fs=require('fs'),path=require('path');
const ts=require('/workspace/or3-parent/stage3/api/node_modules/typescript');
const req=JSON.parse(fs.readFileSync(process.argv[2],'utf8'));
const config=ts.findConfigFile(path.dirname(req.entry),ts.sys.fileExists,'tsconfig.json');
let options={target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.NodeNext,moduleResolution:ts.ModuleResolutionKind.NodeNext,noEmit:true,skipLibCheck:true};
if(config){const read=ts.readConfigFile(config,ts.sys.readFile);if(read.error)throw Error(ts.flattenDiagnosticMessageText(read.error.messageText,' '));options={...ts.parseJsonConfigFileContent(read.config,ts.sys,path.dirname(config)).options,noEmit:true};}
const program=ts.createProgram({rootNames:req.sources,options});const checker=program.getTypeChecker();
function declarationEvidence(n){
 let symbol=ts.isShorthandPropertyAssignment(n)?checker.getShorthandAssignmentValueSymbol(n):checker.getSymbolAtLocation(n);
 if(!symbol&&ts.isPropertyAccessExpression(n))symbol=checker.getSymbolAtLocation(n.name);
 if(!symbol){const t=checker.getTypeAtLocation(n);symbol=t&&t.getSymbol();}
 const seen=new Set();while(symbol&&(symbol.flags&ts.SymbolFlags.Alias)&&!seen.has(symbol)){seen.add(symbol);symbol=checker.getAliasedSymbol(symbol);}
 const ds=symbol?.declarations||[];
 const f=ds.filter(ts.isFunctionDeclaration);if(!f.some(d=>!d.body)||!f.some(d=>d.body))return null;
 return {symbol:symbol.getName(),declarations:f.map(d=>{const s=d.getSourceFile(),p=s.getLineAndCharacterOfPosition(d.getStart(s));return {where:`${s.fileName}:${p.line+1}:${p.character+1}`,body:!!d.body,text:d.getText(s).slice(0,500)};})};
}
const results=[];
for(const row of req.rows){
 const m=/^(.*):(\d+):(\d+)$/.exec(row.finding.where);if(!m)throw Error('bad location');
 const sf=program.getSourceFile(m[1]);if(!sf)throw Error('missing source '+m[1]);
 const pos=sf.getPositionOfLineAndCharacter(+m[2]-1,+m[3]-1);const candidates=[];
 function visit(n){if(n.getStart(sf)===pos&&(ts.isIdentifier(n)||ts.isPropertyAccessExpression(n)||ts.isElementAccessExpression(n)||ts.isShorthandPropertyAssignment(n))){const e=declarationEvidence(n);if(e)candidates.push({n,e});}ts.forEachChild(n,c=>{if(c.pos<=pos&&c.end>pos)visit(c);});}
 visit(sf);candidates.sort((a,b)=>(a.n.end-a.n.getStart(sf))-(b.n.end-b.n.getStart(sf)));
 if(!candidates.length)throw Error('cannot resolve overload set at '+row.finding.where);
 const {n,e}=candidates[0];let effective=n;const wrappers=[];
 while(effective.parent&&ts.isParenthesizedExpression(effective.parent)&&effective.parent.expression===effective){wrappers.push(ts.SyntaxKind[effective.parent.kind]);effective=effective.parent;}
 const p=effective.parent;const callee=!!p&&((ts.isCallExpression(p)||ts.isNewExpression(p))&&p.expression===effective||ts.isTaggedTemplateExpression(p)&&p.tag===effective);
 let role=p?ts.SyntaxKind[p.kind]:'none';
 if(p&&ts.isCallExpression(p))role+=p.expression===effective?'.expression':'.arguments['+p.arguments.indexOf(effective)+']';
 else if(p&&ts.isVariableDeclaration(p)&&p.initializer===effective)role+='.initializer';
 else if(p&&ts.isPropertyAssignment(p)&&p.initializer===effective)role+='.initializer';
 else if(p&&ts.isReturnStatement(p))role+='.expression';
 let context=effective;for(let i=0;i<4&&context.parent;i++){context=context.parent;if(ts.isStatement(context)||ts.isVariableDeclaration(context)||ts.isPropertyAssignment(context))break;}
 const contextStart=context.getStart(sf),contextText=context.getText(sf),referenceOffset=n.getStart(sf)-contextStart;
 const excerptStart=Math.max(0,referenceOffset-700);const excerpt=contextText.slice(excerptStart,Math.max(excerptStart+1800,referenceOffset+n.getWidth(sf)));
 const start=sf.getLineAndCharacterOfPosition(n.getStart(sf)),end=sf.getLineAndCharacterOfPosition(n.end);
 results.push({...row,classification:callee?'callee of a call; fix wrong':'used as a value; fail-closed and correct',is_callee:callee,expression:n.getText(sf),node_kind:ts.SyntaxKind[n.kind],effective_kind:ts.SyntaxKind[effective.kind],parent_role:role,parentheses:wrappers,range:{start_utf16:n.getStart(sf),end_utf16:n.end,start_line:start.line+1,start_column:start.character+1,end_line:end.line+1,end_column:end.character+1},context:excerpt,context_kind:ts.SyntaxKind[context.kind],context_start_utf16:contextStart+excerptStart,context_truncated:excerptStart>0||excerpt.length<contextText.length,overloaded_set:e});
}
fs.writeFileSync(process.argv[3],JSON.stringify({config,options,results},null,2)+'\n');
console.log(JSON.stringify({units:results.length,callee:results.filter(r=>r.is_callee).length,value:results.filter(r=>!r.is_callee).length}));
