const assert=require('assert/strict');
const report=[];
for(const label of ['before','after-final']){
 const ts=require(`/workspace/scratch/${label}/built/local/typescript.js`);
 const options=['showConfig','configFile','configFilePath','help','init','listFiles','listEmittedFiles','project','build','version'];
 const input={options:{},fileNames:[],errors:[],projectReferences:[{path:'/workspace/project',originalPath:'./project'}]};
 const config=ts.convertToTSConfig(input,'/workspace/tsconfig.json',{useCaseSensitiveFileNames:true,getCurrentDirectory:()=>'/workspace'});
 for(const key of options){assert(Object.hasOwn(config.compilerOptions,key));assert.equal(config.compilerOptions[key],undefined);}
 assert(Object.hasOwn(config.references[0],'originalPath'));assert.equal(config.references[0].originalPath,undefined);assert.equal(config.references[0].path,'./project');
 const source=ts.parseJsonText('/workspace/tsconfig.json','{}');
 const parsed=ts.parseJsonSourceFileConfigFileContent(source,{useCaseSensitiveFileNames:true,readDirectory:()=>[],fileExists:()=>false,readFile:()=>undefined},'/workspace',{},'/workspace/tsconfig.json');
 const withSpecs=ts.convertToTSConfig(parsed,'/workspace/tsconfig.json',{useCaseSensitiveFileNames:true,getCurrentDirectory:()=>'/workspace'});
 for(const key of ['include','exclude']){assert(Object.hasOwn(withSpecs,key),key);assert.equal(withSpecs[key],undefined,key);}
 report.push({label,compilerOptionsOwnUndefinedKeys:options,referenceOriginalPathOwnUndefined:true,includeAndExcludeOwnUndefined:true,json:JSON.stringify(config),withSpecsJson:JSON.stringify(withSpecs)});
}
assert.equal(report[0].json,report[1].json);assert.equal(report[0].withSpecsJson,report[1].withSpecsJson);console.log(JSON.stringify(report,null,2));
