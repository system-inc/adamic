// Checker measurement tooling. Writes only to a new output directory.
const fs=require('fs'),path=require('path');
const ts=require('typescript');
const tree=path.resolve(process.argv[2]||'');
const output=path.resolve(process.argv[3]||'');
if(process.argv.length!==4)throw Error('usage: NODE_PATH=<stage3/api/node_modules> node measure.cjs <adapted-tree> <new-output-dir>');
if(fs.existsSync(output))throw Error('refusing existing output directory');
fs.mkdirSync(output,{recursive:true});
const evidence=path.join(__dirname,'evidence');
const prelude='/adamic-prelude/adamic.d.ts', requirePrelude='/adamic-prelude/node-require.d.ts';
const nodeIndex=path.join(path.dirname(require.resolve('@types/node/package.json')),'index.d.ts');
if(JSON.parse(fs.readFileSync(path.join(path.dirname(nodeIndex),'package.json'),'utf8')).version!=='25.3.3')throw Error('expected @types/node 25.3.3');
if(ts.version!=='6.0.3')throw Error(ts.version);
const config=ts.readConfigFile(tree+'/src/compiler/tsconfig.json',ts.sys.readFile);
const parsed=ts.parseJsonConfigFileContent(config.config,ts.sys,tree+'/src/compiler',undefined,tree+'/src/compiler/tsconfig.json');
const adam={strict:true,exactOptionalPropertyTypes:true,noUncheckedIndexedAccess:true,erasableSyntaxOnly:true,verbatimModuleSyntax:true,allowImportingTsExtensions:true,noEmit:true,module:ts.ModuleKind.ESNext,moduleDetection:ts.ModuleDetectionKind.Force,moduleResolution:ts.ModuleResolutionKind.Bundler,target:ts.ScriptTarget.ES2024,lib:['lib.es2024.d.ts'],types:[]};
const roots=ts.sys.readDirectory(tree+'/src/compiler',['.ts']);
function locate(node,pos){let out=node;ts.forEachChild(node,c=>{if(c.getFullStart()<=pos&&pos<c.end)out=locate(c,pos)});return out;}
function site(d,p){if(!d.file)return{};const f=d.file,n=locate(f,d.start);let owner=n;while(owner&&!ts.isFunctionLike(owner)&&!ts.isInterfaceDeclaration(owner)&&!ts.isTypeAliasDeclaration(owner)&&!ts.isClassDeclaration(owner))owner=owner.parent;const names=[];for(let o=owner;o;o=o.parent)if(o.name)names.unshift(o.name.getText(f));const ancestors=[];for(let o=n;o&&ancestors.length<6;o=o.parent)ancestors.push({kind:ts.SyntaxKind[o.kind],text:o.getText(f).slice(0,1500)});return{source:f.text.slice(d.start,d.start+d.length),owner:names.join('.')||'<module>',owner_line:owner?f.getLineAndCharacterOfPosition(owner.getStart(f)).line+1:1,ancestors};}
function run(name,options,files,adamicDeclarations=false){const host=ts.createCompilerHost(options);const read=host.readFile.bind(host);const exists=host.fileExists.bind(host);host.fileExists=f=>f===prelude||f===requirePrelude||exists(f);host.readFile=f=>{let text=f===prelude?fs.readFileSync(path.join(evidence,'prelude-node.txt'),'utf8'):f===requirePrelude?fs.readFileSync(path.join(evidence,'node-require.txt'),'utf8'):read(f);if(text===undefined)return text;if(adamicDeclarations&&f.endsWith('/lib.es5.d.ts'))text=text.replaceAll('interface RegExpExecArray extends Array<string>','interface RegExpExecArray extends Array<string | undefined>').replaceAll('interface RegExpMatchArray extends Array<string>','interface RegExpMatchArray extends Array<string | undefined>').replaceAll('split(separator: string | RegExp, limit?: number): string[];','split(separator: string, limit?: number): string[];\n    split(separator: RegExp, limit?: number): (string | undefined)[];');if(adamicDeclarations&&f.endsWith('/lib.es2018.regexp.d.ts'))text=text.replaceAll('[key: string]: string','[key: string]: string | undefined');if(adamicDeclarations&&f.endsWith('/lib.es2015.symbol.wellknown.d.ts'))text=text.replaceAll('[Symbol.split](string: string, limit?: number): string[]','[Symbol.split](string: string, limit?: number): (string | undefined)[]').replaceAll('split(splitter: { [Symbol.split](string: string, limit?: number): (string | undefined)[]; }, limit?: number): string[];','split(splitter: { [Symbol.split](string: string, limit?: number): (string | undefined)[]; }, limit?: number): (string | undefined)[];');return text;};const p=ts.createProgram({rootNames:files,options,host});const ds=ts.getPreEmitDiagnostics(p);const rows=ds.map(d=>{const pt=d.file?.getLineAndCharacterOfPosition(d.start||0);return{file:d.file?path.relative(tree,d.file.fileName):'',line:pt?pt.line+1:0,column:pt?pt.character+1:0,code:'TS'+d.code,message:ts.flattenDiagnosticMessageText(d.messageText,'\n'),...site(d,p)}});fs.writeFileSync(path.join(output,name+'.json'),JSON.stringify({version:ts.version,options,roots:files.map(f=>path.relative(tree,f)),config_errors:parsed.errors,rows},null,2));console.log(name,rows.length);}

// The production census disables erasableSyntaxOnly and adds its own declarations.
const production={...adam,erasableSyntaxOnly:false};
const censusRoots=[...roots,prelude,nodeIndex,requirePrelude];
run('own',parsed.options,parsed.fileNames);
run('census-inputs',production,censusRoots,true);
run('effective',production,roots);
run('adamic',adam,roots);
run('project-stricter',{...parsed.options,strictBindCallApply:true,useUnknownInCatchVariables:true,noUncheckedIndexedAccess:true,exactOptionalPropertyTypes:true,noEmit:true},parsed.fileNames);
for(const option of ['noUncheckedIndexedAccess','exactOptionalPropertyTypes','useUnknownInCatchVariables','strictBindCallApply'])
    run('census-without-'+option,{...production,[option]:false},censusRoots,true);
