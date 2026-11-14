// Stock TypeScript control and AST attribution, never used for native lowering.
const fs=require('fs'),path=require('path');
const ts=require('typescript');
const root=process.argv[2];
const files=[];function walk(p){for(const d of fs.readdirSync(p,{withFileTypes:true})){const q=path.join(p,d.name);if(d.isDirectory())walk(q);else if(q.endsWith('.ts'))files.push(q)}}walk(root);
const program=ts.createProgram(files,{strict:true,noUncheckedIndexedAccess:true,exactOptionalPropertyTypes:true,module:ts.ModuleKind.ESNext,moduleResolution:ts.ModuleResolutionKind.Bundler,target:ts.ScriptTarget.ES2024,types:[],noEmit:true});
const checker=program.getTypeChecker();
const diagnostics=ts.getPreEmitDiagnostics(program);
const sf=program.getSourceFile(path.join(root,'scanner.ts'));
function visit(n){if(ts.isCallExpression(n)&&n.expression.getText(sf)==='append'&&n.arguments[0]?.getText(sf)==='decimalEscapes'){
 const s=checker.getResolvedSignature(n);console.log(JSON.stringify({stockVersion:ts.version,diagnostics:diagnostics.length,where:sf.getLineAndCharacterOfPosition(n.getStart(sf)),signature:checker.signatureToString(s),typeArguments:checker.getTypeArgumentsForResolvedSignature(s).map(t=>checker.typeToString(t)),parameters:s.parameters.map(p=>checker.typeToString(checker.getTypeOfSymbolAtLocation(p,n))),arguments:n.arguments.map(a=>({text:a.getText(sf),own:checker.typeToString(checker.getTypeAtLocation(a)),contextual:checker.typeToString(checker.getContextualType(a))}))},null,2));}ts.forEachChild(n,visit)}visit(sf);
