// Independent TypeScript 6.0.3 AST/checker audit; no lowering or compiler acceptance claim.
const fs=require('fs');
const ts=require('/workspace/loose-slot-scratch/eleven/head/stage3/api/node_modules/typescript');
const root='/workspace/loose-slot-scratch/eleven/adapted-head/src/compiler';
const roots=ts.sys.readDirectory(root,['.ts'],[],['**/*.ts']).sort();
const program=ts.createProgram(roots,{target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.NodeNext,moduleResolution:ts.ModuleResolutionKind.NodeNext,strict:true,noEmit:true,skipLibCheck:true,allowImportingTsExtensions:true});
const checker=program.getTypeChecker();
const where=n=>{const sf=n.getSourceFile(),p=sf.getLineAndCharacterOfPosition(n.getStart(sf));return `${sf.fileName}:${p.line+1}:${p.character+1}`;};
const owner=n=>{for(;n;n=n.parent)if(ts.isFunctionLike(n))return n;return undefined;};
const nodes=[];const visit=n=>{nodes.push(n);ts.forEachChild(n,visit);};
for(const file of roots)visit(program.getSourceFile(file));
const catches=[],symbols=new Map();
const symbol=n=>ts.isShorthandPropertyAssignment(n.parent)&&n.parent.name===n?checker.getShorthandAssignmentValueSymbol(n.parent):checker.getSymbolAtLocation(n);
for(const n of nodes){if(!ts.isCatchClause(n))continue;const d=n.variableDeclaration;const r={ClauseWhere:where(n),Bound:!!d,Name:d?.name.getText(),BindingWhere:d?where(d.name):null,Captures:[],References:[]};
 if(d&&ts.isIdentifier(d.name)){const sym=checker.getSymbolAtLocation(d.name);symbols.set(sym,r);const outer=owner(n.parent),seen=new Set();const scan=x=>{if(ts.isTypeNode(x))return;if(ts.isIdentifier(x)&&symbol(x)===sym){const own=owner(x.parent);r.References.push({Where:where(x),ParentWhere:where(x.parent),ParentKind:ts.SyntaxKind[x.parent.kind],ClosureWhere:own&&own!==outer?where(own):null});if(own&&own!==outer&&!seen.has(own)){seen.add(own);r.Captures.push({Where:where(own),Text:own.getText()});}}ts.forEachChild(x,scan);};scan(n.block);}
 catches.push(r);
}
const calls=[];
for(const n of nodes){if(!ts.isCallExpression(n))continue;n.arguments.forEach((a,i)=>{let x=a;while(ts.isParenthesizedExpression(x))x=x.expression;if(!ts.isIdentifier(x))return;const binding=symbols.get(checker.getSymbolAtLocation(x));if(!binding)return;const signature=checker.getResolvedSignature(n),parameters=signature?.getParameters();if(!parameters?.length)return;const p=parameters[Math.min(i,parameters.length-1)];let target=checker.getTypeOfSymbolAtLocation(p,a);const decl=p.declarations?.[0];if(decl?.dotDotDotToken)target=checker.getIndexTypeOfType(target,ts.IndexKind.Number);if(!target||!(target.flags&ts.TypeFlags.Unknown))return;calls.push({Where:where(n),ArgumentWhere:where(a),BindingWhere:binding.BindingWhere,Parameter:p.name,ParameterWhere:decl?where(decl):null,ArgumentType:checker.typeToString(checker.getTypeAtLocation(a)),ParameterType:checker.typeToString(target)});});}
fs.writeFileSync(process.argv[2],JSON.stringify({typescript:ts.version,Roots:roots.length,CatchClauses:catches.length,BoundCatchClauses:catches.filter(x=>x.Bound).length,CapturedCatchBindings:catches.filter(x=>x.Captures.length).length,UnknownCalls:calls,Catches:catches},null,2)+'\n');
