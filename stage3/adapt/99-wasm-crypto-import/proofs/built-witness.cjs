'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),child=require('node:child_process'),crypto=require('node:crypto');
function hash(bytes){return crypto.createHash('sha256').update(bytes).digest('hex');}
function observe(file){
    const code=`const Module=require('node:module');const original=Module._load;const loads=[];Module._load=function(name,...args){if(['crypto','node:crypto'].includes(name))loads.push({name,parent:args[0]?.filename,stack:new Error().stack.split('\\n').slice(2,10)});return original.call(this,name,...args)};try{const ts=require(process.argv[1]);console.log(JSON.stringify({node:process.version,version:ts.version,loads}));}finally{Module._load=original;}`;
    return JSON.parse(child.execFileSync(process.execPath,['-e',code,path.resolve(file)],{encoding:'utf8',timeout:60000}));
}
function run(command,cwd){
    const r=child.spawnSync(process.execPath,command,{cwd,encoding:'utf8',timeout:60000});assert.ifError(r.error);assert.equal(r.status,0,r.stderr);
    return{command,exit:r.status,stdout:r.stdout,stderr:r.stderr};
}
function TestBuiltCryptoImport(before,after,scratch){
    before=path.resolve(before);after=path.resolve(after);scratch=path.resolve(scratch);
    const a=observe(path.join(before,'built/local/typescript.js')),b=observe(path.join(after,'built/local/typescript.js'));
    assert.equal(a.node,'v24.19.0');assert.equal(b.node,a.node);assert.equal(b.version,a.version);
    const old=a.loads.filter(x=>x.parent===path.join(before,'built/local/typescript.js'));
    const next=b.loads.filter(x=>x.parent===path.join(after,'built/local/typescript.js'));
    assert(old.some(x=>x.stack.some(s=>s.includes('getNodeSystem'))),'main crypto load not observed inside getNodeSystem');
    assert(next.length>0,'crypto was not loaded at module scope');
    assert(next.every(x=>!x.stack.some(s=>s.includes('getNodeSystem'))),'crypto still loaded inside getNodeSystem');
    fs.mkdirSync(scratch,{recursive:true});
    fs.writeFileSync(path.join(scratch,'main.ts'),'export const answer: number = 6 * 7;\nexport function greet(name: string): string { return `hello ${name}`; }\n');
    fs.writeFileSync(path.join(scratch,'tsconfig.json'),JSON.stringify({compilerOptions:{strict:true,target:'es2024',module:'commonjs',declaration:true,skipLibCheck:true},files:['main.ts']}));
    const runs=[];
    for(const [label,tree] of [['control',before],['after',after]]){
        const command=[path.join(tree,'built/local/tsc.js'),'--project',path.join(scratch,'tsconfig.json'),'--outDir',path.join(scratch,label),'--pretty','false'];
        runs.push({label,...run(command,scratch)});
    }
    assert.deepEqual([runs[1].exit,runs[1].stdout,runs[1].stderr],[runs[0].exit,runs[0].stdout,runs[0].stderr]);
    const output=[];
    for(const file of ['main.js','main.d.ts']){
        const x=fs.readFileSync(path.join(scratch,'control',file)),y=fs.readFileSync(path.join(scratch,'after',file));assert.deepEqual(y,x);
        output.push({file,bytes:x.length,sha256:hash(x),byteIdentical:true});
    }
    // Keep compiler, lib, input and output PATHS identical too. The symlink selects
    // each untouched real built tree; preserve-symlinks also covers tsc's _tsc.js.
    // Remove outputs and metadata before each run, so the second cannot reuse the first.
    const alias=path.join(scratch,'compiler'),out=path.join(scratch,'incremental'),info=path.join(scratch,'project.tsbuildinfo');
    const incremental=[],metadata=[];
    for(const [label,tree] of [['control',before],['after',after]]){
        fs.rmSync(alias,{force:true});fs.symlinkSync(path.join(tree,'built/local'),alias,'dir');
        fs.rmSync(out,{recursive:true,force:true});fs.rmSync(info,{force:true});
        const command=['--preserve-symlinks','--preserve-symlinks-main',path.join(alias,'tsc.js'),'--project',path.join(scratch,'tsconfig.json'),'--outDir',out,'--incremental','--tsBuildInfoFile',info,'--pretty','false'];
        const result=run(command,scratch),bytes=fs.readFileSync(info),data=JSON.parse(bytes);
        const versions=data.fileInfos.map(x=>typeof x==='string'?x:x.version);assert(versions.length>1);assert(versions.every(v=>/^[0-9a-f]{64}$/.test(v)),'incremental versions are not SHA-256');
        metadata.push(bytes);
        const emitted=['main.js','main.d.ts'].map(file=>({file,bytes:fs.readFileSync(path.join(out,file))}));
        for(const e of emitted){assert.deepEqual(e.bytes,fs.readFileSync(path.join(scratch,label,e.file)));}
        incremental.push({label,compilerTarget:fs.realpathSync(alias),...result,metadataBytes:bytes.length,metadataSha256:hash(bytes),fileVersions:versions.length,allVersionsAreSHA256:true,emitted:emitted.map(e=>({file:e.file,bytes:e.bytes.length,sha256:hash(e.bytes)}))});
        fs.writeFileSync(path.join(scratch,label+'.tsbuildinfo'),bytes);
    }
    assert.deepEqual(metadata[1],metadata[0],'.tsbuildinfo bytes differ');
    assert.deepEqual([incremental[1].exit,incremental[1].stdout,incremental[1].stderr],[incremental[0].exit,incremental[0].stdout,incremental[0].stderr]);
    return{test:'TestBuiltCryptoImport',before:a,after:b,cli:runs,output,incremental,tsbuildinfo:{byteIdentical:true,bytes:metadata[0].length,sha256:hash(metadata[0])},loadedDuringModuleEvaluation:true,compilerArtifactsUnmodified:true};
}
module.exports={TestBuiltCryptoImport};
if(require.main===module){const result=TestBuiltCryptoImport(...process.argv.slice(2,5));fs.writeFileSync(process.argv[5],JSON.stringify(result,null,2)+'\n');console.log(JSON.stringify(result,null,2));}
