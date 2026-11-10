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

const draft=JSON.parse(fs.readFileSync('/workspace/scratch/abrq61p/classification-draft.json','utf8'));
const nodes=new Map();for(const file of files.filter(f=>f.includes('/src/compiler/'))){const sf=program.getSourceFile(file);function visit(n){if(ts.isFunctionLike(n)&&n.body)nodes.set(where(n),n);ts.forEachChild(n,visit);}visit(sf);}
const result=[];
for(const row of draft.filter(r=>r.classification?.startsWith('unreached'))){for(const c of row.direct_or_alias_calls.filter(c=>c.external_callback===false)){
 const declaration=nodes.get(c.callback_callee),index=Number(c.via[0].match(/\d+/)[0]),parameter=declaration?.parameters[index];
 if(!parameter||!ts.isIdentifier(parameter.name))throw Error('missing callback binder '+JSON.stringify(c));
 const uses=refsToVariable(parameter).map(n=>{let exp=n;while(transparent(exp.parent,exp))exp=exp.parent;const p=exp.parent;return {where:where(n),call_site:p&&ts.isCallExpression(p)?where(p):null,direct_call:!!p&&ts.isCallExpression(p)&&p.expression===exp,forwarded_argument:!!p&&ts.isCallExpression(p)&&p.arguments.includes(exp)};});
 if(!uses.some(u=>u.direct_call||u.forwarded_argument))throw Error('callback not invoked or forwarded '+c.callback_callee);
 result.push({target:row.where,passing_call:c.call_site,callee:c.callback_callee,parameter:where(parameter),parameter_name:parameter.name.text,uses});
}}
fs.writeFileSync('/workspace/scratch/abrq61p/callback-bindings.json',JSON.stringify(result,null,2)+'\n');console.log(result.length+' internal callback bindings verified by checker references');service.dispose();
