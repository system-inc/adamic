// Verify complete original union declarations and the four reached array sites.
const fs=require('node:fs'),path=require('node:path'),cp=require('node:child_process');
const ts=require('../../../../api/node_modules/typescript');
const [root,out]=process.argv.slice(2).map(p=>path.resolve(p));
if(!root||!out)throw Error('usage: prepare.cjs <pristine-upstream> <declarations>');
cp.execFileSync(process.execPath,[path.resolve(__dirname,'../../../lane4b/original/prepare.cjs'),root,out],{stdio:'inherit'});
const manifest=JSON.parse(fs.readFileSync(path.join(out,'original-manifest.json')));
const program=ts.createProgram([path.join(root,'src/compiler/builder.ts')],{target:ts.ScriptTarget.ES2024,module:ts.ModuleKind.ESNext,moduleResolution:ts.ModuleResolutionKind.Bundler,strict:true,types:[]});
const checker=program.getTypeChecker();
const exported=(file,name)=>{const source=program.getSourceFile(path.join(root,file));const symbol=checker.getExportsOfModule(checker.getSymbolAtLocation(source)).find(s=>s.name===name);if(!symbol)throw Error('missing '+name);return checker.getDeclaredTypeOfSymbol(symbol);};
manifest.fields={};
const target=exported('src/compiler/builder.ts','IncrementalBuildInfo');
for(const member of target.types) manifest.fields[checker.typeToString(member)]=checker.getPropertiesOfType(member).map(s=>s.name).sort();
const pair=JSON.parse(fs.readFileSync(path.resolve(__dirname,'../../lazy-array-priority.json'))).candidate_queue.find(p=>p.type==='IncrementalBuildInfo'&&p.field==='fileNames');
if(!pair||pair.read_count!==4||pair.sites.length!==4||pair.declared_type!=='readonly string[]') throw Error('fileNames census drift');
for(const site of pair.sites){const source=program.getSourceFile(path.join(root,site.file));if(!source||source.text.slice(site.start,site.end)!==site.text)throw Error('original read drift');let access;const visit=node=>{if(ts.isPropertyAccessExpression(node)&&node.getStart(source)===site.start&&node.end===site.end)access=node;ts.forEachChild(node,visit);};visit(source);if(!access||checker.typeToString(checker.getTypeAtLocation(access))!=='readonly string[]')throw Error('original field contract drift');}
manifest.pairs=[pair];
fs.writeFileSync(path.join(out,'file-names-manifest.json'),JSON.stringify(manifest,null,2)+'\n');
console.log('Complete fileNames union: 1 pair / 4 original reads verified');
