/* Source inspection only, using the installed pinned TypeScript 6.0.3 checker. */
const fs=require('fs'),path=require('path');
const ts=require('/workspace/graph-boundary/stage3/api/node_modules/typescript');
const input=process.argv[2],output=process.argv[3];
const sites=JSON.parse(fs.readFileSync(input,'utf8'));
const files=[...new Set(sites.map(s=>s.replace(/:\d+:\d+$/,'')))];
const program=ts.createProgram(files,{target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.NodeNext,moduleResolution:ts.ModuleResolutionKind.NodeNext,strict:true,skipLibCheck:true,noEmit:true,types:[],allowImportingTsExtensions:true});
const checker=program.getTypeChecker();
function isString(t){return !(t.flags&(ts.TypeFlags.Any|ts.TypeFlags.Unknown))&&checker.isTypeAssignableTo(t,checker.getStringType());}
function indexInfo(t,stringOnly=false){
 if(!t)return null;
 if(t.flags & (ts.TypeFlags.StringLike|ts.TypeFlags.NumberLike|ts.TypeFlags.BooleanLike|ts.TypeFlags.Any|ts.TypeFlags.Unknown))return null;
 if(checker.isArrayType(t)||checker.isTupleType(t))return null;
 const infos=checker.getIndexInfosOfType(t).filter(i=>!stringOnly || isString(i.keyType));
 if(!infos.length)return null;
 return {type:checker.typeToString(t,undefined,ts.TypeFormatFlags.NoTruncation),index_signatures:infos.map(i=>({key:checker.typeToString(i.keyType),value:checker.typeToString(i.type),declaration:i.declaration?i.declaration.getText():null}))};
}
const result={typescript_version:ts.version,method:'TypeScript source AST and checker only; no Adamic lowering rerun',compiler_options:program.getCompilerOptions(),sites:{}};
for(const where of sites){
 const [,name,l,c]=where.match(/^(.*):(\d+):(\d+)$/)||[];
 const sf=program.getSourceFile(name);
 if(!sf){result.sites[where]={error:'source file not in inspection program'};continue;}
 const pos=ts.getPositionOfLineAndCharacter(sf,+l-1,+c-1);
 let token=ts.getTokenAtPosition(sf,pos),node=token;
 const candidates=[];
 while(node&&node!==sf){
  // A finding's AST node begins at this token. Do not classify a failure on
  // a child merely because it happens inside an unrelated record literal.
  if(node.getStart(sf)===token.getStart(sf))candidates.push(node);
  node=node.parent;
 }
 const matches=[];
 function literalOf(n){
  if(!n)return null;
  while(ts.isAsExpression(n)||ts.isTypeAssertionExpression(n)||ts.isSatisfiesExpression(n)||ts.isParenthesizedExpression(n))n=n.expression;
  return ts.isObjectLiteralExpression(n)?n:null;
 }
 const literals=new Set();
 for(const n of candidates){
  if(ts.isObjectLiteralExpression(n))literals.add(n);
  if(ts.isVariableDeclaration(n)&&literalOf(n.initializer))literals.add(literalOf(n.initializer));
  if(ts.isVariableStatement(n))for(const d of n.declarationList.declarations)if(literalOf(d.initializer))literals.add(literalOf(d.initializer));
  if(ts.isReturnStatement(n)&&literalOf(n.expression))literals.add(literalOf(n.expression));
  if(ts.isBinaryExpression(n)&&n.operatorToken.kind===ts.SyntaxKind.EqualsToken&&literalOf(n.right))literals.add(literalOf(n.right));
  if((ts.isSpreadAssignment(n)||ts.isPropertyAssignment(n))&&ts.isObjectLiteralExpression(n.parent))literals.add(n.parent);
 }
 // A finding on a literal member or its initializer still belongs to that
 // literal. Stop at a function boundary so callback bodies are not classified
 // by an enclosing object used only to hold the callback.
 for(let ancestor=token;ancestor&&ancestor!==sf;ancestor=ancestor.parent){
  if(ts.isFunctionLike(ancestor))break;
  if(ts.isObjectLiteralExpression(ancestor)){literals.add(ancestor);break;}
 }
 for(const n of literals){
  {
   const own=indexInfo(checker.getTypeAtLocation(n));
   const contextual=indexInfo(checker.getContextualType(n));
   let parent=n.parent,asserted=null;
   if(ts.isAsExpression(parent)||ts.isTypeAssertionExpression(parent)||ts.isSatisfiesExpression(parent))asserted=indexInfo(checker.getTypeAtLocation(parent.type));
   if(own||contextual||asserted)matches.push({basis:'object literal with an index-signature type',node_kind:ts.SyntaxKind[n.kind],source:n.getText(sf),own,contextual,asserted});
  }
 }
 for(const n of candidates){
  if(ts.isElementAccessExpression(n)||ts.isPropertyAccessExpression(n)){
   const receiver=indexInfo(checker.getTypeAtLocation(n.expression),true);
   let key=ts.isElementAccessExpression(n)?n.argumentExpression:null;
   let keyType=key?checker.typeToString(checker.getTypeAtLocation(key)):'string (named property)';
   const keyIsString=!key||isString(checker.getTypeAtLocation(key));
   if(receiver&&keyIsString)matches.push({basis:'string-keyed read/write on an index-signature object',node_kind:ts.SyntaxKind[n.kind],source:n.getText(sf),receiver,key_type:keyType});
  }
 }
 result.sites[where]={line:+l,column:+c,token:token.getText(sf),node_kinds:candidates.map(n=>ts.SyntaxKind[n.kind]),matches};
}
fs.writeFileSync(output,JSON.stringify(result,null,2)+'\n');
console.log(JSON.stringify({typescript_version:ts.version,sites:sites.length,matched:Object.values(result.sites).filter(s=>s.matches?.length).length}));
