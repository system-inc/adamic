const fs=require('fs'),path=require('path');
const {spawnSync}=require('node:child_process');
const {nodeFor}=require('../../../internal/hostconfiguration/configuration.cjs');
// Stock compiler comparison with the selected loader's checker options.
const repo=path.resolve(process.argv[2]);
const logs=path.resolve(process.argv[3]);
fs.mkdirSync(logs,{recursive:true});
const bucket=__dirname;
// An explicit gate API uses the same bundled Node declarations the loader serves.
// Preserve the installed stage3 API path for the standalone adaptation proof.
const nodeModules=fs.realpathSync(repo+(process.env.CENSUS_TYPESCRIPT?'/internal/load/node_types/node_modules':'/stage3/api/node_modules'));
const ts=require(process.env.CENSUS_TYPESCRIPT||nodeModules+'/typescript');
if(ts.version!=='6.0.3'||require(nodeModules+'/@types/node/package.json').version!=='25.3.3')throw Error('pin');
const options={strict:true,noUncheckedIndexedAccess:true,exactOptionalPropertyTypes:true,erasableSyntaxOnly:false,verbatimModuleSyntax:true,allowImportingTsExtensions:true,noEmit:true,module:ts.ModuleKind.ESNext,moduleDetection:ts.ModuleDetectionKind.Force,moduleResolution:ts.ModuleResolutionKind.Bundler,target:ts.ScriptTarget.ES2024,lib:['lib.es2024.d.ts'],types:[]};
let prelude=fs.readFileSync(repo+'/internal/load/prelude.d.ts','utf8').replace('declare const console: {\n\tlog(message: string): void;\n\terror(message: string): void;\n};','declare var console: Console;');
const procStart=prelude.indexOf('declare const process: {');
if(procStart>=0){const end=prelude.indexOf('\n};',procStart);if(end>=0)prelude=prelude.slice(0,procStart)+'declare var process: NodeJS.Process;'+prelude.slice(end+3);}
const mutant=process.argv.includes('--mutant');
const all=[];
for(const [index,row] of JSON.parse(fs.readFileSync(bucket+'/status.json')).entries()){
 if(process.env.ADAMIC_HOST_STOCK_ROW && process.env.ADAMIC_HOST_STOCK_ROW!==row.file)continue;
 if(mutant&&row.file!=='23_newLine.a')continue;
 const nodeBinary=nodeFor(row,index+1);
 // This is a checker, not the fixture's Node oracle: stage0 still runs here
 // when the reference build is absent.
 if(nodeBinary && row.configuration && process.env.ADAMIC_HOST_STOCK_ROW!==row.file){
  const result=spawnSync(nodeBinary,[__filename,repo,logs,...(mutant?['--mutant']:[])],{env:{...process.env,ADAMIC_HOST_STOCK_ROW:row.file},encoding:'utf8'});
  if(result.error)throw result.error;
  if(result.status!==0)throw Error(result.stdout+result.stderr);
  const recorded=JSON.parse(fs.readFileSync(logs+'/'+(mutant?'stock-mutant':'stock')+'.json'));
  all.push(...recorded.results);continue;
 }
 const real=bucket+'/'+row.file,alias=real+'.ts',pp='/adamic-prelude/adamic.d.ts';
 const host=ts.createCompilerHost(options),read=host.readFile.bind(host),exists=host.fileExists.bind(host);
 host.fileExists=f=>f===alias||f===pp||exists(f);
 host.readFile=f=>f===alias?(fs.readFileSync(real,'utf8')+(mutant?'\nconst rerecordMutant: number = _os.EOL;\n':'')):f===pp?prelude:read(f);
 host.getSourceFile=(f,language)=>{const s=host.readFile(f);return s===undefined?undefined:ts.createSourceFile(f,s,language,true)};
 const program=ts.createProgram([alias,pp,nodeModules+'/@types/node/index.d.ts'],options,host);
 const syntax=program.getSyntacticDiagnostics();const ds=syntax.length?syntax:ts.getPreEmitDiagnostics(program);
 function format(d){const pos=d.file?d.file.getLineAndCharacterOfPosition(d.start):null;const name=d.file?d.file.fileName.replace(alias,real).replace(bucket+'/', 'stage3/fixtures/host/').replace(repo+'/',''):null;return(name?`${name}:${pos.line+1}:${pos.character+1}: `:'')+`error TS${d.code}: `+ts.flattenDiagnosticMessageText(d.messageText,'\n');}
 const diagnostics=ds.map(d=>({code:d.code,file:d.file?.fileName.replace(alias,real).replace(bucket+'/', 'stage3/fixtures/host/').replace(repo+'/',''),start:d.start,text:format(d)}));
 fs.writeFileSync(logs+'/'+row.file+(mutant?'.mutant':'')+'.tsc.log',diagnostics.map(d=>d.text).sort().join('\n')+(ds.length?'\n':''));
 all.push({file:row.file,exit:ds.length?1:0,diagnostics,...(!nodeBinary?{nodeReferenceSkipped:true}: {})});
}
fs.writeFileSync(logs+'/'+(mutant?'stock-mutant':'stock')+'.json',JSON.stringify({version:ts.version,options,results:all},null,2)+'\n');
console.log(all.map(x=>x.skipped?`${x.file}: skipped configuration ${x.configuration.name}`:`${x.file}: stock diagnostics=${x.diagnostics.length}`).join('\n'));

if(mutant){
 if(all.length!==1||!all[0].diagnostics.some(d=>d.code===2322))throw Error('SURVIVING stock checker mutant');
 console.log('caught stock checker mutant: TS2322 for _os.EOL assigned to number');
}else{
 const status=JSON.parse(fs.readFileSync(bucket+'/status.json'));
 for(const result of all){
  const row=status.find(r=>r.file===result.file);
  const expected=row.stage0.outcome==='Checker'?row.stage0.what:'';
  const actual=result.diagnostics.map(d=>d.text).join('\n');
  if(actual!==expected)throw Error('Stock comparison differs: '+result.file+'\n'+actual);
 }
 const skipped=all.filter(result=>result.nodeReferenceSkipped).length;
 console.log(skipped?`pass: stock checker agrees with every classified row; ${skipped} reference builds unavailable`:'pass: stock checker agrees with every classified row');
}
