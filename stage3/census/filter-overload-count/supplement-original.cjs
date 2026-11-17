// Supplementary stock-checker inventory, not production Load evidence.
const ts=require('../../api/node_modules/typescript'),fs=require('fs'),path=require('path');
const root=process.argv[2], out=process.argv[3];
const names=ts.sys.readDirectory(root+'/src',['.ts'],['**/testRunner/**','**/typings/**']);
const opts={target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.NodeNext,moduleResolution:ts.ModuleResolutionKind.NodeNext,strict:true,skipLibCheck:true};
const p=ts.createProgram(names,opts),c=p.getTypeChecker(),rows=[],all=[];
function where(n){const f=n.getSourceFile(),lc=f.getLineAndCharacterOfPosition(n.getStart());return path.relative(root,f.fileName)+':'+(lc.line+1)+':'+(lc.character+1)}
for(const f of p.getSourceFiles()){if(!f.fileName.startsWith(root+'/src/'))continue;
function visit(n){if(ts.isCallExpression(n)){const s=c.getResolvedSignature(n),d=s?.declaration;if(d&&d.getSourceFile().fileName.endsWith('/src/compiler/core.ts')&&d.name?.getText()==='filter'){
const narrowing=d.parameters[1]?.type?.type&&ts.isTypePredicateNode(d.parameters[1].type.type);
all.push({where:where(n),declaration_where:where(d),narrowing:!!narrowing,call:n.getText()});
if(narrowing){let fn=n.parent;while(fn&&!ts.isFunctionLike(fn))fn=fn.parent;const arg=n.arguments[0],sym=c.getSymbolAtLocation(arg);
rows.push({where:where(n),enclosing_function:fn?.name?.getText()||'<anonymous>',function_where:fn&&where(fn),call:n.getText(),argument:arg.getText(),argument_type:c.typeToString(c.getTypeAtLocation(arg),arg,ts.TypeFormatFlags.NoTruncation),argument_declarations:sym?.declarations?.map(d=>({where:where(d),text:d.getText(),type_annotation:d.type?.getText()})),predicate:n.arguments[1].getText(),parent:n.parent.getText(),declaration_where:where(d),declaration:d.getText(),result_type:c.typeToString(c.getTypeAtLocation(n),n,ts.TypeFormatFlags.NoTruncation),start:n.getStart(),end:n.end,function_text:fn?.getText()});}}}
ts.forEachChild(n,visit);}visit(f);}
fs.writeFileSync(out,JSON.stringify(rows,null,2));fs.writeFileSync(out+'.manifest.json',JSON.stringify({label:'supplementary stock-checker exhaustive call traversal; not production Load evidence',options:opts,roots:names,source_files:p.getSourceFiles().filter(f=>f.fileName.startsWith(root+'/src/')).map(f=>f.fileName),all_filter_calls:all},null,2));console.log(rows.length);
