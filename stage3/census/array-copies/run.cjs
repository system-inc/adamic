/* Bounded scratch build and one noEmit run per variant. All child output goes to files.
 * Usage: node run.cjs STOCK_ROOT SCRATCH_ROOT OUT_DIRECTORY
 * Requires completed stock npm run build:compiler and stock.json in OUT_DIRECTORY.
 */
const fs=require('fs'),path=require('path'),cp=require('child_process');
const [stockArg,scratchArg,outArg]=process.argv.slice(2),stock=path.resolve(stockArg),scratch=path.resolve(scratchArg),out=path.resolve(outArg);
const started=Date.now(),deadline=started+40*60*1000,steps=[];
function run(name,command,args,options={}){
 const log=path.join(out,name+'.log'),fd=fs.openSync(log,'w'),before=Date.now();
 const result=cp.spawnSync(command,args,{...options,stdio:['ignore',fd,fd],timeout:Math.max(1,deadline-Date.now()),killSignal:'SIGKILL'});
 fs.closeSync(fd);
 const row={name,command,args,cwd:options.cwd,exitCode:result.status,signal:result.signal,error:result.error?.message,seconds:(Date.now()-before)/1000,log:path.basename(log)};
 steps.push(row);console.log(`${name} child exit code: ${result.status}; signal: ${result.signal}`);
 fs.writeFileSync(path.join(out,'run-steps.json'),JSON.stringify({started:new Date(started).toISOString(),seconds:(Date.now()-started)/1000,steps},null,2)+'\n');
 if(result.status!==0)throw Error(`${name} failed: ${result.error?.message||result.signal||result.status}`);
}
if(fs.existsSync(scratch))throw Error('Scratch root must not exist');
fs.mkdirSync(scratch,{recursive:true});
fs.cpSync(path.join(stock,'src'),path.join(scratch,'src'),{recursive:true});
fs.copyFileSync(path.join(stock,'package.json'),path.join(scratch,'package.json'));
fs.symlinkSync(path.join(stock,'node_modules'),path.join(scratch,'node_modules'),'dir');
// Match official bundle options. Use same basename and same copied standard libraries.
const base=path.join(scratch,'baseline'),counted=path.join(scratch,'counted');
for(const dir of [base,counted]){
 fs.mkdirSync(dir);
 for(const name of fs.readdirSync(path.join(stock,'built/local')).filter(n=>/^lib.*\.d\.ts$/.test(n)))fs.copyFileSync(path.join(stock,'built/local',name),path.join(dir,name));
}
const esbuild=require(path.join(stock,'node_modules/esbuild'));
async function build(entry,outfile){await esbuild.build({entryPoints:[entry],outfile,bundle:true,platform:'node',target:['es2020','node14.17'],format:'cjs',sourcemap:'linked',sourcesContent:false,packages:'external',logLevel:'warning'});}
(async()=>{
 await build(path.join(stock,'src/tsc/tsc.ts'),path.join(base,'tsc.js'));
 run('transform',process.execPath,[path.join(out,'transform.cjs'),path.join(out,'../../api/node_modules/typescript/lib/typescript.js'),path.join(out,'stock.json'),scratch]);
 await build(path.join(scratch,'src/tsc/tsc.ts'),path.join(counted,'tsc.js'));
 const common=['--noEmit','-p',path.join(stock,'src/compiler/tsconfig.json'),'--pretty','false','--incremental','false'];
 // Composite projects cannot disable incremental. Use a scratch tsbuildinfo destination instead.
 common.splice(common.indexOf('--incremental'),2,'--tsBuildInfoFile',path.join(scratch,'compile.tsbuildinfo'));
 const env={...process.env,NODE_COMPILE_CACHE:undefined,ARRAY_COPIES_NODE_VERSION:process.version,ARRAY_COPIES_SITES:path.join(out,'stock.json')};
 run('baseline','python3',[path.join(out,'measure-child.py'),path.join(out,'baseline.json'),String(deadline/1000-2),process.execPath,path.join(base,'tsc.js'),...common],{cwd:stock,env});
 // noEmit on a fresh normal Program is not a build-mode cache hit. Remove metadata anyway.
 fs.rmSync(path.join(scratch,'compile.tsbuildinfo'),{force:true});
 run('counted','python3',[path.join(out,'measure-child.py'),path.join(out,'counted-memory.json'),String(deadline/1000-2),process.execPath,'--require',path.join(out,'counter.cjs'),path.join(counted,'tsc.js'),...common],{cwd:stock,env:{...env,ARRAY_COPIES_OUTPUT:path.join(out,'runtime.json')}});
})().catch(e=>{console.error(e);process.exitCode=1;});
