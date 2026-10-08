'use strict';
const assert=require('node:assert/strict');
const ts=require(process.env.TSC_ADAPT_TYPESCRIPT || 'typescript');
module.exports=function additions(before,after) {
    if(before===after)return [];
    const source=ts.createSourceFile('api.d.ts',after,ts.ScriptTarget.Latest,true);
    assert.equal(source.parseDiagnostics.length,0);
    const matches=[];
    function visit(n){if(ts.isInterfaceDeclaration(n) && n.name.text==='JsonSourceFile')matches.push(n);ts.forEachChild(n,visit);}visit(source);
    assert.equal(matches.length,1,'ambiguous public JSON owner');
    const property=matches[0].members.find(m=>m.name?.getText(source)==='extendedSourceFiles');
    assert(property && ts.isPropertySignature(property) && property.questionToken,'missing truthful optional JSON member');
    assert(ts.isArrayTypeNode(property.type) && property.type.elementType.kind===ts.SyntaxKind.StringKeyword,'public JSON member must match exact target string[]');
    const projected=after.slice(0,property.getFullStart())+after.slice(property.end);
    assert.equal(projected,before,'API contains an addition beyond the reviewed JSON owner');
    return [{owner:'JsonSourceFile',property:'extendedSourceFiles',type:'string[]',line:source.getLineAndCharacterOfPosition(property.getStart(source)).line+1}];
};
