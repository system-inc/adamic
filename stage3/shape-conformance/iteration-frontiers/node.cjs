const fs=require('node:fs');
const ts=require('/home/agent/.cache/adamic-stage3/api/node_modules/typescript');
const source=fs.readFileSync(process.argv[2],'utf8');
const compiled=ts.transpileModule(source,{compilerOptions:{target:ts.ScriptTarget.ES2022,module:ts.ModuleKind.CommonJS}}).outputText;
require('node:vm').runInThisContext(compiled+'\nconsole.log(JSON.stringify([directIterationRead(),objectIterationRead(),arrayIterationRead(),hostIterationRead(),keyIterationRead(),emptyIterationRead(),initializedRead()].map(value=>value===undefined?"missing":typeof value.ready)));',{filename:process.argv[2]});
