'use strict';
const fs=require('node:fs'),path=require('node:path'),ts=require('typescript'),vm=require('node:vm'),assert=require('node:assert/strict');
const rules=require('./void-rules.json'),tree=process.argv[2];
function observe(rule, adapted, returnMutant=false) {
    const records=[],diagnostics=[],mappedKeys=[],file={};
    const context={result:undefined,diagnostics,mappedKeys,file,options:{},existing:{},symbol:{},baseName:'name',sourceRestType:{},targetRestType:{},reportUnreliableMarkers:false,node:{moduleSpecifier:{}},Diagnostics:{Cannot_find_module_or_type_declarations_for_side_effect_import_of_0:0},fileSymbol:{},nameType:{},targetType:{mapper:{}}};
    for(const name of ['getTypeFromTypeReference','getTypeFromImportTypeNode','getInternalSymbolName','getPropertiesOfType','getTypeOfSymbol','instantiateType','appendTypeMapping','getTypeParameterFromMappedType','resolveExternalModuleName','resolveExternalModuleSymbol','checkExpressionCached','combined'])
        context[name]=(...args)=>{records.push([name,args.length]);return 42;};
    let script;
    if(rule.parent==='ExpressionStatement')script=(adapted?rule.expression:'void '+rule.expression)+';';
    else {
        const source=ts.createSourceFile(rule.file,rule.before,ts.ScriptTarget.Latest,true);
        let arrow;
        function visit(n){if(ts.isArrowFunction(n)&&n.body.kind===ts.SyntaxKind.VoidExpression)arrow=n;ts.forEachChild(n,visit);}visit(source);
        assert(arrow);
        const old=arrow.getText(source);
        const replacement=rule.expression==='0'?'{}':'{ '+(returnMutant?'return ':'')+rule.expression+'; }';
        script='const test = '+(adapted?old.replace('void '+rule.expression,replacement):old)+'; result = test(file);';
    }
    vm.runInNewContext(ts.transpileModule(script,{compilerOptions:{target:ts.ScriptTarget.ESNext}}).outputText,context);
    return {records,diagnostics,mappedKeys,file,result:context.result};
}
for(const r of rules)assert.deepEqual(observe(r,true),observe(r,false),'effects/order/arrow return at '+r.file+':'+r.line);
const callback=rules.find(r=>r.expression==='diagnostics.push(diag)');
assert(callback);
assert.throws(()=>assert.deepEqual(observe(callback,true,true),observe(callback,false)),/Expected values/);
const report={sites:rules.length,statementCalls:rules.filter(r=>r.parent==='ExpressionStatement').length,arrowBodies:rules.filter(r=>r.parent==='ArrowFunction').length,returnValueMutant:'return diagnostics.push(diag) caught by undefined-result comparison'};
if(tree) {
    for(const file of [...new Set(rules.map(r=>r.file))]){
        const before=fs.readFileSync(path.join(tree,file),'utf8');
        require('./void.cjs').apply(tree);
        assert.equal(fs.readFileSync(path.join(tree,file),'utf8'),before,'second application changes no bytes');
    }
}
console.log(JSON.stringify(report));
