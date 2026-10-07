const ts=require('typescript'), fs=require('fs'), path=require('path');
const root=process.argv[2];
const files=fs.readdirSync(root).filter(f=>f.endsWith('.ts')).sort().map(f=>path.join(root,f));
const program=ts.createProgram(files,{strict:true,noUncheckedIndexedAccess:false,target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.ESNext,skipLibCheck:true});
const checker=program.getTypeChecker();
let lines=0,physicalLines=0,clean=0; const counts={}, sites={};
const nullable=t=>!!(t.flags&(ts.TypeFlags.Null|ts.TypeFlags.Undefined))||!!t.types?.some(nullable);
const add=(key,n,sf)=>{counts[key]=(counts[key]||0)+1;(sites[key]??=[]).push(path.basename(sf.fileName)+':'+(sf.getLineAndCharacterOfPosition(n.getStart(sf)).line+1));};
for(const file of files){
 const sf=program.getSourceFile(file);
 lines+=sf.text.split(/\r?\n/).filter(l=>l.trim()).length;
 physicalLines+=sf.text.split(/\r?\n/).length-(sf.text.endsWith('\n')?1:0);
 function visit(n){
  if(n.kind===ts.SyntaxKind.AnyKeyword)add('any',n,sf);
  if(ts.isAsExpression(n)||ts.isTypeAssertionExpression(n)){
   const isConst=n.type.getText(sf)==='const';
   add(isConst?'as-const':'casts',n,sf);
   if(!isConst)add(ts.isAsExpression(n)?'as-casts':'angle-casts',n,sf);
   if(!isConst){
    const from=checker.getTypeAtLocation(n.expression), to=checker.getTypeFromTypeNode(n.type);
    let key;
    if(from.flags&(ts.TypeFlags.Any|ts.TypeFlags.Unknown)||to.flags&ts.TypeFlags.Any)key='cast-unchecked';
    else if(checker.isTypeAssignableTo(from,to))key='cast-up';
    else if(checker.isTypeAssignableTo(to,from))key='cast-down';
    else key='cast-unrelated';
    add(key,n,sf);
    if(ts.isAsExpression(n.expression)&&n.expression.type.kind===ts.SyntaxKind.UnknownKeyword)add('as-unknown-as',n,sf);
   }
  }
  if(ts.isNonNullExpression(n))add('non-null',n,sf);
  if(n.exclamationToken&&(ts.isPropertyDeclaration(n)||ts.isVariableDeclaration(n)))add(ts.isPropertyDeclaration(n)?'definite-field':'definite-local',n,sf);
  if(ts.isTypePredicateNode(n))add(n.assertsModifier?'assertion':'guard',n,sf);
  if(ts.isMethodSignature(n))add('method-signature',n,sf);
  if(ts.isTypeReferenceNode(n)&&n.typeName.getText(sf)==='Function')add('Function',n,sf);
  if(ts.isSatisfiesExpression(n))add('satisfies',n,sf);
  if((ts.isPropertyAccessExpression(n)||ts.isElementAccessExpression(n)||ts.isCallExpression(n))&&n.questionDotToken){
   add('optional-chain',n,sf);
   const t=checker.getTypeAtLocation(n.expression);
   if(!(t.flags&(ts.TypeFlags.Any|ts.TypeFlags.Unknown))&&!nullable(t))add('optional-nonoptional',n,sf);
  }
  if(ts.isIndexSignatureDeclaration(n))add('index-declaration',n,sf);
  if(ts.isElementAccessExpression(n)&&!(ts.isBinaryExpression(n.parent)&&n.parent.left===n&&n.parent.operatorToken.kind===ts.SyntaxKind.EqualsToken)){
   const t=checker.getTypeAtLocation(n.expression),k=checker.getTypeAtLocation(n.argumentExpression);
   if(!checker.isArrayType(t)&&!checker.isTupleType(t)){
    const idx=checker.getIndexTypeOfType(t,k.flags&ts.TypeFlags.NumberLike?ts.IndexKind.Number:ts.IndexKind.String);
    const literal=k.flags&(ts.TypeFlags.StringLiteral|ts.TypeFlags.NumberLiteral);
    if(idx&&!nullable(idx)&&!(literal&&checker.getPropertyOfType(t,String(k.value))))add('index-read-nonoptional',n,sf);
   }
  }
  if(ts.isCallExpression(n)&&ts.isPropertyAccessExpression(n.expression)&&n.expression.expression.getText(sf)==='Object'){
   const name=n.expression.name.text;
   if(name==='defineProperty')add('defineProperty',n,sf);
   if(name==='assign'&&n.arguments.length&&!ts.isObjectLiteralExpression(n.arguments[0]))add('assign-existing-candidate',n,sf);
  }
  if(ts.isBinaryExpression(n)&&n.operatorToken.kind===ts.SyntaxKind.EqualsToken&&ts.isPropertyAccessExpression(n.left)){
   const target=checker.getTypeAtLocation(n.left.expression);
   const sym=checker.getPropertyOfType(target,n.left.name.text);
   if(!sym||sym.declarations?.every(d=>ts.isPropertyAccessExpression(d)||ts.isBinaryExpression(d)))add('expando-candidate',n,sf);
  }
  ts.forEachChild(n,visit);
 }
 visit(sf);
 // Directive tokens must be in comments, not strings.
 const scan=ts.createScanner(ts.ScriptTarget.Latest,false,ts.LanguageVariant.Standard,sf.text);
 for(let token=scan.scan();token!==ts.SyntaxKind.EndOfFileToken;token=scan.scan()){
  if(token===ts.SyntaxKind.SingleLineCommentTrivia||token===ts.SyntaxKind.MultiLineCommentTrivia){
   for(const m of scan.getTokenText().matchAll(/@ts-(ignore|expect-error)\b/g))add('ts-'+m[1],sf,sf);
  }
 }
 // Use the documented seven-family clean-file definition, not every new survey metric.
 const dirty=['any','casts','as-unknown-as','non-null','guard','assertion','method-signature'].some(k=>(sites[k]||[]).some(s=>s.startsWith(path.basename(file)+':')));
 if(!dirty)clean++;
}
for(const k of ['definite-field','ts-ignore','ts-expect-error','Function','satisfies'])counts[k]??=0;
console.log(JSON.stringify({physicalLines,physicalRates:Object.fromEntries(Object.entries(counts).map(([k,v])=>[k,+(1000*v/physicalLines).toFixed(3)])),typescript:ts.version,files:files.length,nonblankLines:lines,cleanFiles:clean,cleanPercent:100*clean/files.length,counts,rates:Object.fromEntries(Object.entries(counts).map(([k,v])=>[k,+(1000*v/lines).toFixed(3)])),sites},null,2));
