// Discovery only: replace the smallest enclosing executable body by a throw.
const fs=require('fs'),ts=require(process.env.STEP31_TYPESCRIPT);
const [file,line,column,order]=process.argv.slice(2);const source=fs.readFileSync(file,'utf8');
const sf=ts.createSourceFile(file,source,ts.ScriptTarget.Latest,true),pos=sf.getLineStarts()[+line-1]+(+column-1);
let node=sf;function find(n){if(n.getStart(sf)<=pos&&pos<n.end){node=n;ts.forEachChild(n,find);}}find(sf);
const path=require('path');
const relative=file.split('/tree/')[1];
const originalFile=path.join('/workspace/cache/step31-binder-driver-slice-v2',relative);
const program=ts.createProgram([originalFile],{target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.ESNext,moduleResolution:ts.ModuleResolutionKind.Bundler,strict:true,skipLibCheck:true});
const checker=program.getTypeChecker(),original=program.getSourceFile(originalFile);
function originals(predicate){const found=[];function visit(n){if(predicate(n))found.push(n);ts.forEachChild(n,visit);}visit(original);return found;}
let fn=node;while(fn&&!ts.isFunctionLike(fn))fn=fn.parent;
if(!fn||!fn.body){
 let declaration=node;while(declaration&&!ts.isVariableDeclaration(declaration))declaration=declaration.parent;
 if(!declaration?.initializer)throw Error(`needs explicit placeholder: ${file}:${line}:${column} (${ts.SyntaxKind[node.kind]})`);
 let start=declaration.initializer.getStart(sf),end=declaration.initializer.end; const marker=`binder discovery stop ${order}`;
 if(declaration.initializer.getText(sf).includes('binder discovery stop')){
  let statement=declaration;while(statement&&!ts.isVariableStatement(statement))statement=statement.parent;
  if(!statement||statement.declarationList.declarations.length!==1||!declaration.type)throw Error('direct-throw placeholder needs one annotated variable');
  start=statement.getStart(sf);end=statement.end;
  const replacement=source.slice(start,declaration.initializer.getStart(sf)).replace(/\bconst\b/,'let').replace(/=\s*$/,'').trimEnd()+`; throw ${JSON.stringify(marker)};`;
  fs.writeFileSync(file,source.slice(0,start)+replacement+source.slice(end));
  console.log(JSON.stringify({file,line:+line,column:+column,order:+order,function:'<module initializer>',start,end,removed:source.slice(start,end),replacement,marker}));process.exit(0);
 }
 const replacement=`((): never => { throw ${JSON.stringify(marker)}; })()`;
 let prefix=source.slice(0,start);
 if(!declaration.type){
  const candidates=originals(n=>ts.isVariableDeclaration(n)&&n.name.getText(original)===declaration.name.getText(sf));
  if(candidates.length!==1)throw Error('ambiguous original variable type');
  const type=checker.typeToString(checker.getTypeAtLocation(candidates[0]),candidates[0],ts.TypeFormatFlags.NoTruncation);
  if(/\bany\b/.test(type))throw Error('original initializer has any; needs explicit sound placeholder');
  const insertion=declaration.name.end;prefix=source.slice(0,insertion)+': '+type+source.slice(insertion,start);
 }
 fs.writeFileSync(file,prefix+replacement+source.slice(end));
 console.log(JSON.stringify({file,line:+line,column:+column,order:+order,function:'<initializer>',start,end,removed:source.slice(start,end),replacement,marker}));process.exit(0);
}
let start=fn.body.getStart(sf),end=fn.body.end;
const marker=`binder discovery stop ${order}`;
let replacement=`{ throw ${JSON.stringify(marker)}; }`;
if(!ts.isBlock(fn.body))replacement=`((): never => ${replacement})()`;
if(source.slice(start,end).includes('binder discovery stop')){
 // A signature-only stop in an already throwing function can safely promise never.
 if(fn.type&&fn.type.getText(sf)!=='never'){start=fn.type.getStart(sf);end=fn.type.end;replacement='never';}
 else throw Error('already stubbed without a return annotation');
}
let prefix=source.slice(0,start);
if(!fn.type&&fn.name){
 const candidates=originals(n=>ts.isFunctionLike(n)&&n.name?.getText(original)===fn.name.getText(sf));
 if(candidates.length!==1)throw Error('ambiguous original function result type');
 const signature=checker.getSignatureFromDeclaration(candidates[0]);
 const type=checker.typeToString(checker.getReturnTypeOfSignature(signature),candidates[0],ts.TypeFormatFlags.NoTruncation);
 if(/\bany\b/.test(type))throw Error('original inferred result has any; needs explicit sound placeholder');
 prefix=source.slice(0,fn.parameters.end+1)+': '+type+source.slice(fn.parameters.end+1,start);
}
const next=prefix+replacement+source.slice(end);fs.writeFileSync(file,next);
console.log(JSON.stringify({file,line:+line,column:+column,order:+order,function:fn.name?.getText(sf)||'<callback>',start,end,removed:source.slice(start,end),replacement,marker}));
