'use strict';
const fs=require('node:fs'), assert=require('node:assert/strict'), ts=require('typescript');
const spec=require('./resolve-host.json');
const optional=['directoryExists','realpath','getCurrentDirectory','getDirectories','trace','getGlobalTypingsCacheLocation','useCaseSensitiveFileNames'];
const parameters={fileExists:['fileName'],readFile:['fileName'],directoryExists:['directoryName'],realpath:['path'],getCurrentDirectory:[],getDirectories:['path'],trace:['s'],getGlobalTypingsCacheLocation:[]};
const sf=ts.createSourceFile('host.ts',spec.after,ts.ScriptTarget.Latest,true);
const forwarders=[];
function visit(n) {
    if(ts.isArrowFunction(n)) {
        assert(n.parameters.every(p=>!p.dotDotDotToken&&!p.initializer&&!p.questionToken));
        assert(ts.isCallExpression(n.body));
        let access=n.body.expression;
        if(ts.isNonNullExpression(access)) access=access.expression;
        assert(ts.isPropertyAccessExpression(access));
        assert.equal(access.expression.getText(sf),'nodeSystem');
        const name=access.name.text, wanted=parameters[name]; assert(wanted);
        assert.deepEqual(n.parameters.map(p=>p.name.getText(sf)),wanted);
        assert.deepEqual(n.body.arguments.map(a=>a.getText(sf)),wanted);
        forwarders.push(name);
    }
    ts.forEachChild(n,visit);
}
visit(sf); assert.deepEqual(forwarders.sort(),Object.keys(parameters).sort());
const emitted=ts.transpileModule(spec.after.replace(/const modulePath = resolveJSModule[^;]+;$/, 'module.exports = moduleResolutionHost;'),{compilerOptions:{target:ts.ScriptTarget.ES2024,alwaysStrict:false}}).outputText;
let presenceCases=0;
for(let mask=0;mask<(1<<optional.length);mask++) {
    const system={fileExists(){throw Error('not called');},readFile(){throw Error('not called');}};
    optional.forEach((k,i)=>{if(mask&(1<<i))system[k]=k==='useCaseSensitiveFileNames'?false:function(){throw Error('not called');};});
    const m={exports:undefined};new Function('nodeSystem','module',emitted)(system,m);
    assert.deepEqual(Object.keys(m.exports).sort(),Object.keys(system).sort());
    for(const k of optional)assert.equal(k in m.exports,k in system,k);
    presenceCases++;
}
console.log(JSON.stringify({forwarders,exactParameters:true,noHiddenRest:true,nodeSystemReceiver:true,presenceCases,undefinedSlots:0,noCallsDuringPresenceCheck:true},null,2));
