// Enumerate syntax and checker types with the pinned stock compiler API.
const fs = require('node:fs'), path = require('node:path'), crypto = require('node:crypto');
const ts = require('typescript');
if (ts.version !== '6.0.3') throw Error('TypeScript 6.0.3 required');
const [treeArg, output, fixtureEntry] = process.argv.slice(2), tree = path.resolve(treeArg);
const entry = path.join(tree, fixtureEntry || 'src/tsc/tsc.ts');
const configPath = path.join(tree, 'src/compiler/tsconfig.json');
let options = {strict:true,target:ts.ScriptTarget.ES2020,module:ts.ModuleKind.ESNext,moduleResolution:ts.ModuleResolutionKind.Bundler,noEmit:true,types:[]};
if (fs.existsSync(configPath)) {
 const parsed = ts.getParsedCommandLineOfConfigFile(configPath, {}, {...ts.sys,onUnRecoverableConfigFileDiagnostic:d=>{throw Error(ts.flattenDiagnosticMessageText(d.messageText,'\n'));}});
 options = {...parsed.options,noEmit:true,composite:false,isolatedDeclarations:false};
}
// No project references: census the single source program, not built declarations.
const program = ts.createProgram([entry], options), checker = program.getTypeChecker();
const relative = name => path.relative(tree,name).split(path.sep).join('/');
const files = program.getSourceFiles().filter(f=>!f.isDeclarationFile && f.fileName.startsWith(tree+path.sep)).sort((a,b)=>a.fileName.localeCompare(b.fileName));
const sites = [], declarations = [];
function declarationType(node) {
 if(ts.isNamedTupleMember(node)) return checker.getTypeFromTypeNode(node.type);
 let symbol=checker.getSymbolAtLocation(node.name);
 if(symbol && (symbol.flags & ts.SymbolFlags.Alias)) symbol=checker.getAliasedSymbol(symbol);
 if(symbol && !(symbol.flags & ts.SymbolFlags.Value)) {
  if(symbol.flags & ts.SymbolFlags.Type) return checker.getDeclaredTypeOfSymbol(symbol);
  return undefined; // a type-only namespace has no value type to census
 }
 return symbol ? checker.getTypeOfSymbolAtLocation(symbol,node.name) : checker.getTypeAtLocation(node.name);
}
function row(file,node,kind,anchor=node) {
 const start=anchor.getStart(file), lc=file.getLineAndCharacterOfPosition(start);
 return {file:relative(file.fileName),kind,start,end:anchor.end,line:lc.line+1,column:lc.character+1,node_start:node.getStart(file),byte_node_start:Buffer.byteLength(file.text.slice(0,node.getStart(file)),'utf8'),node_end:node.end,byte_node_end:Buffer.byteLength(file.text.slice(0,node.end),'utf8'),syntax:ts.SyntaxKind[node.kind],text:node.getText(file).slice(0,400)};
}
for (const file of files) {
 function visit(node) {
  if (node.kind===ts.SyntaxKind.AnyKeyword) sites.push(row(file,node,'explicit_any'));
  if (ts.isDeclaration(node) && node.name) {
   const type=declarationType(node), r=row(file,node,'any_declaration',node.name);
   if(type) {
   r.name=node.name.getText(file); r.type=checker.typeToString(type); r.any=!!(type.flags & ts.TypeFlags.Any);
   declarations.push(r); if(r.any) sites.push(r);
   }
  }
  if(ts.isAsExpression(node)) {
   const token=node.getChildren(file).find(n=>n.kind===ts.SyntaxKind.AsKeyword);
   if(!token) throw Error('missing as token');
   const r=row(file,node,'as_cast',token), from=checker.getTypeAtLocation(node.expression), to=checker.getTypeAtLocation(node);
   r.source_type=checker.typeToString(from);r.target_type=checker.typeToString(to);
   r.source_any=!!(from.flags & ts.TypeFlags.Any);r.target_any=!!(to.flags & ts.TypeFlags.Any);
   r.assignable=checker.isTypeAssignableTo(from,to);r.const_assertion=ts.isTypeReferenceNode(node.type)&&node.type.typeName.getText(file)==='const';
   sites.push(r);
  }
  ts.forEachChild(node,visit);
 }
 visit(file);
}
const diagnostics=program.getSemanticDiagnostics().map(d=>({file:d.file?relative(d.file.fileName):null,start:d.start,code:d.code,message:ts.flattenDiagnosticMessageText(d.messageText,'\n')}));
const result={typescript:ts.version,entry:relative(entry),files:files.map(f=>({file:relative(f.fileName),sha256:crypto.createHash('sha256').update(f.text).digest('hex')})),sites,declarations,diagnostics};
fs.writeFileSync(output,JSON.stringify(result,null,2)+'\n');
console.log(JSON.stringify({files:files.length,sites:sites.length,diagnostics:diagnostics.length}));
