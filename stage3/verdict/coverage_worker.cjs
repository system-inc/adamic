// Run the real CLI entry repeatedly, with fresh output/exit and per-case probes.
const fs = require('node:fs'), path = require('node:path'), readline = require('node:readline');
const Module=require('node:module'), vm=require('node:vm');
const bundle=path.resolve(process.argv[2]);
const wrapper=new vm.Script(Module.wrap(fs.readFileSync(bundle,'utf8')),{filename:bundle}).runInThisContext();
function freshCompiler() {
 const instance={exports:{}};
 wrapper.call(instance.exports,instance.exports,Module.createRequire(bundle),instance,bundle,path.dirname(bundle));
 return instance.exports;
}
readline.createInterface({input:process.stdin}).on('line', line=>{
 const job=JSON.parse(line); let stdout='', exit=0;
 process.chdir(job.cwd);
 const compiler=freshCompiler();
 const stop={};
 const system=Object.assign(compiler.verdictSystem,{args:job.args,
  write(text){stdout+=text;},exit(code){exit=code;throw stop;}});
 try { compiler.verdictExecute(system,job.args); }
 catch (error) { if (error!==stop) { process.stdout.write(JSON.stringify({error:String(error.stack || error)})+'\n');return; } }
 fs.writeFileSync(path.join(job.cwd,'.verdict-coverage.json'),JSON.stringify(Array.from(compiler.verdictHits.entries()).filter(x=>x[1]).map(x=>x[0])));
 process.stdout.write(JSON.stringify({exit,stdout_base64:Buffer.from(stdout,'utf8').toString('base64'),stderr:''})+'\n');
});
