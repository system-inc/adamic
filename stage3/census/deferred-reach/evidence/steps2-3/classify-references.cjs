const fs=require('fs'),path=require('path'),ts=require('/workspace/adamic/stage3/api/node_modules/typescript');
const root='/workspace/scratch/abrq61p-adapted',files=[];
function walk(d){for(const e of fs.readdirSync(d,{withFileTypes:true})){const p=path.join(d,e.name);if(e.isDirectory())walk(p);else if(/\.tsx?$/.test(p))files.push(p);}}walk(root+'/src');files.sort();
const options={target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.NodeNext,moduleResolution:ts.ModuleResolutionKind.NodeNext,skipLibCheck:true};
const host={getScriptFileNames:()=>files,getScriptVersion:()=> '0',getScriptSnapshot:f=>fs.existsSync(f)?ts.ScriptSnapshot.fromString(fs.readFileSync(f,'utf8')):undefined,getCurrentDirectory:()=>root,getCompilationSettings:()=>options,getDefaultLibFileName:o=>ts.getDefaultLibFilePath(o),fileExists:ts.sys.fileExists,readFile:ts.sys.readFile,readDirectory:ts.sys.readDirectory,directoryExists:ts.sys.directoryExists,getDirectories:ts.sys.getDirectories};
const service=ts.createLanguageService(host),program=service.getProgram(),checker=program.getTypeChecker();
function where(n){const sf=n.getSourceFile(),pos=n.getStart(sf),loc=sf.getLineAndCharacterOfPosition(pos);return path.relative(root,sf.fileName)+':'+(loc.line+1)+':'+(loc.character+1);}
function owner(n){let top=n,fn;const functions=[];for(let a=n.parent;a&&!ts.isSourceFile(a);a=a.parent){top=a;if(ts.isFunctionLike(a)){fn ||= a;functions.push({where:where(a),kind:ts.SyntaxKind[a.kind],type_parameters:a.typeParameters?.length||0,has_body:!!a.body});}}return {unit:fn?where(fn):where(top),top_unit:where(top),top_kind:ts.SyntaxKind[top.kind],caller_name:fn?(fn.name?.getText()||'<'+ts.SyntaxKind[fn.kind]+'>'):'<top-level>',caller_kind:fn?ts.SyntaxKind[fn.kind]:ts.SyntaxKind[top.kind],functions};}

const variableRefs=new Map();
function refsToVariable(decl){const key=where(decl);if(variableRefs.has(key))return variableRefs.get(key);const out=[];for(const g of service.findReferences(decl.getSourceFile().fileName,decl.name.getStart())||[])for(const r of g.references){if(r.isDefinition)continue;const sf=program.getSourceFile(r.fileName);if(sf&&r.fileName.startsWith(root+'/src/'))out.push(ts.getTokenAtPosition(sf,r.textSpan.start));}variableRefs.set(key,out);return out;}
function transparent(p,n){return p&&p.expression===n&&(ts.isParenthesizedExpression(p)||ts.isAsExpression(p)||ts.isTypeAssertionExpression(p)||ts.isNonNullExpression(p)||ts.isSatisfiesExpression(p));}
function follow(start,via,seen,out){let n=start;while(transparent(n.parent,n))n=n.parent;
 while(n.parent&&(ts.isPropertyAccessExpression(n.parent)||ts.isElementAccessExpression(n.parent))&&n.parent.expression===n){n=n.parent;while(transparent(n.parent,n))n=n.parent;}
 if(n.parent&&ts.isCallExpression(n.parent)&&n.parent.expression===n){out.push({...owner(n.parent),call_site:where(n.parent),via});return;}
 let decl;
 if(n.parent&&ts.isVariableDeclaration(n.parent)&&n.parent.initializer===n)decl=n.parent;
 if(n.parent&&ts.isPropertyAssignment(n.parent)&&n.parent.initializer===n){let p=n.parent.parent;while(p&&!ts.isSourceFile(p)&&!ts.isVariableDeclaration(p)&&!ts.isFunctionLike(p))p=p.parent;if(p&&ts.isVariableDeclaration(p))decl=p;}
 if(!decl||!ts.isIdentifier(decl.name))return;const key=where(decl);if(seen.has(key))return;const next=new Set(seen);next.add(key);for(const ref of refsToVariable(decl))follow(ref,[...via,decl.name.text],next,out);
}
const result=[];
for(const file of files.filter(p=>p.startsWith(root+'/src/compiler/'))){const sf=program.getSourceFile(file);function visit(n){if(ts.isFunctionDeclaration(n)&&n.name&&n.typeParameters?.length){const symbol=checker.getSymbolAtLocation(n.name);const implementation=symbol?.declarations?.find(d=>ts.isFunctionDeclaration(d)&&d.body);const row={where:where(n),name:n.name.text,has_body:!!n.body,implementation:implementation?where(implementation):null,implementation_generic:!!implementation?.typeParameters?.length,references:[],calls:[]};
 for(const g of service.findReferences(file,n.name.getStart(sf))||[])for(const r of g.references){if(r.isDefinition)continue;const source=program.getSourceFile(r.fileName);if(!source||!r.fileName.startsWith(root+'/src/'))continue;const token=ts.getTokenAtPosition(source,r.textSpan.start);row.references.push({where:where(token),...owner(token)});
 let exp=token;while(transparent(exp.parent,exp))exp=exp.parent;
 if(exp.parent&&ts.isPropertyAccessExpression(exp.parent)&&exp.parent.name===exp)exp=exp.parent;
 else if(exp.parent&&ts.isElementAccessExpression(exp.parent)&&exp.parent.argumentExpression===exp&&ts.isStringLiteralLike(exp))exp=exp.parent;
 while(transparent(exp.parent,exp))exp=exp.parent;
 if(exp.parent&&ts.isCallExpression(exp.parent)&&exp.parent.expression===exp)row.calls.push({...owner(exp.parent),call_site:where(exp.parent),via:[]});
 else follow(token,[],new Set(),row.calls);
 if(exp.parent&&ts.isCallExpression(exp.parent)&&exp.parent.arguments.includes(exp)){
 const call=exp.parent,decl=checker.getResolvedSignature(call)?.declaration;
 const symbol=decl?.name&&checker.getSymbolAtLocation(decl.name);
 const implementation=decl?.body?decl:symbol?.declarations?.find(d=>d.body);
 const internal=!!implementation&&implementation.getSourceFile().fileName.startsWith(root+'/src/compiler/');
 row.calls.push({...owner(call),call_site:where(call),via:['callback argument '+call.arguments.indexOf(exp)],external_callback:!internal,callback_callee:implementation?where(implementation):decl?where(decl):null});
 }

 }
 const dedup=new Map();for(const c of row.calls)dedup.set(JSON.stringify(c),c);row.calls=[...dedup.values()].sort((a,b)=>JSON.stringify(a).localeCompare(JSON.stringify(b)));row.references.sort((a,b)=>a.where.localeCompare(b.where));result.push(row);console.log(row.where,row.references.length,row.calls.length);
 }ts.forEachChild(n,visit);}visit(sf);}
fs.writeFileSync('/workspace/scratch/abrq61p/classification-references.json',JSON.stringify({typescript:ts.version,source_files:files.length,units:result},null,2)+'\n');service.dispose();
