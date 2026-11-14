#!/usr/bin/env node
// AST-selected edits; symbol-resolved parser uses; exact signature arrows.
const ts=require('../../api/node_modules/typescript'),fs=require('fs'),path=require('path');
const [rootArg,heatArg,out]=process.argv.slice(2),root=path.resolve(rootArg);
const heat=JSON.parse(fs.readFileSync(heatArg)),sites=heat.sites.filter(s=>s.public);
const cfg=ts.getParsedCommandLineOfConfigFile(root+'/src/compiler/tsconfig.json',{}, {...ts.sys,onUnRecoverableConfigFileDiagnostic:d=>{throw Error(ts.flattenDiagnosticMessageText(d.messageText,'\n'))}});
const roots=ts.sys.readDirectory(root+'/src/compiler',['.ts']),program=ts.createProgram(roots,cfg.options),checker=program.getTypeChecker(),printer=ts.createPrinter(),result=[];
function walk(n,fn){fn(n);ts.forEachChild(n,c=>walk(c,fn));}
function print(n,sf){return printer.printNode(ts.EmitHint.Unspecified,n,sf);}
function loc(n){const f=n.getSourceFile(),p=f.getLineAndCharacterOfPosition(n.getStart());return path.relative(root,f.fileName)+':'+(p.line+1)+':'+(p.character+1);}
const pending=[];
for(const site of sites){const sf=program.getSourceFile(root+'/'+site.file);let pattern;walk(sf,n=>{if(ts.isObjectBindingPattern(n)&&sf.getLineAndCharacterOfPosition(n.getStart()).line+1===site.line)pattern=n;});if(!pattern)throw Error('missing site '+site.file);
const decl=pattern.parent,stmt=decl.parent.parent,receiver=decl.initializer,edits=[],members=[],imports=new Set();if(!ts.isVariableStatement(stmt)||stmt.declarationList.declarations.length!==1)throw Error('unsupported statement');
let replacement='';
if(site.type==='NodeFactory'){
 for(const e of pattern.elements){const symbol=checker.getSymbolAtLocation(e.name),name=(e.propertyName||e.name).text;let calls=0;
 walk(sf,n=>{if(ts.isIdentifier(n)&&n!==e.name&&checker.getSymbolAtLocation(n)===symbol){if(!ts.isCallExpression(n.parent)||n.parent.expression!==n)throw Error('unsafe parser use '+loc(n));if(checker.resolveName(receiver.text,n,ts.SymbolFlags.Value,false)!==checker.getSymbolAtLocation(receiver))throw Error('shadowed factory '+loc(n));edits.push({start:n.getStart(),end:n.end,text:receiver.text+'.'+name});calls++;}});members.push({name,local:e.name.text,rewrittenUses:calls});}
 replacement='';
}else if(site.type==='TransformationContext'){
 const kept=[];const arrows=[];
 for(const e of pattern.elements){const name=(e.propertyName||e.name).text,member=checker.getTypeAtLocation(receiver).getProperty(name);if(!member?.declarations?.some(ts.isMethodSignature)){kept.push(e.getText(sf));continue;}
 const signatures=checker.getTypeOfSymbolAtLocation(member,e).getCallSignatures();if(signatures.length!==1)throw Error('overloads '+name);const signature=signatures[0],d=signature.declaration;
 if(d.typeParameters?.length||d.parameters.some(p=>p.dotDotDotToken||!ts.isIdentifier(p.name)))throw Error('non-exact signature '+name);
 const params=d.parameters.map(p=>print(p,d.getSourceFile())).join(', '),ret=print(d.type,d.getSourceFile());
 for(const n of [...d.parameters.map(p=>p.type),d.type])if(n)walk(n,x=>{if(ts.isIdentifier(x))imports.add(x.text);});
 arrows.push(`const ${e.name.text} = (${params}): ${ret} => ${receiver.text}.${name}(${d.parameters.map(p=>p.name.text).join(', ')});`);
 const symbol=checker.getSymbolAtLocation(e.name);let direct=0,value=0;walk(sf,n=>{if(ts.isIdentifier(n)&&n!==e.name&&checker.getSymbolAtLocation(n)===symbol){if(ts.isCallExpression(n.parent)&&n.parent.expression===n)direct++;else value++;}});
 members.push({name,local:e.name.text,parameters:params,returns:ret,directUses:direct,valueUses:value});}
 replacement=(kept.length?`const { ${kept.join(', ')} } = ${receiver.text};\n    `:'')+arrows.join('\n    ');
 const imp=sf.statements.find(s=>ts.isImportDeclaration(s)&&s.importClause?.namedBindings&&ts.isNamedImports(s.importClause.namedBindings)&&s.moduleSpecifier.text.endsWith('/_namespaces/ts.js'));
 if(!imp)throw Error('missing ts namespace import '+site.file);
 const existing=new Set(imp.importClause.namedBindings.elements.map(e=>e.name.text));const missing=[...imports].filter(n=>!existing.has(n)&&n!=='undefined');
 if(missing.length){const b=imp.importClause.namedBindings;edits.push({start:b.end-1,end:b.end-1,text:missing.map(n=>'    '+n+',\n').join('')});}
}else if(site.type==='PrintHandlers'){
 const view=pattern.elements.map(e=>{const name=(e.propertyName||e.name).text;return `readonly ${name}?: PrintHandlers["${name}"];`;}).join(' ');
 replacement=`const handlersView: { ${view} } = handlers;\n    `+pattern.elements.map(e=>{const name=(e.propertyName||e.name).text;members.push({name,local:e.name.text,default:e.initializer?.getText(sf)||null});return `const ${e.name.text} = handlersView.${name}${e.initializer?' ?? '+e.initializer.getText(sf):''};`;}).join('\n    ');
}else throw Error('unknown site');
edits.push({start:stmt.getStart(),end:stmt.end,text:replacement});let text=sf.text;for(const e of edits.sort((a,b)=>b.start-a.start))text=text.slice(0,e.start)+e.text+text.slice(e.end);
pending.push({file:sf.fileName,text});result.push({file:site.file,site:site.line,kind:site.type,members});}
for(const p of pending)fs.writeFileSync(p.file,p.text);
// Rebuild and verify each inserted context identifier resolves to its transformer's parameter.
const after=ts.createProgram(roots,cfg.options),ac=after.getTypeChecker();let arrows=0;
for(const row of result.filter(r=>r.kind==='TransformationContext')){const sf=after.getSourceFile(root+'/'+row.file);row.bindingChecks=[];
 walk(sf,n=>{if(ts.isVariableDeclaration(n)&&ts.isIdentifier(n.name)&&row.members.some(m=>m.local===n.name.text)&&n.initializer&&ts.isArrowFunction(n.initializer)&&ts.isCallExpression(n.initializer.body)&&ts.isPropertyAccessExpression(n.initializer.body.expression)){
 const receiver=n.initializer.body.expression.expression;if(!ts.isIdentifier(receiver)||receiver.text!=='context')return;
 const symbol=ac.getSymbolAtLocation(receiver),d=symbol?.valueDeclaration;if(!d||!ts.isParameter(d)||d.name.getText()!=='context')throw Error('wrong arrow context '+loc(n));
 let owner=n;while(owner&&!ts.isFunctionDeclaration(owner))owner=owner.parent;if(d.parent!==owner)throw Error('arrow outside transformer scope '+loc(n));
 row.bindingChecks.push({arrow:n.name.text,contextParameter:loc(d),verified:true});arrows++;
 }});if(row.bindingChecks.length!==row.members.length)throw Error('missing arrow binding check '+row.file);}
fs.writeFileSync(out,JSON.stringify({roots:roots.length,sites:result.length,parserUses:result[0].members.reduce((a,m)=>a+m.rewrittenUses,0),transformerArrows:arrows,files:result},null,2)+'\n');
