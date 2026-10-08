const fs=require('node:fs');
const ts=require('/home/agent/.cache/adamic-stage3/api/node_modules/typescript');
const source=fs.readFileSync(process.argv[2],'utf8');
const compiled=ts.transpileModule(source,{compilerOptions:{target:ts.ScriptTarget.ES2022,module:ts.ModuleKind.CommonJS}}).outputText;
const owners=process.argv.slice(3);
require('node:vm').runInThisContext(compiled+'\nconsole.log(JSON.stringify(['+owners.map(n=>n+'()').join(',')+'].map(value=>typeof value.ready)));',{filename:process.argv[2]});
