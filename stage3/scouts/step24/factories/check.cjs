// Source Node is the oracle; backend comparisons are only claimed when compilation succeeds.
const fs=require('node:fs'),path=require('node:path'),os=require('node:os'),cp=require('node:child_process'),assert=require('node:assert/strict');
const ts=require(process.env.TYPESCRIPT_API||'/workspace/cache/tsc-census/npm/node_modules/typescript');
const base=__dirname,cli=process.argv[2]||'/workspace/cache/step24-adamic';
const work=fs.mkdtempSync(path.join(os.tmpdir(),'step24-factories-'));
const expected={
 '01-binary-complete':'BinaryExpression:3:+:4\n',
 '02-variable-complete':'answer:42\nundefined\n',
 '03-numeric-conditional':'0:1\n',
 '04-sourcefile-after-escape':'let answer = 42;:0\n',
 '05-parser-finish':'true:2:7:12\n',
};
function run(command,args,env={}){const r=cp.spawnSync(command,args,{encoding:'utf8',timeout:120000,env:{...process.env,...env}});if(r.error)throw r.error;return{command:[command,...args],exit:r.status,signal:r.signal,stdout:r.stdout,stderr:r.stderr};}
const oracle=path.resolve(base,'../../../../oracle/node.mjs');
function source(file,label){return run('node',['--disable-warning=ExperimentalWarning',oracle,file]);}
const observations=[],mutants=[];let nativeCount=0;
for(const[name,golden]of Object.entries(expected)){
 const file=path.join(base,'fixtures',name+'.a'),node=source(file,name);assert.equal(node.exit,0,name+' Node exit');assert.equal(node.stderr,'',name+' Node stderr');assert.equal(node.stdout,golden,name+' Node golden');
 const stage0=run(cli,['c',file]),header=fs.readFileSync(file,'utf8').split('\n')[0];
 if(header.startsWith('// a-check: refused ')){const reason=header.slice('// a-check: refused '.length);assert.equal(stage0.exit,1);assert.ok(stage0.stderr.includes('refuses '+reason),name+' expected refusal '+stage0.stderr);}else{assert.equal(stage0.exit,0,name+' clean compile '+stage0.stderr);}
 const result={file:'fixtures/'+name+'.a',node,stage0,header};
 if(stage0.exit===0){const binary=path.join(work,name),build=run(cli,['build',file,'-o',binary,'--sanitize']);assert.equal(build.exit,0,build.stderr);const native=run(binary,[],{ASAN_OPTIONS:'detect_leaks=1:halt_on_error=1',UBSAN_OPTIONS:'halt_on_error=1'});assert.equal(native.exit,node.exit);assert.equal(native.stdout,node.stdout);assert.equal(native.stderr,node.stderr);
 const js=run(cli,['js',file]);assert.equal(js.exit,0,js.stderr);const jsfile=path.join(work,name+'-backend.mjs');fs.writeFileSync(jsfile,js.stdout);const backend=run('node',['--disable-warning=ExperimentalWarning',oracle,jsfile]);assert.equal(backend.exit,node.exit,backend.stderr);assert.equal(backend.stdout,node.stdout);assert.equal(backend.stderr,node.stderr);
 const countedBinary=path.join(work,name+'-count'),countBuild=run(cli,['build',file,'-o',countedBinary,'--count']);assert.equal(countBuild.exit,0,countBuild.stderr);const counted=run(countedBinary,[]);assert.equal(counted.exit,0);assert.equal(counted.stdout,node.stdout);result.native=native;result.javascript=backend;result.counted=counted;nativeCount++;
 }
 observations.push(result);
 const mutantFile=path.join(base,'mutants',name+'.a'),mutant=source(mutantFile,name+'-mutant');assert.ok(mutant.stdout!==golden||mutant.exit!==node.exit||mutant.stderr!==node.stderr,name+' mutant survived Node golden');const mutantStage0=run(cli,['c',mutantFile]),mutantHeader=fs.readFileSync(mutantFile,'utf8').split('\n')[0];
 if(mutantHeader.startsWith('// a-check: type error '))assert.ok(mutantStage0.stderr.includes(mutantHeader.slice('// a-check: type error '.length)),mutantStage0.stderr);else if(mutantHeader.startsWith('// a-check: refused '))assert.ok(mutantStage0.stderr.includes('refuses '+mutantHeader.slice('// a-check: refused '.length)),mutantStage0.stderr);else assert.equal(mutantStage0.exit,0,mutantStage0.stderr);
 const record={file:'mutants/'+name+'.a',caughtBy:'source Node stdout/exit/stderr golden',node:mutant,stage0:mutantStage0,header:mutantHeader};
 if(mutantStage0.exit===0){const binary=path.join(work,name+'-mutant-native'),build=run(cli,['build',mutantFile,'-o',binary,'--sanitize']);assert.equal(build.exit,0,build.stderr);const native=run(binary,[],{ASAN_OPTIONS:'detect_leaks=1:halt_on_error=1',UBSAN_OPTIONS:'halt_on_error=1'});assert.ok(native.stdout!==golden||native.exit!==node.exit||native.stderr!==node.stderr,name+' native mutant survived');assert.equal(native.exit,mutant.exit);assert.equal(native.stdout,mutant.stdout);assert.equal(native.stderr,mutant.stderr);record.native=native;record.caughtBy+=' and native stdout/exit/stderr golden';}
 mutants.push(record);
 console.log('PASS '+name+' Node, header, '+(result.native?'native/JS/leaks/counts':'recorded current refusal')+'; mutant caught');
}
fs.writeFileSync(path.join(base,'data/fixtures.json'),JSON.stringify({nodeVersion:process.version,typescriptVersion:ts.version,observations,mutants},null,2)+'\n');
const counts=['# Fixture counts','', 'Five factory reductions and five source mutants. Three pending staged-object programs are current refusals; two compile. No internal/oracle fixture was added, so its global counts table is untouched.', '', '| Fixture | Native counted result |','| --- | --- |'];
for(const o of observations)counts.push('| '+o.file+' | '+(o.counted?o.counted.stderr.trim().replaceAll('\n','<br>'):'Refused, not executed natively')+' |');fs.writeFileSync(path.join(base,'counts.md'),counts.join('\n')+'\n');
console.log('PASS five source goldens, five mutants, '+nativeCount+' native/JS/leak/count checks');
