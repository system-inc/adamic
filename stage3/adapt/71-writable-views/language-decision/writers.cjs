'use strict';
const fs=require('node:fs'),path=require('node:path'),vm=require('node:vm'),assert=require('node:assert/strict'),os=require('node:os'),child=require('node:child_process'),crypto=require('node:crypto');
const {ts,options}=require('./audit.cjs');
const library=path.resolve(process.argv[2]),mutation=process.argv[3];
if(!library.endsWith('/built/local/typescript.js'))throw Error('usage: node writers.cjs <tree>/built/local/typescript.js [flags|parent]');
const root='/unit71-actual-writers',file=root+'/main.ts',source=fs.readFileSync(path.join(__dirname,'controls/tsc-writers.a'),'utf8');
const settings={...options,module:ts.ModuleKind.CommonJS,moduleResolution:ts.ModuleResolutionKind.Node10,ignoreDeprecations:"6.0",noEmit:false,allowImportingTsExtensions:false};
const host=ts.createCompilerHost(settings),get=host.getSourceFile;
host.getSourceFile=(f,v,e,n)=>f===file?ts.createSourceFile(f,source,v,true,ts.ScriptKind.TS):get(f,v,e,n);
host.resolveModuleNames=names=>names.map(name=>name==='__TSC_PUBLIC__'?{resolvedFileName:library.replace(/\.js$/,'.d.ts'),extension:ts.Extension.Dts,isExternalLibraryImport:true}:undefined);
let javascript;host.writeFile=(f,s)=>{if(f.endsWith('main.js'))javascript=s;};
const program=ts.createProgram([file],settings,host),diagnostics=ts.getPreEmitDiagnostics(program);
assert.equal(diagnostics.length,0,diagnostics.map(d=>ts.flattenDiagnosticMessageText(d.messageText,'\n')).join('\n'));
assert.equal(program.emit().emitSkipped,false);
let runtime, scratch;
try {
    if(mutation){
        assert(['flags','parent'].includes(mutation));let text=fs.readFileSync(library,'utf8');
        const before=mutation==='flags'?'node.flags = newFlags;':'visited.parent = void 0;';
        const after=mutation==='flags'?'node.flags = node.flags;':'visited.parent = node.parent;';
        assert.equal(text.split(before).length,2,'unique actual writer');text=text.replace(before,after);
        scratch=fs.mkdtempSync(path.join(os.tmpdir(),'unit71-writer-mutant-'));const file=path.join(scratch,'typescript.cjs');fs.writeFileSync(file,text);runtime=require(file);
    } else runtime=require(library);
    const lines=[];vm.runInNewContext(javascript,{runtime,exports:{},console:{log:v=>lines.push(v)}});
    assert.equal(lines.length,1);const parts=lines[0].split(' ');
    assert.equal(parts[0],'16','initial legal flags holder');
    assert.equal(parts[1],'0','flags counterexample observation');
    assert.equal(parts[2],'true','original parent holder intact');
    assert.equal(parts[3],'undefined','parent counterexample observation');
    if(!mutation){
        const mutants=[];
        for(const name of ['flags','parent']){
            const result=child.spawnSync(process.execPath,[__filename,library,name],{encoding:'utf8'});
            assert.equal(result.status,1,name);assert(result.stderr.includes(name+' counterexample observation'),result.stderr);
            mutants.push({name,exit:result.status,caught_by:'external Node counterexample observation; stock TypeScript diagnostics remain zero'});
        }
        console.log(JSON.stringify({typescript:ts.version,diagnostics:0,library_sha256:crypto.createHash('sha256').update(fs.readFileSync(library)).digest('hex'),
            output:lines,flags_site:'src/compiler/utilities.ts:10689:10',parent_site:'src/compiler/utilities.ts:12365:6',
            witness:'controls/tsc-writers.a',scope:'actual unchanged built tsc exports, with legal narrowed Node holders; no claim these calls occur during the baseline compiler test run',mutants},null,2));
    }
} finally { if(scratch)fs.rmSync(scratch,{recursive:true}); }
