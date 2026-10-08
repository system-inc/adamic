const fs=require('node:fs'),path=require('node:path'),cp=require('node:child_process'),assert=require('node:assert/strict'),ts=require('typescript');
const [compilerArg,scratchArg]=process.argv.slice(2), compiler=path.resolve(compilerArg), scratch=path.resolve(scratchArg);
assert.equal(ts.version,'6.0.3');fs.mkdirSync(scratch,{recursive:true});
const oracle=path.resolve(__dirname,'../../../../oracle/adamic.mjs');
const runtime=path.join(scratch,'node_modules/adamic');fs.mkdirSync(runtime,{recursive:true});
fs.copyFileSync(oracle,path.join(runtime,'index.mjs'));fs.writeFileSync(path.join(runtime,'package.json'),JSON.stringify({name:'adamic',type:'module',exports:'./index.mjs'}));
function run(command,label,env={}){
 const r=cp.spawnSync(command[0],command.slice(1),{encoding:'utf8',timeout:120000,maxBuffer:16*1024*1024,env:{...process.env,...env}});
 assert.equal(r.error,undefined,label);assert.equal(r.signal,null,label);
 fs.writeFileSync(path.join(scratch,label+'.stdout'),r.stdout);fs.writeFileSync(path.join(scratch,label+'.stderr'),r.stderr);
 return {command,exit:r.status,stdout:r.stdout,stderr:r.stderr};
}
const results=[];
for(const row of JSON.parse(fs.readFileSync(path.join(__dirname,'fixtures.json')))){
 const source=path.join(__dirname,row.file),text=fs.readFileSync(source,'utf8');
 const nodeFile=path.join(scratch,row.file+'.mjs');fs.writeFileSync(nodeFile,ts.transpileModule(text,{compilerOptions:{target:ts.ScriptTarget.ES2022,module:ts.ModuleKind.ESNext}}).outputText);
 const node=run([process.execPath,nodeFile],row.file+'.node');assert.deepEqual([node.exit,node.stdout,node.stderr],[0,row.node_stdout,'']);
 const refused=run([compiler,'c',source],row.file+'.a');assert.equal(refused.exit,1);assert.ok(refused.stderr.includes('refuses the non-null assertion !'));assert.ok(text.startsWith('// a-check: refused the non-null assertion !\n'));
 const tsPath=path.join(scratch,row.file.replace(/\.a$/,'.ts'));fs.writeFileSync(tsPath,text);
 const emitted=run([compiler,'c',tsPath,'--explain-checks'],row.file+'.c');assert.equal(emitted.exit,0,emitted.stderr);assert.ok(emitted.stderr.includes('non-null checks: proven '+row.proven+' checked '+row.checked));
 const binary=path.join(scratch,row.file+'.native');const build=run([compiler,'build',tsPath,'-o',binary,'--sanitize'],row.file+'.build');assert.equal(build.exit,0,build.stderr);
 const native=run([binary],row.file+'.native',{ASAN_OPTIONS:row.fails?'detect_leaks=0':'detect_leaks=1',UBSAN_OPTIONS:'halt_on_error=1'});
 const js=run([compiler,'js',tsPath],row.file+'.js');assert.equal(js.exit,0,js.stderr);
 const jsPath=path.join(scratch,row.file+'.compiled.mjs');fs.writeFileSync(jsPath,js.stdout);const javascript=run([process.execPath,jsPath],row.file+'.javascript');
 let assertion;const file=ts.createSourceFile(tsPath,text,ts.ScriptTarget.Latest,true);function visit(n){if(ts.isNonNullExpression(n))assertion=n;ts.forEachChild(n,visit);}visit(file);assert.ok(assertion);
 const pos=file.getLineAndCharacterOfPosition(assertion.getStart(file));
 const panic='adamic: panic: non-null assertion failed at '+tsPath+':'+(pos.line+1)+':'+(pos.character+1)+': '+assertion.getText(file)+' is null or undefined\n';
 for(const actual of [native,javascript])assert.deepEqual([actual.exit,actual.stdout,actual.stderr],row.fails?[70,row.runtime_stdout,panic]:[0,row.node_stdout,'']);
 const counted=path.join(scratch,row.file+'.counted');const countBuild=run([compiler,'build',tsPath,'-o',counted,'--count'],row.file+'.count-build');assert.equal(countBuild.exit,0,countBuild.stderr);const count=run([counted],row.file+'.count');
 results.push({file:row.file,site:row.site,node,adamic_source:refused,explain:emitted.stderr,native,javascript,count,temporary_typescript_control:tsPath});
}
fs.writeFileSync(path.join(__dirname,'observations.json'),JSON.stringify({compiler_revision:'93ebcf520c87eb55b47ada1d39189796959b2778',node:process.version,results},null,2)+'\n');
console.log('3 Node goldens, 3 exact .a refusals, 3 .ts controls on both backends, 3 count runs passed');
