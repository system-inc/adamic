'use strict';
const assert = require('node:assert/strict'),fs = require('node:fs'),path = require('node:path'),cp = require('node:child_process');
assert.equal(process.argv.length,5,'usage: node witness.cjs <before library> <after library> <output>');
const [beforeFile,afterFile,out] = process.argv.slice(2).map(p=>path.resolve(p));
const before=require(beforeFile),after=require(afterFile);fs.mkdirSync(out,{recursive:true});
function cli(library,cwd,args,label) {
    const run=cp.spawnSync(process.execPath,['-e','const ts=require(process.argv[1]);ts.executeCommandLine(ts.sys,()=>{},process.argv.slice(2));',library,...args],{cwd,encoding:'utf8'});
    fs.writeFileSync(path.join(cwd,label+'.stdout.log'),run.stdout||'');fs.writeFileSync(path.join(cwd,label+'.stderr.log'),run.stderr||'');
    return run;
}
const commands=[['ordinary',['--init']],['strict-false',['--init','--strict','false']],...['outFile','strict','target','rootDir','lib'].map(k=>[k+'-null',['--init','--'+k,'null']])];
const rows=[];
for(const [name,args] of commands) {
    const row={name,args};let original, originalRead;
    for(const [label,file] of [['before',beforeFile],['after',afterFile]]) {
        const cwd=path.join(out,name,label);fs.mkdirSync(cwd,{recursive:true});fs.writeFileSync(path.join(cwd,'index.ts'),'export {};\n');
        const init=cli(file,cwd,args,'init');assert.equal(init.status,0,init.stdout);
        const config=fs.readFileSync(path.join(cwd,'tsconfig.json'));
        if(label==='before')original=config;else assert(original.equals(config),name+' output changed');
        const read=cli(file,cwd,['--showConfig'],'readback');
        if(name==='lib-null') {
            assert.equal(read.status,1);
            assert((read.stderr||'').includes("TypeError: Cannot read properties of undefined (reading 'map')"));
        } else assert.equal(read.status,0,read.stdout);
        const outcome={exit:read.status,exception:((read.stderr||'').match(/^TypeError:.*$/m)||[])[0]||null};
        if(label==='before')originalRead=outcome;else assert.deepEqual(outcome,originalRead,name+' readback changed');
        row[label]={bytes:config.length,initExit:init.status,readbackExit:read.status,exception:outcome.exception};
    }
    row.byteIdentical=true;rows.push(row);
}
function optionsFor(name) {
    if(name==='paths-toJSON') {const paths={'@app/*':['./src/*']};Object.defineProperty(paths,'toJSON',{value(){return undefined;}});return {paths};}
    if(name==='plugins-toJSON')return {plugins:[{name:'example',toJSON(){return undefined;}}]};
    if(name==='plugins-two-undefined')return {plugins:[{name:'a',toJSON(){return undefined;}},{name:'b',toJSON(){return undefined;}}]};
    if(name==='custom-map') {
        const plugins=[{name:'a',toJSON(){return undefined;}},{name:'b',toJSON(){return undefined;}}];
        plugins.map=function(callback,thisArg){return Array.prototype.map.call(this,callback,thisArg).filter(value=>value!==undefined);};
        return {plugins};
    }
    if(name==='array-hole')return {types:new Array(2)};
    if(name==='array-explicit-undefined')return {types:[undefined,'x',undefined]};
    if(name==='own-null')return {strict:null};
    if(name==='empty-string')return {outFile:''};
    if(name==='zero')return {maxNodeModuleJsDepth:0};
    if(name==='empty-array')return {types:[]};
    if(name==='inherited')return Object.create({strict:undefined});
    throw Error(name);
}
const api=[];
for(const name of ['paths-toJSON','plugins-toJSON','plugins-two-undefined','custom-map','array-hole','array-explicit-undefined','own-null','empty-string','zero','empty-array','inherited']) {
    const a=before.generateTSConfig(optionsFor(name),'\n'),b=after.generateTSConfig(optionsFor(name),'\n');assert.equal(b,a,name+' output changed');
    for(const [label,text] of [['before',a],['after',b]])fs.writeFileSync(path.join(out,name+'-'+label+'.tsconfig.json.txt'),text);
    if(name==='paths-toJSON')assert(b.includes('"paths": undefined,'));
    if(name==='plugins-toJSON')assert(b.includes('"plugins": [],'));
    if(name==='custom-map')assert(b.includes('"plugins": [],'));
    if(name==='plugins-two-undefined')assert(b.includes('"plugins": [, ],'));
    if(name==='array-explicit-undefined')assert(b.includes('"types": [, "x", ],'));
    api.push({name,byteIdentical:true,bytes:Buffer.byteLength(a)});
}
const thrown=[];
for(const name of ['toJSON-throws','invalid-enum']) {
    function invoke(ts) {
        if(name==='invalid-enum')return ts.generateTSConfig({target:-12345},'\n');
        const paths={a:['b']};Object.defineProperty(paths,'toJSON',{value(){throw new Error('hook failure');}});return ts.generateTSConfig({paths},'\n');
    }
    const observed=[];
    for(const ts of [before,after]){try{invoke(ts);assert.fail('expected throw');}catch(error){observed.push({name:error.name,message:error.message});}}
    assert.deepEqual(observed[0],observed[1]);assert.equal(observed[0].name,'Error');thrown.push({case:name,error:observed[0]});
}
const result={cli:rows,api,thrown};fs.writeFileSync(path.join(out,'result.json'),JSON.stringify(result,null,2)+'\n');console.log(JSON.stringify(result,null,2));
