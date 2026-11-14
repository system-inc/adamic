const ts=require('typescript'),fs=require('fs'),path=require('path');
const tree=process.argv[2];
const roots=ts.sys.readDirectory(path.join(tree,'src/compiler'),['.ts']);
const p=ts.createProgram(roots,{strict:true,exactOptionalPropertyTypes:true,noUncheckedIndexedAccess:true,erasableSyntaxOnly:false,verbatimModuleSyntax:true,allowImportingTsExtensions:true,noEmit:true,module:ts.ModuleKind.ESNext,moduleDetection:ts.ModuleDetectionKind.Force,moduleResolution:ts.ModuleResolutionKind.Bundler,target:ts.ScriptTarget.ES2024,lib:['lib.es2024.d.ts'],types:["node"],typeRoots:[path.join(__dirname,'work/stage3/api/node_modules/@types')]});
const diagnostics=ts.getPreEmitDiagnostics(p).map(d=>{const lc=d.file&&d.file.getLineAndCharacterOfPosition(d.start);return {file:d.file&&path.relative(tree,d.file.fileName),line:lc&&lc.line+1,column:lc&&lc.character+1,code:d.code,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')}});
fs.writeFileSync(process.argv[3],JSON.stringify({typescript:ts.version,roots:roots.length,diagnostics},null,2)+'\n');
