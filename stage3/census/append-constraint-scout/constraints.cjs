// Annotate array-view findings using stock checker declarations; preserve all rows for audit.
const fs=require('fs'),path=require('path'),ts=require('typescript');
const root=process.argv[2],rows=JSON.parse(fs.readFileSync(process.argv[3],'utf8'));
const files=[];function walk(p){for(const d of fs.readdirSync(p,{withFileTypes:true})){const q=path.join(p,d.name);if(d.isDirectory())walk(q);else if(q.endsWith('.ts'))files.push(q)}}walk(root);
const program=ts.createProgram(files,{strict:true,noUncheckedIndexedAccess:true,exactOptionalPropertyTypes:true,module:ts.ModuleKind.ESNext,moduleResolution:ts.ModuleResolutionKind.Bundler,target:ts.ScriptTarget.ES2024,types:[],noEmit:true});const checker=program.getTypeChecker();
function arrayElements(text){
 const sf=ts.createSourceFile('target.ts','type Target = '+text+';',ts.ScriptTarget.Latest,true);if(sf.parseDiagnostics.length)return [];
 const out=[];function unparen(n){while(ts.isParenthesizedTypeNode(n))n=n.type;return n}
 function visit(n){n=unparen(n);if(ts.isUnionTypeNode(n)){n.types.forEach(visit);return}if(ts.isTypeOperatorNode(n)&&n.operator===ts.SyntaxKind.ReadonlyKeyword){visit(n.type);return}if(ts.isArrayTypeNode(n)){out.push(unparen(n.elementType).getText(sf));return}if(ts.isTypeReferenceNode(n)&&/^(?:Array|ReadonlyArray|NodeArray|SortedReadonlyArray|SortedArray)$/.test(n.typeName.getText(sf))&&n.typeArguments?.length===1)out.push(unparen(n.typeArguments[0]).getText(sf))}
 const alias=sf.statements[0];if(alias&&ts.isTypeAliasDeclaration(alias))visit(alias.type);return out;
}
for(const row of rows){
 row.target_array_elements=arrayElements(row.target);
 const sf=program.getSourceFile(path.join(root,row.file)),pos=sf.getPositionOfLineAndCharacter(row.line-1,row.column-1);let node=sf;
 function descend(n){if(n.getStart(sf)<=pos&&pos<n.end){node=n;ts.forEachChild(n,descend)}}descend(sf);
 while(node.parent&&node.parent!==sf&&node.parent.getStart(sf)===pos)node=node.parent;
 let call=node.parent;while(call&&!ts.isCallExpression(call)&&!ts.isNewExpression(call))call=call.parent;
 if(ts.isCallExpression(node)||ts.isNewExpression(node)){
  const source=/^a value of type (.*?) seen as /.exec(row.reason)?.[1];
  const own=checker.typeToString(checker.getTypeAtLocation(node));
  if(source&&source.replace(/\s/g,'')!==own.replace(/\s/g,'')){call=node;row.judgment_site='inferred type argument';row.expression_type=own;}
 }
 if(!call&&(ts.isCallExpression(node)||ts.isNewExpression(node)))call=node;
 if(!call)continue;
 row.callee=call.expression.getText(sf);
 let arg=node;if(call===node&&row.judgment_site){row.argument=(call.arguments||[]).map(a=>a.getText(sf)).join(', ');}else{while(arg.parent&&arg.parent!==call)arg=arg.parent;row.argument=arg.getText(sf);}
 const s=checker.getResolvedSignature(call);if(!s)continue;
 const target=s.target||s;
 row.constraints=(target.typeParameters||[]).map(t=>({parameter:checker.typeToString(t),constraint:checker.getBaseConstraintOfType(t)?checker.typeToString(checker.getBaseConstraintOfType(t)):null}));
 if(s.declaration){const d=s.declaration,f=d.getSourceFile(),lc=f.getLineAndCharacterOfPosition(d.getStart(f));row.declaration=path.relative(root,f.fileName)+':'+(lc.line+1)}
}
console.log(JSON.stringify(rows,null,2));
