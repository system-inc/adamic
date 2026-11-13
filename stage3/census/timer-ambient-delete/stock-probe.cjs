const ts=require('/workspace/timer-ambient-delete/stage3/api/node_modules/typescript');
const path=require('path');
const file=path.join(__dirname,'forwarders.ts');
const options={strict:true,noUncheckedIndexedAccess:true,exactOptionalPropertyTypes:true,noEmit:true,module:ts.ModuleKind.ESNext,moduleResolution:ts.ModuleResolutionKind.Bundler,target:ts.ScriptTarget.ES2024,lib:['lib.es2024.d.ts'],types:[],typeRoots:['/workspace/timer-ambient-delete/internal/load/node_types/node_modules/@types']};
const p=ts.createProgram([file,'/workspace/timer-ambient-delete/internal/load/node_types/node_modules/@types/node/index.d.ts'],options);
const checker=p.getTypeChecker(),sf=p.getSourceFile(file),globals={};
function visit(n){if(ts.isCallExpression(n)&&ts.isIdentifier(n.expression)&&['setTimeout','clearTimeout'].includes(n.expression.text)){const sym=checker.getSymbolAtLocation(n.expression);globals[n.expression.text]={type:checker.typeToString(checker.getTypeAtLocation(n.expression),undefined,ts.TypeFormatFlags.NoTruncation),signatures:checker.getTypeAtLocation(n.expression).getCallSignatures().map(s=>checker.signatureToString(s,undefined,ts.TypeFormatFlags.NoTruncation)),declarations:sym.declarations.map(d=>({file:d.getSourceFile().fileName,text:d.getText()}))};}ts.forEachChild(n,visit)}visit(sf);
const diagnostics=ts.getPreEmitDiagnostics(p).map(d=>{const pos=d.file&&d.file.getLineAndCharacterOfPosition(d.start);return {code:d.code,file:d.file?.fileName,line:pos&&pos.line+1,column:pos&&pos.character+1,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')}});
console.log(JSON.stringify({version:ts.version,options,globals,diagnostics},null,2));process.exit(diagnostics.length?1:0);
