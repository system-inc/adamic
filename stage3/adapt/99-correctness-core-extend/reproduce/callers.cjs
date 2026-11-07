const fs=require('fs'),path=require('path'),ts=require('/workspace/adamic/stage3/api/node_modules/typescript');
const tree=process.argv[2];
const files=ts.sys.readDirectory(tree+'/src',['.ts'],undefined,['compiler/**/*.ts','services/**/*.ts','server/**/*.ts']);
const program=ts.createProgram(files,{target:ts.ScriptTarget.ES2020,module:ts.ModuleKind.NodeNext,moduleResolution:ts.ModuleResolutionKind.NodeNext,strict:true,skipLibCheck:true});
const checker=program.getTypeChecker(),rows=[];
for(const source of program.getSourceFiles())if(files.includes(source.fileName)){
function visit(node){if(ts.isCallExpression(node)){
const sig=checker.getResolvedSignature(node),d=sig?.declaration;
if(d&&ts.isFunctionDeclaration(d)&&['extend','copyProperties'].includes(d.name?.text)&&d.getSourceFile().fileName.endsWith('/compiler/core.ts')){
const pos=source.getLineAndCharacterOfPosition(node.getStart(source));
rows.push({owner:d.name.text,file:path.relative(tree,source.fileName),line:pos.line+1,call:node.getText(source),arguments:node.arguments.map(a=>{const t=checker.getTypeAtLocation(a),s=t.aliasSymbol||t.getSymbol();return {expression:a.getText(source),declared:(()=>{const e=ts.isBinaryExpression(a)?a.left:a;const symbol=checker.getSymbolAtLocation(ts.isPropertyAccessExpression(e)?e.name:e);const d=symbol?.valueDeclaration||symbol?.declarations?.[0];return d?{type:checker.typeToString(checker.getTypeOfSymbolAtLocation(symbol,d),d,ts.TypeFormatFlags.NoTruncation),text:d.getText(),file:path.relative(tree,d.getSourceFile().fileName)}:ts.isCallExpression(e)?{returnType:checker.typeToString(checker.getReturnTypeOfSignature(checker.getResolvedSignature(e)),e,ts.TypeFormatFlags.NoTruncation)}:null;})(),type:checker.typeToString(t,a,ts.TypeFormatFlags.NoTruncation),stringIndex:checker.getIndexTypeOfType(t,ts.IndexKind.String)&&checker.typeToString(checker.getIndexTypeOfType(t,ts.IndexKind.String)),classification:checker.getIndexTypeOfType(t,ts.IndexKind.String)?'string index signature':s?.declarations?.some(ts.isInterfaceDeclaration)?'fixed interface':'other',declarations:s?.declarations?.map(x=>({file:path.relative(tree,x.getSourceFile().fileName),text:x.getText().slice(0,12000)}))};})});
}}ts.forEachChild(node,visit)}visit(source);
}
console.log(JSON.stringify({roots:files.length,rows},null,2));
