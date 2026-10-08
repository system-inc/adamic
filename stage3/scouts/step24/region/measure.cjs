const fs=require('node:fs'),path=require('node:path'),crypto=require('node:crypto'),assert=require('node:assert/strict');
const {makeTracker}=require('./tracker.cjs');
const scratch=path.resolve(process.argv[2]),tree=path.resolve(process.argv[3]),projects=path.resolve(process.argv[4]),out=path.resolve(process.argv[5]);
fs.mkdirSync(out,{recursive:true});
const tracker=globalThis.__region=makeTracker();
const ts=require(path.join(scratch,'instrumented.cjs')),control=require(path.join(scratch,'control.cjs'));
assert.equal(ts.version,'6.0.3');
const bootstrap=tracker.end([],ts);fs.writeFileSync(path.join(out,'bootstrap.json'),JSON.stringify(bootstrap,null,2)+'\n');
const hash=x=>crypto.createHash('sha256').update(x).digest('hex');
function syntaxDigest(api,roots){let text='';const seen=new Set();function visit(n){if(seen.has(n))return;seen.add(n);text+=`${n.kind}:${n.pos}:${n.end}:${n.flags}\n`;api.forEachChild(n,visit);for(const j of n.jsDoc||[])visit(j);}for(const root of roots)visit(root);return hash(text);}
function files(root){return fs.readdirSync(root,{withFileTypes:true}).flatMap(e=>e.isDirectory()?files(path.join(root,e.name)):e.isFile()?[path.join(root,e.name)]:[]).sort();}
const manifest=files(path.join(tree,'src/compiler')).map(file=>({path:path.relative(tree,file),sha256:hash(fs.readFileSync(file))}));
fs.writeFileSync(path.join(out,'scanner-inputs.json'),JSON.stringify(manifest,null,2)+'\n');
const started=Date.now();tracker.begin();
const roots=manifest.map(row=>ts.createSourceFile(row.path,fs.readFileSync(path.join(tree,row.path),'utf8'),ts.ScriptTarget.Latest,true));
const digest=syntaxDigest(ts,roots);const scanner=tracker.end(roots,ts);
const originals=manifest.map(row=>control.createSourceFile(row.path,fs.readFileSync(path.join(tree,row.path),'utf8'),control.ScriptTarget.Latest,true));
assert.equal(digest,syntaxDigest(control,originals),'Scanner corpus AST instrumentation changed semantics');
scanner.inputFiles=manifest.length;scanner.syntaxDigest=digest;scanner.wallSeconds=(Date.now()-started)/1000;
fs.writeFileSync(path.join(out,'scanner.json'),JSON.stringify(scanner,null,2)+'\n');console.log('scanner '+JSON.stringify(scanner.totals));
const rows=JSON.parse(fs.readFileSync(path.join(projects,'manifest.json'),'utf8'));assert.equal(rows.length,301);
const acceptance=[];const acceptanceStart=Date.now();
function compile(api,file){const config=api.readConfigFile(file,api.sys.readFile);if(config.error)throw Error('Cannot read config');const parsed=api.parseJsonConfigFileContent(config.config,api.sys,path.dirname(file));assert.equal(parsed.errors.length,0);const host=api.createCompilerHost(parsed.options,true);const program=api.createProgram(parsed.fileNames,parsed.options,host);const diagnostics=api.getPreEmitDiagnostics(program);const formatHost={getCurrentDirectory:()=>path.dirname(file),getCanonicalFileName:f=>f,getNewLine:()=> '\n'};const stdout=api.formatDiagnostics(diagnostics,formatHost);return {program,stdout,exit:diagnostics.length?2:0,digest:syntaxDigest(api,program.getSourceFiles())};}
for(const row of rows){const start=Date.now();tracker.begin();const actual=compile(ts,row.config);const counts=tracker.end(actual.program.getSourceFiles(),ts);const baseline=compile(control,row.config);
 assert.equal(actual.stdout,baseline.stdout,row.id+' instrumentation diagnostic diff');assert.equal(actual.digest,baseline.digest,row.id+' instrumentation AST diff');assert.equal(actual.exit,baseline.exit);
 assert.equal(actual.stdout,fs.readFileSync(path.join(row.golden,'golden.stdout'),'utf8'),row.id+' golden diff');assert.equal(actual.exit,Number(fs.readFileSync(path.join(row.golden,'golden.exit'),'utf8')));assert.equal(fs.readFileSync(path.join(row.golden,'golden.stderr'),'utf8'),'');
 counts.id=row.id;counts.wallSeconds=(Date.now()-start)/1000;counts.stdoutSha256=hash(actual.stdout);counts.syntaxDigest=actual.digest;counts.expectedExit=actual.exit;counts.instrumentationAgrees=true;counts.goldenAgrees=true;
 fs.writeFileSync(path.join(out,row.id+'.json'),JSON.stringify(counts,null,2)+'\n');acceptance.push({id:row.id,totals:counts.totals,rootSourceFiles:counts.rootSourceFiles,wallSeconds:counts.wallSeconds});
 console.log(row.id+' '+JSON.stringify(counts.totals));
}
const summary={node:process.version,sourcePin:'050880ce59e30b356b686bd3144efe24f875ebc8',runtimeDesign:'d4108344c3ea62ee0208cb26ef6f9eae20af01db',scanner:{inputFiles:scanner.inputFiles,totals:scanner.totals,wallSeconds:scanner.wallSeconds},bootstrap:bootstrap.totals,acceptance:{projects:acceptance.length,totals:{},wallSeconds:(Date.now()-acceptanceStart)/1000},cases:acceptance};
for(const row of acceptance)for(const [key,value]of Object.entries(row.totals))summary.acceptance.totals[key]=(summary.acceptance.totals[key]||0)+value;
fs.writeFileSync(path.join(out,'summary.json'),JSON.stringify(summary,null,2)+'\n');console.log('SUMMARY '+JSON.stringify(summary.acceptance));
