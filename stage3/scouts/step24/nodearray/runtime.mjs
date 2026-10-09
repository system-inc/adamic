// Run the upstream source on Node, independently transpiled by stock TypeScript.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import {createRequire,registerHooks} from 'node:module';
import {fileURLToPath,pathToFileURL} from 'node:url';
const require=createRequire(import.meta.url);
const stock=require(process.env.NODEARRAY_TYPESCRIPT);
const [sourceArg,inputsArg,manifestArg,outputArg]=process.argv.slice(2);
const source=path.resolve(sourceArg),inputs=path.resolve(inputsArg),output=path.resolve(outputArg);
const files=JSON.parse(fs.readFileSync(manifestArg,'utf8'));
const fields=['pos','end','hasTrailingComma','transformFlags','isMissingList','__tsDebuggerDisplay'];
const phases={},ids=new WeakMap(),allObjects=new WeakSet(),extraFields={};
let nextId=0,events=0;
function observe(value,phase,site) {
    if(!Array.isArray(value))return value;
    if(!ids.has(value)){ids.set(value,++nextId);allObjects.add(value);}
    events++;
    const bucket=phases[phase]??={events:0,unique_objects:0,seen:new WeakSet(),fields:{}};
    bucket.events++;
    if(!bucket.seen.has(value)){bucket.seen.add(value);bucket.unique_objects++;}
    for(const name of Object.getOwnPropertyNames(value))if(name!=='length'&&!/^\d+$/.test(name)&&!fields.includes(name))extraFields[name]=(extraFields[name]||0)+1;
    for(const field of fields) {
        const stats=bucket.fields[field]??={absent:0,present_undefined:0,present_value:0,inherited_value:0,examples:{}};
        const own=Object.hasOwn(value,field),present=field in value;
        const state=!present?'absent':!own?'inherited_value':value[field]===undefined?'present_undefined':'present_value';
        stats[state]++;
        stats.examples[state]??={site,object:ids.get(value),length:value.length,value:state==='present_value'?String(value[field]):undefined};
    }
    return value;
}
globalThis.__nodearrayObserve=observe;
globalThis.__nodearrayRead=(receiver,key,site)=>{observe(receiver,'read',site);return receiver[key];};
globalThis.__nodearrayWrite=(receiver,key,value,site)=>{observe(receiver,'before write',site);receiver[key]=value;observe(receiver,'after write',site);return value;};
registerHooks({
    resolve(specifier,context,nextResolve) {
        if(specifier.startsWith('.')&&specifier.endsWith('.js')) {
            const url=new URL(specifier.slice(0,-3)+'.ts',context.parentURL);
            if(fs.existsSync(fileURLToPath(url)))return {url:url.href,shortCircuit:true};
        }
        return nextResolve(specifier,context);
    },
    load(url,context,nextLoad) {
        if(url.endsWith('.ts')) {
            const text=fs.readFileSync(fileURLToPath(url),'utf8');
            const result=stock.transpileModule(text,{fileName:fileURLToPath(url),compilerOptions:{module:stock.ModuleKind.ESNext,target:stock.ScriptTarget.ESNext,verbatimModuleSyntax:false}});
            return {format:'module',source:result.outputText,shortCircuit:true};
        }
        return nextLoad(url,context);
    }
});
const ts=await import(pathToFileURL(path.join(source,'src/compiler/_namespaces/ts.ts')).href);
const digest=crypto.createHash('sha256'),corpus=[],passes={};
let lexicalArrays=nextId;
const scanner=ts.createScanner(ts.ScriptTarget.Latest,true);
let lexicalTokens=0;
for(const name of files) {
    const text=fs.readFileSync(path.join(inputs,name),'utf8');
    corpus.push({file:name,sha256:crypto.createHash('sha256').update(text).digest('hex'),bytes:Buffer.byteLength(text)});
    scanner.setText(text);
    while(true){const kind=scanner.scan();lexicalTokens++;if(kind===ts.SyntaxKind.EndOfFileToken)break;}
}
lexicalArrays=nextId-lexicalArrays;
function treeSnapshot(file,phase) {
    let arrays=0,nodes=0;
    const seen=new Set();
    function visit(node) {
        if(seen.has(node))return;
        seen.add(node);nodes++;
        digest.update(`N:${node.kind}:${node.pos}:${node.end}\n`);
        ts.forEachChild(node,visit,values=>{
            arrays++;observe(values,phase,`${file.fileName}:${node.pos}`);
            digest.update(`A:${values.length}:`);
            for(const field of fields) {
                const state=Object.hasOwn(values,field)?values[field]===undefined?'undefined':String(values[field]):field in values?'inherited':'absent';
                digest.update(`${field}=${state};`);
            }
            digest.update('\n');
            for(const child of values)visit(child);
        });
    }
    visit(file);return {arrays,nodes,parse_diagnostics:file.parseDiagnostics.length};
}
for(const mode of ['parse','incremental','debug']) {
    if(mode==='debug')ts.Debug.enableDebugInfo();
    const results=[];
    for(const name of files) {
        const text=fs.readFileSync(path.join(inputs,name),'utf8');
        const kind=name.endsWith('.json')?ts.ScriptKind.JSON:ts.ScriptKind.TS;
        let file=ts.createSourceFile(name,text,ts.ScriptTarget.Latest,true,kind);
        if(mode==='incremental') {
            // Insert a trivia-only prefix. Exercise reuse and range updates without changing declarations.
            const prefix='/* step24 incremental */\n';
            try {
                file=ts.updateSourceFile(file,prefix+text,{span:{start:0,length:0},newLength:prefix.length});
            } catch(error) {throw new Error('incremental input: '+name,{cause:error});}
        }
        results.push({file:name,...treeSnapshot(file,`${mode} final tree`)});
    }
    passes[mode]={files:results.length,arrays:results.reduce((a,r)=>a+r.arrays,0),nodes:results.reduce((a,r)=>a+r.nodes,0),parse_diagnostics:results.reduce((a,r)=>a+r.parse_diagnostics,0),per_file:results};
}
for(const bucket of Object.values(phases))delete bucket.seen;
const report={typescript:stock.version,node:process.version,files:files.length,lexical:{tokens:lexicalTokens,nodearrays_observed:lexicalArrays},passes,phases,observations:events,unique_array_objects:nextId,extra_own_fields:extraFields,semantic_digest:digest.digest('hex'),corpus};
fs.mkdirSync(output,{recursive:true});
fs.writeFileSync(path.join(output,'runtime.json'),JSON.stringify(report,null,2)+'\n');
console.log(JSON.stringify({files:report.files,lexical:report.lexical,passes:Object.fromEntries(Object.entries(passes).map(([k,v])=>[k,{files:v.files,arrays:v.arrays,nodes:v.nodes,parse_diagnostics:v.parse_diagnostics}])),events,unique_array_objects:nextId,extra_own_fields:extraFields,semantic_digest:report.semantic_digest},null,2));
