'use strict';
const fs=require('node:fs'),vm=require('node:vm'),{createRequire}=require('node:module'),path=require('node:path');
const [file,...args]=process.argv.slice(2),source=fs.readFileSync(file,'utf8'),originalExit=process.exit,originalArgv=process.argv;
const taskModule={exports:{}};
const context=vm.createContext({require:createRequire(file),module:taskModule,exports:taskModule.exports,process,console,Buffer,setTimeout,clearTimeout,setInterval,clearInterval,setImmediate,clearImmediate,__filename:file,__dirname:path.dirname(file)});
class CompilerExit{constructor(code){this.code=code;}}
process.argv=[originalArgv[0],file,...args];process.exit=code=>{vm.runInContext('dumpTracingLegend();',context);throw new CompilerExit(code||0);};
let code=0;try{vm.runInContext(source,context,{filename:file});vm.runInContext('dumpTracingLegend();',context);}catch(e){if(e instanceof CompilerExit)code=e.code;else throw e;}finally{process.exit=originalExit;process.argv=originalArgv;}
process.exitCode=code;
