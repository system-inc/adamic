'use strict';
const fs=require('fs'),path=require('path'),ts=require('typescript'),assert=require('assert/strict');
const s=__dirname,tree=s+'/build-change';
const {compiler,nodes}=require(s+'/work/stage3/adapt/47-host-errors/adapt.cjs');
const program=compiler(tree,ts),checker=program.getTypeChecker();
const resolver=program.getSourceFile(tree+'/src/compiler/moduleNameResolver.ts');
const seed=resolver.statements.find(n=>ts.isFunctionDeclaration(n)&&n.name?.text==='resolveJSModule');assert(seed);
function functionForDeclaration(d){
 if(d.body&&ts.isFunctionLike(d))return d;
 if(ts.isVariableDeclaration(d)&&d.initializer){const n=d.initializer;if(ts.isArrowFunction(n)||ts.isFunctionExpression(n))return n;}
 return undefined;
}
function label(n){const sf=n.getSourceFile(),pos=sf.getLineAndCharacterOfPosition(n.getStart(sf));return path.relative(tree,sf.fileName)+':'+(pos.line+1)+':'+(n.name?.getText(sf)||n.parent?.name?.getText(sf)||'<anonymous>');}
const queue=[seed],seen=new Set(),edges=[],accesses=new Map();
const members=new Set(['fileExists','readFile','directoryExists','realpath','getCurrentDirectory','getDirectories','trace','getGlobalTypingsCacheLocation','useCaseSensitiveFileNames']);
while(queue.length){const fn=queue.shift();if(seen.has(fn))continue;seen.add(fn);const targets=new Set();
 function visit(n){
  if(n!==fn&&ts.isFunctionLike(n)&&n.body)targets.add(n);
  if(ts.isIdentifier(n)){
   let symbol=checker.getSymbolAtLocation(n);
   if(symbol?.flags&ts.SymbolFlags.Alias)symbol=checker.getAliasedSymbol(symbol);
   for(const d of symbol?.declarations||[]){const f=functionForDeclaration(d);if(f&&!f.getSourceFile().isDeclarationFile)targets.add(f);}
  }
  if(ts.isPropertyAccessExpression(n)&&members.has(n.name.text)){
   const sf=n.getSourceFile(),pos=sf.getLineAndCharacterOfPosition(n.getStart(sf));
   const key=sf.fileName+':'+n.pos;
   let p=n.parent;while(ts.isNonNullExpression(p)||ts.isParenthesizedExpression(p))p=p.parent;
   const invoked=ts.isCallExpression(p)&&p.expression.getStart(sf)===n.getStart(sf);
   const optional=invoked&&!!p.questionDotToken;
   const write=(ts.isBinaryExpression(p)&&p.left===n&&p.operatorToken.kind>=ts.SyntaxKind.FirstAssignment&&p.operatorToken.kind<=ts.SyntaxKind.LastAssignment)||ts.isDeleteExpression(p)||(ts.isPrefixUnaryExpression(p)||ts.isPostfixUnaryExpression(p))&&[ts.SyntaxKind.PlusPlusToken,ts.SyntaxKind.MinusMinusToken].includes(p.operator);
   accesses.set(key,{file:path.relative(tree,sf.fileName),line:pos.line+1,member:n.name.text,receiver:n.expression.getText(sf),text:n.getText(sf),parent:p.getText(sf),invoked,optional,write,owner:label(fn)});
  }
  ts.forEachChild(n,visit);
 }
 visit(fn.body);for(const target of targets){edges.push([label(fn),label(target)]);queue.push(target);}
}
const sys=program.getSourceFile(tree+'/src/compiler/sys.ts');
const decl=nodes(ts,sys,n=>ts.isVariableDeclaration(n)&&n.name.getText(sys)==='nodeSystem');assert.equal(decl.length,1);
const sym=checker.getSymbolAtLocation(decl[0].name);const refs=nodes(ts,sys,n=>ts.isIdentifier(n)&&n!==decl[0].name&&checker.getSymbolAtLocation(n)===sym);
const sysWrites=refs.filter(n=>{let p=n.parent; if(ts.isPropertyAccessExpression(p)||ts.isElementAccessExpression(p))p=p.parent;return ts.isBinaryExpression(p)&&p.left.getStart(sys)===n.getStart(sys)&&p.operatorToken.kind>=ts.SyntaxKind.FirstAssignment&&p.operatorToken.kind<=ts.SyntaxKind.LastAssignment||ts.isDeleteExpression(p);});assert.equal(sysWrites.length,0);
const result={upstream:'050880ce59e30b356b686bd3144efe24f875ebc8',graphKind:'conservative symbol-bound call/reference closure; includes dead conditional bodies and nested closures',functions:[...seen].map(label).sort(),edges:edges.sort(),accesses:[...accesses.values()].sort((a,b)=>a.file.localeCompare(b.file)||a.line-b.line),nodeSystemReferences:refs.map(n=>({line:sys.getLineAndCharacterOfPosition(n.getStart(sys)).line+1,text:n.parent.getText(sys)})),nodeSystemWritesAfterConstruction:sysWrites.length};
fs.writeFileSync(s+'/evidence/resolver-audit.json',JSON.stringify(result,null,2)+'\n');console.log(JSON.stringify({functions:seen.size,edges:edges.length,memberAccesses:accesses.size,nodeSystemWritesAfterConstruction:sysWrites.length}));
