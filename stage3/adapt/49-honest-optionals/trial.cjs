// A proposal uses a virtual compiler host; failed trial edits never reach the input tree.
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),ts=require('typescript');
const root=path.resolve(process.argv[2]),out=path.resolve(process.argv[3]),rules=JSON.parse(fs.readFileSync(path.join(__dirname,'sites.json')));
const configPath=path.join(root,'src/compiler/tsconfig.json'),read=ts.readConfigFile(configPath,ts.sys.readFile),config=ts.parseJsonConfigFileContent(read.config,ts.sys,path.dirname(configPath),undefined,configPath),options={...config.options,noEmit:true,composite:false,isolatedDeclarations:false};
const texts=new Map();for(const file of new Set(rules.map(r=>r.file))){const original=fs.readFileSync(path.join(root,file),'utf8');texts.set(path.join(root,file),require('./plan.cjs').removeAssertions(file,original,rules.filter(r=>r.file===file)).text);}
function program(changes){const host=ts.createCompilerHost(options),original=host.getSourceFile.bind(host);host.getSourceFile=(name,...args)=>changes.has(name)?ts.createSourceFile(name,changes.get(name),ts.ScriptTarget.Latest,true):original(name,...args);return ts.createProgram([path.join(root,'src/tsc/tsc.ts')],options,host);}
const before=program(new Map()),after=program(texts);
const report=d=>({file:d.file?path.relative(root,d.file.fileName):null,...(d.file?{line:d.file.getLineAndCharacterOfPosition(d.start).line+1,column:d.file.getLineAndCharacterOfPosition(d.start).character+1}:{}),code:d.code,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')});
const result={before:before.getSemanticDiagnostics().map(report),after:after.getSemanticDiagnostics().map(report)};
fs.mkdirSync(out,{recursive:true});for(const [name,text] of texts){const file=path.join(out,path.relative(root,name));fs.mkdirSync(path.dirname(file),{recursive:true});fs.writeFileSync(file,text);}
fs.writeFileSync(path.join(out,'diagnostics.json'),JSON.stringify(result,null,2)+'\n');console.log(JSON.stringify(result,null,2));
