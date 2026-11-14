'use strict';
// Node-only path proof. Capture/counter overlays are evidence instrumentation,
// never adapter output. The direct checks below cover ONLY getDirectories/trace/getCurrentDirectory.
const fs = require('node:fs'), path = require('node:path'), Module = require('node:module'), assert = require('node:assert/strict');
const ts = require('typescript');
const names = ['fileExists', 'readFile', 'directoryExists', 'realpath', 'getCurrentDirectory', 'getDirectories', 'trace', 'getGlobalTypingsCacheLocation'];
const root = path.resolve(process.argv[4]);
fs.mkdirSync(path.join(root, 'node_modules', 'ok'), {recursive:true});
fs.mkdirSync(path.join(root, 'node_modules', 'broken'), {recursive:true});
fs.writeFileSync(path.join(root, 'node_modules', 'ok', 'package.json'), '{"main":"index.js"}');
fs.writeFileSync(path.join(root, 'node_modules', 'ok', 'index.js'), 'module.exports = {answer:42};');
// Missing Node10 entry + an exports map invokes the diagnostic Bundler retry.
fs.writeFileSync(path.join(root, 'node_modules', 'broken', 'package.json'), '{"main":"missing.js","exports":"./exported.js"}');
function load(bundle, mode, mutant) {
    let source = fs.readFileSync(bundle, 'utf8');
    if(mutant) {
        assert.equal(source.split('nodeSystem.readFile(fileName)').length,2);
        source=source.replace('nodeSystem.readFile(fileName)','(0, nodeSystem.readFile)(fileName)');
    }
    const sf = ts.createSourceFile(bundle, source, ts.ScriptTarget.Latest, true, ts.ScriptKind.JS);
    const sites=[];
    function visit(n) {
        if(ts.isCallExpression(n) && n.expression.getText(sf)==='resolveJSModule' && n.arguments[0]?.getText(sf)==='moduleName' && n.arguments[1]?.getText(sf)==='baseDir') sites.push(n);
        ts.forEachChild(n,visit);
    }
    visit(sf); assert.equal(sites.length,1);
    const site=sites[0], arg=site.arguments[2];
    const adapted=arg.getText(sf)==='moduleResolutionHost';
    assert.equal(adapted,mode==='change');
    let argText=arg.getText(sf);
    source=source.slice(0,arg.getStart(sf))+'globalThis.__resolveHostEvidence('+argText+')'+source.slice(arg.end);
    const m=new Module(bundle,module); m.filename=bundle; m.paths=Module._nodeModulePaths(path.dirname(bundle)); m._compile(source,bundle);
    return m.exports;
}
function run(bundle,mode,present,mutant=false) {
    let host;
    const forwarded=Object.fromEntries(names.map(k=>[k,0]));
    globalThis.__resolveHostEvidence = value => {
        host=value;
        if(mode==='change') for(const k of names) if(k in value) {
            const f=value[k];
            // Arity here is instrumentation only; adapter forwarders have exact parameters.
            value[k]=function(a) { forwarded[k]++; return Reflect.apply(f,this,arguments); };
        }
        return value;
    };
    const api=load(bundle,mode,mutant), system=api.sys;
    const calls=Object.fromEntries(names.map(k=>[k,[]]));
    system.getGlobalTypingsCacheLocation=function(){ assert.equal(this,system); return undefined; };
    if(present) system.trace=function(s){ assert.equal(this,system); assert.equal(s,'trace argument'); };
    else { delete system.realpath; delete system.trace; }
    for(const k of names) if(k in system) {
        const original=system[k];
        system[k]=function() { assert.equal(this,system,`receiver: ${k}`); calls[k].push([...arguments]); return Reflect.apply(original,this,arguments); };
    }
    const outputs=[];
    const perResolution=[];
    for(const name of ['ok','./node_modules/ok/index.js','broken']) {
        const previous=Object.fromEntries(names.map(k=>[k,calls[k].length]));
        const result=system.require(root,name);
        perResolution.push({name,resolved:!result.error,counts:Object.fromEntries(names.map(k=>[k,calls[k].length-previous[k]]))});
        if(mutant) { assert(result.error); assert.match(result.error.message,/receiver: readFile/); throw result.error; }
        outputs.push({name,module:result.module,modulePath:result.modulePath,error:result.error?.message});
        if(name==='broken') assert(result.error);
        else assert.deepEqual(result.module,{answer:42});
        assert(host);
        for(const k of names) assert.equal(k in host,k in system,`presence: ${k}`);
        assert.equal(host.useCaseSensitiveFileNames,system.useCaseSensitiveFileNames);
    }
    const pathCounts=Object.fromEntries(names.map(k=>[k,calls[k].length]));
    if(mode==='change') assert.deepEqual(forwarded,pathCounts);
    for(const k of names) {
        if(k==='trace'||k==='getDirectories'||k==='getCurrentDirectory'||(!present&&k==='realpath')) assert.equal(pathCounts[k],0,k);
        else assert(pathCounts[k]>0,k+' not reached');
    }
    const direct={};
    // Forwarder checks, explicitly NOT a path proof.
    for(const [k,args] of [['getDirectories',[root]],['trace',['trace argument']],['getCurrentDirectory',[]]]) if(k in host) {
        const before=calls[k].length;
        Reflect.apply(host[k],host,args);
        assert.equal(calls[k].length,before+1);
        assert.deepEqual(calls[k].at(-1),args);
        direct[k]={calls:1,receiver:'nodeSystem',arguments:args};
    }
    delete globalThis.__resolveHostEvidence;
    return {outputs,pathCounts,perResolution,calls,direct};
}
const results=[];
for(const present of [true,false]) {
    const main=run(path.resolve(process.argv[2]),'base',present);
    const change=run(path.resolve(process.argv[3]),'change',present);
    const a=Buffer.from(JSON.stringify(main)),b=Buffer.from(JSON.stringify(change));
    assert(a.equals(b),'Node output bytes differ');
    fs.writeFileSync(path.join(root,'main-'+present+'.out'),a);
    fs.writeFileSync(path.join(root,'adapted-'+present+'.out'),b);
    results.push({optionalRealpath:present?'present':'absent',byteIdentical:true,bytes:a.length,...change});
}
assert.throws(()=>run(path.resolve(process.argv[3]),'change',true,true),/receiver: readFile/);
console.log(JSON.stringify({runtime:process.version,heldToNode:true,results,receiverMutantRed:true,receiverMutant:{member:'readFile',mutation:'(0, nodeSystem.readFile)(fileName)',observed:'receiver: readFile'}},null,2));
