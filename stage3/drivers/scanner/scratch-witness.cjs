'use strict';
const fs=require('node:fs'),vm=require('node:vm'),ts=require(process.env.SCANNER_TYPESCRIPT);
if(ts.version!=='6.0.3')throw Error('requires stock TypeScript 6.0.3');
const file=process.argv[2],source=fs.readFileSync(file,'utf8');
const output=ts.transpileModule(source,{fileName:file.replace(/\.a$/,'.ts'),compilerOptions:{target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.CommonJS}}).outputText;
const witnessModule={exports:{}};
vm.runInNewContext(output,{console,exports:witnessModule.exports,module:witnessModule,require,process,Buffer,Error,Map,Set,Uint16Array},{filename:file});
