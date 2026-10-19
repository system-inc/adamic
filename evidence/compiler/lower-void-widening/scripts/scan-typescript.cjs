const ts = require('/workspace/adamic/stage3/api/node_modules/typescript');
const fs = require('fs');
const path = require('path');
const root = '/tmp/void-typescript';
const configPath = root + '/src/compiler/tsconfig.json';
const config = ts.readConfigFile(configPath, ts.sys.readFile);
const parsed = ts.parseJsonConfigFileContent(config.config, ts.sys, path.dirname(configPath));
parsed.options.typeRoots = ['/workspace/adamic/stage3/api/node_modules/@types'];
const program = ts.createProgram(parsed.fileNames, parsed.options);
const checker = program.getTypeChecker();
const hits = [];
function onlyVoid(type) {
 const calls = checker.getSignaturesOfType(type, ts.SignatureKind.Call);
 return calls.length > 0 && calls.every(sig => checker.getReturnTypeOfSignature(sig).flags === ts.TypeFlags.Void);
}
function wider(type) {
 if(type.isUnion() && type.types.some(wider)) return true;
 return checker.getSignaturesOfType(type, ts.SignatureKind.Call).some(sig => checker.getReturnTypeOfSignature(sig).flags !== ts.TypeFlags.Void);
}
for(const file of program.getSourceFiles()) {
 if(!file.fileName.startsWith(root+'/src/compiler/') || file.isDeclarationFile) continue;
 function visit(node) {
  if(ts.isExpressionNode(node) && !ts.isParenthesizedExpression(node)) {
   let source = checker.getTypeAtLocation(node), target = checker.getContextualType(node);
   let site = node;
   while(ts.isParenthesizedExpression(site.parent)) site = site.parent;
   const parent = site.parent;
   let returns = ts.isReturnStatement(parent) || (ts.isArrowFunction(parent) && parent.body===site);
   if(ts.isCallExpression(parent)||ts.isNewExpression(parent)) {
    const index = (parent.arguments || []).indexOf(site);
    const sig = checker.getResolvedSignature(parent);
    if(index>=0 && sig && sig.parameters.length) {
     const param = sig.parameters[Math.min(index,sig.parameters.length-1)];
     target = checker.getTypeOfSymbolAtLocation(param, site);
     if(param.valueDeclaration && param.valueDeclaration.dotDotDotToken) target = checker.getIndexTypeOfType(target, ts.IndexKind.Number);
    }
   }
   if(ts.isShorthandPropertyAssignment(parent)) {
    const symbol = checker.getShorthandAssignmentValueSymbol(parent);
    if(symbol) source = checker.getTypeOfSymbolAtLocation(symbol,node);
    const object = checker.getContextualType(parent.parent);
    if(object) {
     const prop = checker.getPropertyOfType(object,parent.name.text);
     if(prop) target = checker.getTypeOfSymbolAtLocation(prop,node);
    }
   }
   if(returns) {
    let fn = parent;
    while(fn && !ts.isFunctionLike(fn)) fn = fn.parent;
    if(fn) target = checker.getReturnTypeOfSignature(checker.getSignatureFromDeclaration(fn));
   }
   if(target && ((returns && source.flags===ts.TypeFlags.Void && target.flags!==ts.TypeFlags.Void) || (onlyVoid(source)&&wider(target)))) {
    const pos=file.getLineAndCharacterOfPosition(node.getStart());
    hits.push({file:path.relative(root,file.fileName),line:pos.line+1,column:pos.character+1,source:checker.typeToString(source),target:checker.typeToString(target),text:node.getText().slice(0,180)});
   }
  }
  ts.forEachChild(node,visit);
 }
 visit(file);
}
const diagnostics = ts.getPreEmitDiagnostics(program);
fs.writeFileSync('/tmp/void-evidence/typescript-impact.json',JSON.stringify({version:ts.version,commit:'050880ce59e30b356b686bd3144efe24f875ebc8',files:parsed.fileNames.length,diagnostics:diagnostics.length,diagnosticDetails:diagnostics.map(d=>ts.flattenDiagnosticMessageText(d.messageText,"\n")),hits},null,2)+'\n');
console.log(JSON.stringify({files:parsed.fileNames.length,diagnostics:diagnostics.length,sites:hits.length,hitFiles:new Set(hits.map(x=>x.file)).size}));
