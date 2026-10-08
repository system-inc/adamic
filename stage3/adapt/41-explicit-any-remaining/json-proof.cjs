"use strict";
const fs=require("node:fs"),path=require("node:path"),assert=require("node:assert/strict"),vm=require("node:vm"),ts=require("typescript");
const [beforeTree,afterTree,builtModule,output]=process.argv.slice(2), runtime=require(path.resolve(builtModule));
const extract=(tree,file,names)=>{const s=ts.createSourceFile(file,fs.readFileSync(path.join(tree,"src/compiler",file),"utf8"),ts.ScriptTarget.Latest,true);return names.map(name=>{const nodes=s.statements.filter(n=>ts.isFunctionDeclaration(n)&&n.name?.text===name);assert.equal(nodes.length,1,name);return nodes[0].getText(s).replace(/^export /,"");}).join("\n");};
const configNames=["convertConfigFileToObject","convertToJson"],mapNames=["isStringOrNull","isRawSourceMap","tryParseRawSourceMap"];
const sources=tree=>extract(tree,"commandLineParser.ts",configNames)+"\n"+extract(tree,"sourcemap.ts",mapNames);
const compile=text=>vm.runInNewContext(ts.transpileModule(text+"\n({convertConfigFileToObject,convertToJson,tryParseRawSourceMap});",{compilerOptions:{target:ts.ScriptTarget.ESNext}}).outputText,{...runtime});
const jsonDomain=v=>v===undefined||v===null||["string","number","boolean"].includes(typeof v)||Array.isArray(v)&&v.every(jsonDomain)||typeof v==="object"&&Object.values(v).every(jsonDomain);
const run=functions=>{
 const inputs=['{}','{"a":[true,false,null,3,"text",{"b":-2}]}','{"bad":undefined,"array":[1,undefined,2]}','[1,{"a":2}]','7',''];
 const configs=inputs.map(text=>{const s=runtime.parseJsonText("case.json",text),errors=[];const value=functions.convertConfigFileToObject(s,errors,undefined);assert(value&&typeof value==="object"&&!Array.isArray(value),"config root object");assert(jsonDomain(value),"recursive converter domain");const checkErrors=[];const checked=functions.convertToJson(s,s.statements[0]?.expression,checkErrors,false,undefined);assert(jsonDomain(checked),"check-only recovery domain");return {value,errors:errors.map(e=>e.code),checked,checkErrors:checkErrors.map(e=>e.code)};});
 const maps=[{version:3,file:"x.js",mappings:"",sources:["a.ts"],sourcesContent:[null],names:[]},{version:3,file:"x",mappings:"",sources:[7]},null,[],3,{version:2,file:"x",mappings:"",sources:[]}].map(v=>functions.tryParseRawSourceMap(JSON.stringify(v)));
 assert(maps[0],"valid source map");assert(maps.slice(1).every(v=>v===undefined),"invalid shapes rejected");assert.equal(functions.tryParseRawSourceMap("{"),undefined,"parse exception retained");return {configs,maps};
};
const before=run(compile(sources(beforeTree))),after=run(compile(sources(afterTree)));assert.equal(JSON.stringify(after),JSON.stringify(before),"actual converter and source map observations");
const text=sources(afterTree);assert.equal(text.split('x.version === 3').length,2);assert.throws(()=>run(compile(text.replace('x.version === 3','x.version === 2'))),assert.AssertionError);
const evidence={inputs:8,recursiveDomain:true,configRootRecovery:true,invalidSourceMapsRejected:true,observationsIdentical:true,mutants:[{change:"actual source-map version check accepts 2 instead of 3",caughtBy:"runtime valid and invalid shape assertions"}]};fs.writeFileSync(output,JSON.stringify(evidence,null,2)+"\n");console.log(JSON.stringify(evidence));
