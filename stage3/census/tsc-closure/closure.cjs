// Independently enumerate literal source edges, then compare stock TS's source program.
const fs=require('node:fs'), path=require('node:path'), crypto=require('node:crypto');
const ts=require('typescript');
if(ts.version!=='6.0.3') throw new Error('TypeScript 6.0.3 required');
const [rootArg,entryArg,output]=process.argv.slice(2), root=path.resolve(rootArg), entry=path.resolve(root,entryArg);
const files=new Map(),edges=[],host=[],unsupported=[];
const relative=file=>path.relative(root,file).split(path.sep).join('/');
function resolve(from,specifier){
 const base=path.resolve(path.dirname(from),specifier),stem=base.replace(/\.js$/,'');
 const options=[base,stem+'.ts',stem+'.a',path.join(base,'index.ts')];
 return options.find(p=>fs.existsSync(p)&&fs.statSync(p).isFile());
}
function visit(file){
 if(files.has(file))return;
 if(!file.startsWith(root+path.sep))throw new Error('source import escaped tree: '+file);
 const data=fs.readFileSync(file);files.set(file,{file:relative(file),bytes:data.length,sha256:crypto.createHash('sha256').update(data).digest('hex')});
 const ast=ts.createSourceFile(file,data.toString('utf8'),ts.ScriptTarget.Latest,true);
 function edge(node,specifier){
  const p=ast.getLineAndCharacterOfPosition(node.getStart(ast));
  const row={from:relative(file),specifier,line:p.line+1,column:p.character+1};
  if(!specifier.startsWith('.')){host.push(row);return;}
  const target=resolve(file,specifier);if(!target)throw new Error('unresolved source import: '+JSON.stringify(row));
  row.to=relative(target);edges.push(row);visit(target);
 }
 function walk(node){
  if((ts.isImportDeclaration(node)||ts.isExportDeclaration(node))&&node.moduleSpecifier){if(!ts.isStringLiteralLike(node.moduleSpecifier))throw new Error('nonliteral module');edge(node,node.moduleSpecifier.text);}
  if(ts.isImportTypeNode(node)&&ts.isLiteralTypeNode(node.argument)&&ts.isStringLiteralLike(node.argument.literal))edge(node,node.argument.literal.text);
  if(ts.isCallExpression(node)&&((ts.isIdentifier(node.expression)&&node.expression.text==='require')||node.expression.kind===ts.SyntaxKind.ImportKeyword)){
   if(node.arguments.length===1&&ts.isStringLiteralLike(node.arguments[0]))edge(node,node.arguments[0].text);
   else unsupported.push({file:relative(file),expression:node.getText(ast)});
  }
  ts.forEachChild(node,walk);
 }
 walk(ast);
}
visit(entry);
const program=ts.createProgram([entry],{target:ts.ScriptTarget.ES2020,module:ts.ModuleKind.ESNext,moduleResolution:ts.ModuleResolutionKind.Bundler,noResolve:false,noEmit:true,skipLibCheck:true,types:[]});
const independent=program.getSourceFiles().filter(f=>!f.isDeclarationFile&&f.fileName.startsWith(root+path.sep)).map(f=>relative(f.fileName)).sort();
const expected=[...files.values()].map(f=>f.file).sort();
if(JSON.stringify(independent)!==JSON.stringify(expected))throw new Error('stock program source closure disagrees: '+JSON.stringify({expected,independent}));
const result={root,entry:relative(entry),typescript:ts.version,files:[...files.values()].sort((a,b)=>a.file.localeCompare(b.file)),edges,host,unsupported,independent};
fs.writeFileSync(output,JSON.stringify(result,null,2)+'\n');console.log('closure '+expected.length+' files; outside compiler '+expected.filter(f=>!f.startsWith('src/compiler/')).join(', '));
