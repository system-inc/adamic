const ts=require('/tmp/destructure-price-scratch/work/stage3/api/node_modules/typescript');
const fs=require('fs'),path=require('path');
const root='/tmp/destructure-price-scratch/adapted';
const config=ts.readConfigFile(root+'/src/compiler/tsconfig.json',ts.sys.readFile);
const cfg=ts.parseJsonConfigFileContent(config.config,ts.sys,root+'/src/compiler');
const program=ts.createProgram(cfg.fileNames,cfg.options), checker=program.getTypeChecker();
const out=[],files=new Set();
function loc(n){const s=n.getSourceFile(),p=s.getLineAndCharacterOfPosition(n.getStart());return {file:path.relative(root,s.fileName),line:p.line+1,column:p.character+1};}
function enclosing(n){for(let p=n.parent;p;p=p.parent)if(ts.isFunctionLike(p))return p.name?.getText()||p.parent?.name?.getText()||'<anonymous>';return '<top-level>';}
for(const sf of program.getSourceFiles())if(sf.fileName.startsWith(root+'/src/compiler/')&&!sf.isDeclarationFile){
 function walk(n){if(ts.isObjectBindingPattern(n)){
 files.add(sf.fileName); const owner=n.parent;
 const init=owner.initializer; const typ=init?checker.getTypeAtLocation(init):checker.getTypeAtLocation(n);
 const members=n.elements.map(e=>{const name=(e.propertyName||e.name).getText();const symbol=typ.getProperty(name);return {name,local:e.name.getText(),method:!!symbol?.declarations?.some(ts.isMethodSignature),declarations:symbol?.declarations?.map(d=>({...loc(d),kind:ts.SyntaxKind[d.kind],text:d.getText()}))||[]};});
 if(members.some(m=>m.method))out.push({...loc(n),enclosing:enclosing(n),type:checker.typeToString(typ),typeDeclarations:(typ.aliasSymbol||typ.symbol)?.declarations?.map(d=>({...loc(d),kind:ts.SyntaxKind[d.kind]}))||[],members,initializer:init?.getText()||null});
 }ts.forEachChild(n,walk);}walk(sf);
}
fs.writeFileSync('/tmp/destructure-price-scratch/ast-sites.json',JSON.stringify(out,null,2));
fs.writeFileSync('/tmp/destructure-price-scratch/destructuring-files.txt',[...files].sort().join('\n'));
console.log(JSON.stringify({roots:cfg.fileNames.length,destructuringFiles:files.size,methodSites:out.length}));
