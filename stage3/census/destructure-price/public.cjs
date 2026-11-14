const ts=require('/tmp/destructure-price-scratch/work/stage3/api/node_modules/typescript'),fs=require('fs');
const dir='/tmp/destructure-price-scratch/adapted/built/local/';
const names=['Timer','PrintHandlers','ModuleSpecifierPreferences','NodeFactory','RecursiveDirectoryWatcherHost','CreateSystemWatchFunctions','TransformationContext','MutateMapOptions','MutateMapSkippingNewValuesOptions','EvaluationResolver'];
const bundled=fs.readFileSync(dir+'typescript.d.ts','utf8'), sf=ts.createSourceFile('typescript.d.ts',bundled,99,true), found={};
function visit(n){if(n.name&&names.includes(n.name.text)&&(ts.isInterfaceDeclaration(n)||ts.isTypeAliasDeclaration(n)||ts.isClassDeclaration(n)))found[n.name.text]={text:n.getText(sf),line:sf.getLineAndCharacterOfPosition(n.getStart()).line+1};ts.forEachChild(n,visit);}visit(sf);
const shim=fs.readFileSync(dir+'tsserverlibrary.d.ts','utf8');
fs.writeFileSync('/tmp/destructure-price-scratch/public-types.json',JSON.stringify({names:names.map(name=>({name,typescript:!!found[name],tsserverlibraryLiteral:false,tsserverlibraryReexport:!!found[name],declaration:found[name]||null})),shim},null,2));
console.log(names.map(n=>n+': '+!!found[n]).join('\n'));
