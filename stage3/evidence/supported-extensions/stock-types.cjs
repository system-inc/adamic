const fs=require('fs'),path=require('path'),ts=require('typescript'),assert=require('assert/strict');
const tree=process.argv[2];
const root=path.join(tree,'src/compiler');
const roots=fs.readdirSync(root,{recursive:true}).filter(n=>n.endsWith('.ts')).map(n=>path.join(root,n)).sort();assert.equal(roots.length,79);
const options={strict:true,noUncheckedIndexedAccess:true,exactOptionalPropertyTypes:true,verbatimModuleSyntax:true,allowImportingTsExtensions:true,noEmit:true,module:ts.ModuleKind.ESNext,moduleDetection:ts.ModuleDetectionKind.Force,moduleResolution:ts.ModuleResolutionKind.Bundler,target:ts.ScriptTarget.ES2024,lib:['lib.es2024.d.ts'],types:['node'],typeRoots:['/workspace/adamic/stage3/api/node_modules/@types']};
const program=ts.createProgram(roots,options),diagnostics=ts.getPreEmitDiagnostics(program);
console.log(ts.formatDiagnosticsWithColorAndContext(diagnostics,{getCanonicalFileName:f=>f,getCurrentDirectory:()=>process.cwd(),getNewLine:()=>String.fromCharCode(10)}));
console.log(JSON.stringify({typescript:ts.version,roots:roots.length,diagnostics:diagnostics.length}));process.exit(diagnostics.length?1:0);
