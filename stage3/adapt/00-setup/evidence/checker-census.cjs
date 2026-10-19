'use strict';
const fs = require('node:fs');
const path = require('node:path');
const cp = require('node:child_process');
const crypto = require('node:crypto');
if (process.argv.length < 5) throw Error('usage: node checker-census.cjs REPOSITORY ADAPTED_TREE SCRATCH [--option-controls]');
const repo = path.resolve(process.argv[2]);
const tree = path.resolve(process.argv[3]);
const scratch = path.resolve(process.argv[4]);
fs.mkdirSync(scratch, {recursive:true});
const tsPath = path.join(repo, 'stage3/api/node_modules/typescript');
const ts = require(tsPath);
if (ts.version !== '6.0.3') throw Error('wrong TypeScript version');
const hash = text => crypto.createHash('sha256').update(text).digest('hex');
const walk = dir => fs.readdirSync(dir,{withFileTypes:true}).flatMap(e => e.isDirectory() ? walk(path.join(dir,e.name)) : [path.join(dir,e.name)]);
const roots = walk(path.join(tree,'src/compiler')).filter(p=>p.endsWith('.ts')).sort();
const options = {strict:true, noUncheckedIndexedAccess:true, exactOptionalPropertyTypes:true, erasableSyntaxOnly:false, verbatimModuleSyntax:true, allowImportingTsExtensions:true, noEmit:true, module:'esnext', moduleDetection:'force', moduleResolution:'bundler', target:'es2024', lib:['es2024'], types:[], skipLibCheck:false};
const original = fs.readFileSync(path.join(repo,'internal/load/prelude.d.ts'),'utf8');
let removed = 0;
const filtered = original.replace(/(interface (?:MapIterator|SetIterator)<T> \{\r?\n)([\t ]*next\(\): AdamicCollectionIteratorResult<T>;)/g, (_,start,line) => {removed++; return start + ' '.repeat(line.length);});
if(removed!==2) throw Error('expected exactly two iterator next lines');
const prelude = path.join(scratch,'prelude-without-collection-next.d.ts');
fs.writeFileSync(prelude,filtered);
const originalPrelude = path.join(scratch,'prelude-original.d.ts');
fs.writeFileSync(originalPrelude,original);
const libraryEdits = {
 'lib.es5.d.ts': [
  ['replacer: (substring: string, ...args: any[]) => string','replacer: (substring: string, ...args: any[]) => unknown'],
  ['interface RegExpExecArray extends Array<string>','interface RegExpExecArray extends Array<string | undefined>'],
  ['interface RegExpMatchArray extends Array<string>','interface RegExpMatchArray extends Array<string | undefined>'],
  ['split(separator: string | RegExp, limit?: number): string[];','split(separator: string, limit?: number): string[];\n    split(separator: RegExp, limit?: number): (string | undefined)[];']],
 'lib.es2021.string.d.ts': [['replacer: (substring: string, ...args: any[]) => string','replacer: (substring: string, ...args: any[]) => unknown']],
 'lib.es2018.regexp.d.ts': [['[key: string]: string','[key: string]: string | undefined']],
 'lib.es2015.symbol.wellknown.d.ts': [
  ['replacer: (substring: string, ...args: any[]) => string','replacer: (substring: string, ...args: any[]) => unknown'],
  ['[Symbol.split](string: string, limit?: number): string[]','[Symbol.split](string: string, limit?: number): (string | undefined)[]'],
  ['split(splitter: { [Symbol.split](string: string, limit?: number): (string | undefined)[]; }, limit?: number): string[];','split(splitter: { [Symbol.split](string: string, limit?: number): (string | undefined)[]; }, limit?: number): (string | undefined)[];']]
};
const patched = path.join(scratch,'loader-typescript');
fs.cpSync(path.join(tsPath,'lib'),path.join(patched,'lib'),{recursive:true});
for(const [name,edits] of Object.entries(libraryEdits)) {
 const file=path.join(patched,'lib',name); let text=fs.readFileSync(file,'utf8');
 for(const [before,after] of edits) {if(!text.includes(before))throw Error('library edit drift '+name);text=text.split(before).join(after);}
 fs.writeFileSync(file,text);
}
function label(p) {
 if(p.startsWith(tree+'/'))return path.relative(tree,p);
 if(p===prelude)return '<prelude-without-collection-next.d.ts>';
 if(p===originalPrelude)return '<prelude-original.d.ts>';
 if(p.startsWith(patched+'/lib/'))return '<loader-lib>/'+path.basename(p);
 if(p.startsWith(tsPath+'/lib/'))return '<stock-lib>/'+path.basename(p);
 return p;
}
function diagnostic(d) {
 const loc=d.file && d.start!==undefined ? d.file.getLineAndCharacterOfPosition(d.start):null;
 return {file:d.file?label(d.file.fileName):null,line:loc?loc.line+1:null,column:loc?loc.character+1:null,code:'TS'+d.code,message:ts.flattenDiagnosticMessageText(d.messageText,'\n'),...(d.relatedInformation?{relatedInformation:d.relatedInformation.map(diagnostic)}:{})};
}
const controls = process.argv[5] === '--option-controls';
const runs = controls ? JSON.parse(fs.readFileSync(path.join(scratch, 'raw.json'), 'utf8')).runs : [];
function run(name,libs,preludeFile,extra={}) {
 const files=roots.concat(preludeFile?[preludeFile]:[]);
 const config={compilerOptions:{...options,...extra},files};
 const configPath=path.join(scratch,name+'.json'); fs.writeFileSync(configPath,JSON.stringify(config,null,2)+'\n');
 const logPath=path.join(scratch,name+'.log'); const fd=fs.openSync(logPath,'w');
 const cli=path.join(libs==='loader'?patched:tsPath,'lib/tsc.js');
 const command=['timeout','60',process.execPath,cli,'--project',configPath,'--pretty','false'];
 const result=cp.spawnSync(command[0],command.slice(1),{cwd:tree,stdio:['ignore',fd,fd],timeout:65000});fs.closeSync(fd);
 if(result.error || result.signal || ![0,2].includes(result.status))throw Error('tsc CLI failed '+name+' '+JSON.stringify(result));
 const parsed=ts.parseJsonConfigFileContent(config,ts.sys,scratch);
 if(parsed.errors.length)throw Error('config errors');
 const host=ts.createCompilerHost(parsed.options);
 const libDir=path.join(libs==='loader'?patched:tsPath,'lib');
 host.getDefaultLibLocation=()=>libDir;host.getDefaultLibFileName=opts=>path.join(libDir,ts.getDefaultLibFileName(opts));
 const program=ts.createProgram({rootNames:files,options:parsed.options,host});
 const diagnostics=ts.getPreEmitDiagnostics(program);
 const text=ts.formatDiagnostics(diagnostics,{getCanonicalFileName:x=>x,getCurrentDirectory:()=>tree,getNewLine:()=> '\n'});
 const actual=fs.readFileSync(logPath,'utf8');
 if(text!==actual) {fs.writeFileSync(path.join(scratch,name+'.api.log'),text);throw Error('CLI/API mismatch '+name);}
 const sources=program.getSourceFiles();
 const record={name,libraryProfile:libs,prelude:preludeFile?label(preludeFile):null,options:config.compilerOptions,command:command.join(' '),exit:result.status,cliApiByteIdentical:true,logSha256:hash(actual),rootCount:roots.length,sourceFiles:sources.map(f=>({file:label(f.fileName),sha256:hash(f.text)})).sort((a,b)=>a.file.localeCompare(b.file)),diagnostics:diagnostics.map(diagnostic)};
 runs.push(record);console.log(name,record.exit,record.rootCount,record.sourceFiles.length,record.diagnostics.length);
 fs.writeFileSync(path.join(scratch,'raw.json'),JSON.stringify({typescriptVersion:ts.version,base:cp.execFileSync('git',['rev-parse','HEAD'],{cwd:repo,encoding:'utf8'}).trim(),roots:roots.map(label),prelude:{originalSha256:hash(original),filteredSha256:hash(filtered),removedNextDeclarations:removed},runs},null,2)+'\n');
}
if (controls) {
 run('stock-no-exact-control','stock',null,{exactOptionalPropertyTypes:false});
 run('stock-known-catch-control','stock',null,{useUnknownInCatchVariables:false});
} else {
run('stock-no-prelude','stock',null);
run('stock-filtered-prelude','stock',prelude);
run('loader-no-prelude','loader',null);
run('loader-filtered-prelude','loader',prelude);
run('stock-no-unchecked-control','stock',null,{noUncheckedIndexedAccess:false});
run('loader-original-prelude-mutant','loader',originalPrelude);
}
