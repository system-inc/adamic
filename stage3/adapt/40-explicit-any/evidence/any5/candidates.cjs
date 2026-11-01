'use strict';
const fs=require('node:fs'),path=require('node:path'),ts=require('typescript');
const tree=process.argv[2],configFile=path.join(tree,'src/compiler/tsconfig.json');
const read=ts.readConfigFile(configFile,ts.sys.readFile),config=ts.parseJsonConfigFileContent(read.config,ts.sys,path.dirname(configFile),{},configFile);
const specs=[['types.ts','ParsedCommandLine','raw'],['commandLineParser.ts','readConfigFile','config'],['commandLineParser.ts','parseConfigFileTextToJson','config'],['commandLineParser.ts','ParsedTsconfig','raw'],['tsbuildPublic.ts',null,'timerToBuildInvalidatedProject']];
const walk=(n,f)=>{f(n);ts.forEachChild(n,c=>walk(c,f));};
const owner=n=>{for(let p=n.parent;p;p=p.parent)if(p.name&&(ts.isInterfaceDeclaration(p)||ts.isFunctionDeclaration(p)))return p.name.text;};
const summarize=d=>({file:d.file&&path.relative(tree,d.file.fileName),line:d.file&&d.file.getLineAndCharacterOfPosition(d.start).line+1,code:d.code,message:ts.flattenDiagnosticMessageText(d.messageText,' ')});
const base=ts.createProgram(config.fileNames,config.options),baseline=ts.getPreEmitDiagnostics(base).map(summarize),keys=new Set(baseline.map(x=>JSON.stringify(x)));
const out=[];
for(const [file,enclosing,key]of specs){const filename=path.join(tree,'src/compiler',file),source=base.getSourceFile(filename),found=[];walk(source,n=>{if(ts.isPropertySignature(n)&&n.name?.text===key&&(!enclosing||owner(n)===enclosing))found.push(n);});if(found.length!==1)throw Error('ambiguous '+key);const d=found[0],text=source.text.slice(0,d.type.getStart(source))+'unknown'+source.text.slice(d.type.end);const host=ts.createCompilerHost(config.options),get=host.readFile;host.readFile=f=>path.resolve(f)===filename?text:get(f);const program=ts.createProgram(config.fileNames,config.options,host);const diagnostics=ts.getPreEmitDiagnostics(program).map(summarize);out.push({owner:enclosing,key,replacement:'unknown',addedDiagnostics:diagnostics.filter(x=>!keys.has(JSON.stringify(x))),baselineCount:baseline.length,total:diagnostics.length});}
console.log(JSON.stringify(out,null,2));
