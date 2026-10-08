// Focused scout oracle: stock source on Node, then the two production backends.
const fs = require('node:fs');
const path = require('node:path');
const os = require('node:os');
const cp = require('node:child_process');
const assert = require('node:assert/strict');
const ts = require(process.env.STAGE3_TYPESCRIPT || 'typescript');
assert.equal(ts.version,'6.0.3');
const scratch = fs.mkdtempSync(path.join(os.tmpdir(),'adamic-step24-'));
const adamic = path.resolve(process.argv[2]);
const entries = [
    {name:'lookahead', output:'true:10:8:9:1::0\ntrue:10:8:9:1::0\n', mutant:'lookahead commits a truthy result', from:'if (!result || isLookahead)', to:'if (!result)'},
    {name:'tryparse-boolean', output:'true:11:11:11:2:next:4\nfalse:11:11:11:2:next:4\n', mutant:'false no longer rewinds', from:'if (!result || isLookahead)',to:'if (isLookahead)'},
    {name:'tryparse-node', output:'true:10:8:9:1::0\nconstructor:11:11:11:2:next:4\n',mutant:'accept a non-constructor literal',from:'literalNode.text === "constructor"',to:'literalNode.text !== "constructor"'},
    {name:'scan-range',output:'true:30:10:8:9:1::0:7\ntrue:30:10\n',mutant:'omit comment directive restoration',from:'commentDirectives = saveErrorExpectations;',to:'commentDirectives = 99;'},
];
const options = {strict:true,exactOptionalPropertyTypes:true,noUncheckedIndexedAccess:true,verbatimModuleSyntax:true,erasableSyntaxOnly:true,module:ts.ModuleKind.ESNext,moduleResolution:ts.ModuleResolutionKind.Bundler,target:ts.ScriptTarget.ES2024,lib:['lib.es2024.d.ts'],types:[],noEmit:true};
const prelude = path.join(scratch,'console.d.ts');
fs.writeFileSync(prelude,'declare const console: { log(message: string): void };\n');
function stock(source, label) {
    // Only the external scratch path is .ts; every committed program is .a.
    const input = path.join(scratch,label+'.ts');
    fs.writeFileSync(input,source);
    const program = ts.createProgram([input,prelude],options);
    const diagnostics = ts.getPreEmitDiagnostics(program);
    assert.equal(diagnostics.length,0,diagnostics.map(d=>ts.flattenDiagnosticMessageText(d.messageText,' ')).join('\n'));
    const output = ts.transpileModule(source,{compilerOptions:{...options,noEmit:false}}).outputText;
    const file = path.join(scratch,label+'.mjs');
    fs.writeFileSync(file,output);
    return run(process.execPath,[file]);
}
function run(command,args) {
    const r = cp.spawnSync(command,args,{encoding:'utf8',timeout:120000});
    if(r.error) throw r.error;
    return {status:r.status,signal:r.signal,stdout:r.stdout,stderr:r.stderr};
}
function same(actual,expected,label) { assert.deepEqual(actual,expected,label); }
const results=[];
const started=Date.now();
for(const entry of entries) {
    const sourcePath=path.join(__dirname,'fixtures',entry.name+'.a');
    const source=fs.readFileSync(sourcePath,'utf8');
    const expected={status:0,signal:null,stdout:entry.output,stderr:''};
    same(stock(source,entry.name),expected,entry.name+' stock source Node');
    same(run(process.execPath,['--disable-warning=ExperimentalWarning',path.resolve(__dirname,'../../../../oracle/node.mjs'),sourcePath]),expected,entry.name+' direct source Node');
    assert.equal(source.split(entry.from).length,2,'mutant target must be unique');
    const mutation=stock(source.replace(entry.from,entry.to),entry.name+'-mutant');
    assert.equal(mutation.status,0,'mutant must run, not fail type checking or compilation');
    assert.notEqual(mutation.stdout,expected.stdout,'mutant survived');
    const record={fixture:entry.name,sourceNode:expected,mutant:entry.mutant,mutantObservation:mutation,caughtBy:'stdout comparison to the stock source Node baseline',native:{},javascript:{}};
    for(const mode of ['native','javascript']) {
        const artifact=path.join(scratch,entry.name+(mode==='native'?'.native':'.backend.mjs'));
        const build=mode==='native'?run(adamic,['build',sourcePath,'-o',artifact,'--count','--sanitize']):run(adamic,['js',sourcePath]);
        if(build.status!==0) {
            const gap=entry.name==='lookahead'?'PrefixUnaryExpression on a number':entry.name==='tryparse-node'?'PrefixUnaryExpression on a value':null;
            assert(gap && build.stderr.includes(gap),entry.name+' unexpected compiler failure: '+build.stderr);
            record[mode]={compiled:false,build}; continue;
        }
        if(mode==='javascript') fs.writeFileSync(artifact,build.stdout);
        const observation=mode==='native'?run(artifact,[]):run(process.execPath,['--disable-warning=ExperimentalWarning',path.resolve(__dirname,'../../../../oracle/node.mjs'),artifact]);
        const counts=mode==='native'?observation.stderr:undefined;
        // Counted builds print counters to stderr; retain the raw evidence as well.
        if(mode==='native') {
            assert.equal(observation.status,0);
            assert.equal(observation.stdout,entry.output);
            assert(!/ERROR:|runtime error:|LeakSanitizer/.test(observation.stderr));
        } else same(observation,expected,entry.name+' backend Node');
        record[mode]={compiled:true,observation,...(counts?{counts}: {})};
        const mutantPath=path.join(scratch,entry.name+'-mutant.a');
        fs.writeFileSync(mutantPath,source.replace(entry.from,entry.to));
        const mutantArtifact=path.join(scratch,entry.name+'-mutant.'+(mode==='native'?'native':'mjs'));
        const mutantBuild=mode==='native'?run(adamic,['build',mutantPath,'-o',mutantArtifact,'--sanitize']):run(adamic,['js',mutantPath]);
        assert.equal(mutantBuild.status,0,entry.name+' mutant must compile');
        if(mode==='javascript') fs.writeFileSync(mutantArtifact,mutantBuild.stdout);
        const mutantRun=mode==='native'?run(mutantArtifact,[]):run(process.execPath,['--disable-warning=ExperimentalWarning',path.resolve(__dirname,'../../../../oracle/node.mjs'),mutantArtifact]);
        assert.equal(mutantRun.status,0,entry.name+' mutant must run without a sanitizer or compiler failure');
        assert.notEqual(mutantRun.stdout,expected.stdout,entry.name+' backend mutant survived');
        record[mode].mutantObservation=mutantRun;
        record[mode].mutantCaughtBy='stdout comparison to source Node';
    }
    results.push(record);
    console.log(`${entry.name}: source Node PASS, mutant KILLED, native ${record.native.compiled?'PASS':'BLOCKED'}, backend Node ${record.javascript.compiled?'PASS':'BLOCKED'}`);
}
const report={compilerCommit:cp.execFileSync('git',['rev-parse','HEAD'],{encoding:'utf8'}).trim(),node:process.version,typescript:ts.version,nproc:cp.execFileSync('nproc',{encoding:'utf8'}).trim(),cpuMax:fs.readFileSync('/sys/fs/cgroup/cpu.max','utf8').trim(),wallSeconds:(Date.now()-started)/1000,results};
fs.writeFileSync(path.join(__dirname,'observations.json'),JSON.stringify(report,null,2)+'\n');
const lines=['# Focused fixture counts','','Four entry programs; helper functions are embedded in each fixture. This scout is outside internal/oracle/testdata, so its registry is local to the authorized territory.','', '| Fixture | Source Node | Mutants killed | Native | Backend Node | Allocation counters |','|---|---|---:|---|---|---|'];
for(const r of results) lines.push(`| ${r.fixture}.a | pass | 1 | ${r.native.compiled?'pass':'blocked'} | ${r.javascript.compiled?'pass':'blocked'} | ${r.native.compiled?r.native.counts.trim().replaceAll('\n','; '):'unavailable: no native artifact'} |`);
fs.writeFileSync(path.join(__dirname,'counts.md'),lines.join('\n')+'\n');
console.log(`4 Node baselines pass; 4/4 semantic mutants killed. Native accepted ${results.filter(r=>r.native.compiled).length}/4. Wall ${report.wallSeconds}s. Scratch ${scratch}`);
