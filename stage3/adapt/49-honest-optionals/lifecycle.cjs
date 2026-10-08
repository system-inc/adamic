// The scanner initializes text before returning its methods; printer entrypoints
// install a writer before emitting and preserve the optional saved writer.
const ts=require('typescript'),assert=require('node:assert/strict');
const {owner}=require('./plan.cjs');
module.exports=function lifecycle(text,file){
 const name=file.endsWith('/scanner.ts')?'text':'writer',scope=name==='text'?'createScanner':'createPrinter';
 const source=ts.createSourceFile('/49/input.ts',text,ts.ScriptTarget.Latest,true);
 const host={getSourceFile:n=>n===source.fileName?source:undefined,getDefaultLibFileName:()=>'',writeFile(){},getCurrentDirectory:()=>'/49',getDirectories:()=>[],fileExists:n=>n===source.fileName,readFile:n=>n===source.fileName?text:undefined,getCanonicalFileName:n=>n,useCaseSensitiveFileNames:()=>true,getNewLine:()=> '\n'};
 const program=ts.createProgram([source.fileName],{noLib:true,noResolve:true},host),checker=program.getTypeChecker();
 let declaration;function find(n){if(ts.isVariableDeclaration(n)&&ts.isIdentifier(n.name)&&n.name.text===name&&owner(n)===scope&&ts.isFunctionDeclaration(n.parent.parent.parent.parent)&&n.parent.parent.parent.parent.name?.text===scope){assert.ok(!declaration);declaration=n;}ts.forEachChild(n,find);}find(source);assert.ok(declaration);
 const symbol=checker.getSymbolAtLocation(declaration.name),edits=[];
 if(!declaration.type)edits.push({start:declaration.name.end,end:declaration.name.end,text:': string | undefined'});
 else if(!ts.isUnionTypeNode(declaration.type)||!declaration.type.types.some(t=>t.kind===ts.SyntaxKind.UndefinedKeyword))edits.push({start:declaration.type.end,end:declaration.type.end,text:' | undefined'});
 function visit(n){if(ts.isIdentifier(n)&&checker.getSymbolAtLocation(n)===symbol){const p=n.parent;
  const optionalInput=name==='text'&&ts.isCallExpression(p)&&ts.isIdentifier(p.expression)&&p.expression.text==='setText';
  const saved=name==='writer'&&ts.isVariableDeclaration(p)&&p.initializer===n;
  const guard=name==='writer'&&(ts.isPrefixUnaryExpression(p)&&p.operator===ts.SyntaxKind.ExclamationToken||ts.isTypeOfExpression(p)||ts.isBinaryExpression(p)&&p.left===n&&(p.operatorToken.kind===ts.SyntaxKind.AmpersandAmpersandToken||p.operatorToken.kind===ts.SyntaxKind.BarBarToken)||ts.isIfStatement(p)&&p.expression===n);
  const write=ts.isVariableDeclaration(p)&&p.name===n||ts.isBinaryExpression(p)&&p.left===n&&p.operatorToken.kind===ts.SyntaxKind.EqualsToken;
  if(!write&&!saved&&!guard&&!optionalInput&&!(ts.isNonNullExpression(p)&&p.expression===n))edits.push({start:n.end,end:n.end,text:'!'});
 }ts.forEachChild(n,visit);}visit(source);
 for(const e of edits.sort((a,b)=>b.start-a.start))text=text.slice(0,e.start)+e.text+text.slice(e.end);return text;
};
