const fs=require('fs'),path=require('path');const ts=require(process.env.TYPESCRIPT_API ? path.resolve(process.env.TYPESCRIPT_API) : path.resolve(__dirname, '../../api/node_modules/typescript'));
const tree=process.argv[2],out=process.argv[3],roots=ts.sys.readDirectory(tree+'/src',['.ts','.tsx'],undefined,['**/*']);
const program=ts.createProgram(roots,{strict:true,target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.NodeNext,moduleResolution:ts.ModuleResolutionKind.NodeNext,types:[],skipLibCheck:true}),checker=program.getTypeChecker(),rows=[];
function walk(root,f){const stack=[root];while(stack.length){const n=stack.pop();f(n);ts.forEachChild(n,c=>{stack.push(c)});}}
for(const sf of program.getSourceFiles()){
 if(!sf.fileName.startsWith(tree+'/src/'))continue;
 const locals=new Map();
 walk(sf,n=>{
  if(!ts.isCallExpression(n))return;
  const expr=n.expression,spelling=ts.isIdentifier(expr)?expr.text:ts.isPropertyAccessExpression(expr)?expr.name.text:'';
  if(!['sameMap','arrayFrom','arrayToMap','reduceLeft'].includes(spelling))return;
  const sig=checker.getResolvedSignature(n),d=sig?.declaration;
  if(!d||!ts.isFunctionDeclaration(d)||d.getSourceFile().fileName!==tree+'/src/compiler/core.ts')return;
  const row={_end:n.end,set:d.name.text,file:path.relative(tree,sf.fileName),line:sf.getLineAndCharacterOfPosition(n.getStart()).line+1,call:n.getText(sf),signature:checker.signatureToString(sig,undefined,ts.TypeFormatFlags.NoTruncation),uses:[]};
  if(n.arguments[1]){const cb=checker.getSignaturesOfType(checker.getTypeAtLocation(n.arguments[1]),ts.SignatureKind.Call)[0];if(cb){row.callbackParameters=cb.parameters.map(p=>checker.typeToString(checker.getTypeOfSymbolAtLocation(p,n.arguments[1]),undefined,ts.TypeFormatFlags.NoTruncation));row.callbackResult=checker.typeToString(checker.getReturnTypeOfSignature(cb),undefined,ts.TypeFormatFlags.NoTruncation);}}
  rows.push(row);
  let p=n.parent;let child=n;
  while(ts.isParenthesizedExpression(p)||ts.isNonNullExpression(p)||ts.isAsExpression(p)||ts.isConditionalExpression(p)&&(p.whenTrue===child||p.whenFalse===child)){child=p;p=p.parent;}
  if(ts.isBinaryExpression(p)&&p.operatorToken.kind===ts.SyntaxKind.EqualsToken&&ts.isIdentifier(p.left)){row.assignment=p.left.text;locals.set(checker.getSymbolAtLocation(p.left),row);}
  if(ts.isVariableDeclaration(p)&&ts.isIdentifier(p.name)){row.binding=p.name.text;locals.set(checker.getSymbolAtLocation(p.name),row);}
 });
 if(locals.size)walk(sf,n=>{
  if(!ts.isIdentifier(n))return;
  const row=locals.get(checker.getSymbolAtLocation(n));if(!row||n.getStart()<row._end||ts.isVariableDeclaration(n.parent)&&n.parent.name===n)return;
  let p=n.parent;
  while(ts.isParenthesizedExpression(p)||ts.isNonNullExpression(p)||ts.isAsExpression(p))p=p.parent;
  let context=p;while(context.parent&&!ts.isStatement(context)&&!ts.isArrowFunction(context))context=context.parent;
  row.uses.push({line:sf.getLineAndCharacterOfPosition(n.getStart()).line+1,kind:ts.SyntaxKind[p.kind],identity:ts.isBinaryExpression(p)&&[ts.SyntaxKind.EqualsEqualsEqualsToken,ts.SyntaxKind.ExclamationEqualsEqualsToken].includes(p.operatorToken.kind),text:p.getText(sf),context:context.getText(sf)});
 });
}
fs.writeFileSync(out,JSON.stringify(rows,null,2)+'\n');console.log(JSON.stringify({calls:rows.length,identityReaders:rows.filter(r=>r.set==='sameMap'&&r.uses.some(u=>u.identity)).map(r=>({file:r.file,line:r.line,uses:r.uses.filter(u=>u.identity)}))},null,2));
