'use strict';
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto'),ts=require('typescript');
const pins=require('./owners.json'),amendment=require('./import-amendment.json'),[before,after,out]=process.argv.slice(2);assert(before&&after&&out,'usage: node compare.cjs BEFORE AFTER OUTPUT');fs.mkdirSync(out,{recursive:true});
const hash=s=>crypto.createHash('sha256').update(s).digest('hex'),normalize=s=>s.replace(/\r\n/g,'\n');
function files(dir){return fs.readdirSync(dir,{recursive:true}).filter(f=>fs.statSync(path.join(dir,f)).isFile()).sort();}
const sourceNames=files(before+'/src');assert.deepEqual(sourceNames,files(after+'/src'));const changed=[];
for(const name of sourceNames){const file='src/'+name,a=fs.readFileSync(before+'/'+file,'utf8'),b=fs.readFileSync(after+'/'+file,'utf8');let expected=a;
 const replace=(old,next)=>{const eol=expected.includes('\r\n')?'\r\n':'\n';old=old.replace(/\n/g,eol);next=next.replace(/\n/g,eol);assert.equal(expected.split(old).length-1,1,file+': exact source span');expected=expected.replace(old,next);};
 if(file===amendment.file)replace(amendment.oldAfter,amendment.newAfter);
 for(const pin of pins.filter(p=>p.file===file))replace(pin.before,pin.after);
 assert.equal(b,expected,file+': source changed outside owned spans');if(a!==b)changed.push(file);
}
assert.deepEqual(changed.sort(),[...new Set(pins.map(p=>p.file))].sort());
const printer=ts.createPrinter({removeComments:true,newLine:ts.NewLineKind.LineFeed});
function parse(text){const source=ts.createSourceFile('bundle.js',text,ts.ScriptTarget.Latest,true,ts.ScriptKind.JS);assert.equal(source.parseDiagnostics.length,0);return source;}
function printed(source){const tokens=[];function visit(n){if(ts.isParenthesizedExpression(n)){visit(n.expression);return;}tokens.push(n.kind,ts.isVariableDeclarationList(n)?n.flags&(ts.NodeFlags.Let|ts.NodeFlags.Const|ts.NodeFlags.Using|ts.NodeFlags.AwaitUsing):0);if(ts.isIdentifier(n)||ts.isPrivateIdentifier(n)||ts.isLiteralExpression(n)||[ts.SyntaxKind.TemplateHead,ts.SyntaxKind.TemplateMiddle,ts.SyntaxKind.TemplateTail].includes(n.kind))tokens.push(n.text);ts.forEachChild(n,visit);tokens.push(-1);}visit(source);return hash(JSON.stringify(tokens));}
function reverted(text){const source=parse(text),counts={helpers:0,properties:0,deadExportAnnotations:0,calls:0};
 const transformed=ts.transform(source,[context=>{const visit=node=>{
  if(ts.isFunctionDeclaration(node)&&/^orPredicate[23]$/.test(node.name?.text||'')){const arity=+node.name.text.at(-1),expected=parse('function '+node.name.text+'('+Array.from({length:arity},(_,i)=>'f'+(i+1)).join(',')+'){return p => '+Array.from({length:arity},(_,i)=>'f'+(i+1)+'(p)').join(' || ')+';}');assert.equal(printed(node),printed(expected.statements[0]),'helper body differs from the required disjunction');counts.helpers++;return undefined;}
  if(ts.isPropertyAssignment(node)&&/^orPredicate[23]$/.test(node.name.getText(source))){const init=node.initializer;assert(ts.isArrowFunction(init)&&ts.isIdentifier(init.body)&&init.body.text===node.name.getText(source));counts.properties++;return undefined;}
  if(ts.isShorthandPropertyAssignment(node)&&/^orPredicate[23]$/.test(node.name.text)){const assignment=node.parent.parent,outer=assignment.parent.parent;assert(ts.isBinaryExpression(assignment)&&assignment.operatorToken.kind===ts.SyntaxKind.EqualsToken&&assignment.left.getText(source)==='module.exports');assert(ts.isBinaryExpression(outer)&&outer.operatorToken.kind===ts.SyntaxKind.AmpersandAmpersandToken&&ts.isNumericLiteral(outer.left)&&outer.left.text==='0','only the generated dead ESM export annotation');counts.deadExportAnnotations++;return undefined;}
  if(ts.isCallExpression(node)&&ts.isIdentifier(node.expression)&&/^orPredicate[23]$/.test(node.expression.text)){counts.calls++;return ts.factory.updateCallExpression(node,ts.factory.createIdentifier('or'),node.typeArguments,ts.visitNodes(node.arguments,visit));}
  return ts.visitEachChild(node,visit,context);
 };return n=>ts.visitNode(n,visit);}]);
 const result=printed(transformed.transformed[0]);transformed.dispose();return {text:result,counts};
}
const declarations=[];
for(const name of ['typescript.d.ts','tsserverlibrary.d.ts']){
 const a=fs.readFileSync(before+'/built/local/'+name),b=fs.readFileSync(after+'/built/local/'+name);assert(a.equals(b),'STOP: public declaration moved: '+name);declarations.push({name,byteIdentical:true,sha256:hash(a),orLines:a.toString().split('\n').filter(l=>/\bor\s*<|\borPredicate[23]\b/.test(l))});
}
const internalName='typescript.internal.d.ts',internalBefore=fs.readFileSync(before+'/built/local/'+internalName,'utf8'),internalAfter=fs.readFileSync(after+'/built/local/'+internalName,'utf8');
function matchingLines(text){return text.split('\n').flatMap((line,i)=>/\bfunction (or|orPredicate[23])</.test(line)?[{line:i+1,text:line}]:[]);}
const internal={byteIdentical:internalBefore===internalAfter,before:matchingLines(internalBefore),after:matchingLines(internalAfter)};
fs.writeFileSync(out+'/internal-lines.json',JSON.stringify(internal,null,2)+'\n');
const jsNames=files(before+'/built/local').filter(n=>/\.(?:c|m)?js$/.test(n));assert.deepEqual(jsNames,files(after+'/built/local').filter(n=>/\.(?:c|m)?js$/.test(n)));const javascript=[];
for(const name of jsNames){const a=fs.readFileSync(before+'/built/local/'+name,'utf8'),b=fs.readFileSync(after+'/built/local/'+name,'utf8'),same=a===b;let proof;
 if(!same){proof=reverted(b);assert.equal(proof.text,printed(parse(a)),'unmodeled JavaScript change: '+name);assert.equal(proof.counts.helpers,2,name+': helper definitions');assert(proof.counts.calls===7||proof.counts.calls===19,name+': rewritten calls');}
 javascript.push({name,byteIdentical:same,beforeSha256:hash(a),afterSha256:hash(b),proof:proof?.counts});
}
const a=require(path.resolve(before+'/built/local/typescript.js')),b=require(path.resolve(after+'/built/local/typescript.js')),keysBefore=Object.keys(a).sort(),keysAfter=Object.keys(b).sort();assert.deepEqual(keysAfter.filter(k=>!keysBefore.includes(k)),['orPredicate2','orPredicate3']);assert.deepEqual(keysBefore.filter(k=>!keysAfter.includes(k)),[]);
const summary={sourceFiles:sourceNames.length,changedSourceFiles:changed,onlyPinnedSourceSpansChanged:true,publicDeclarations:declarations,internalDeclarations:internal,javascript,runtimeObjectAddedProperties:['orPredicate2','orPredicate3'],runtimeObjectRemovedProperties:[]};fs.writeFileSync(out+'/comparison.json',JSON.stringify(summary,null,2)+'\n');console.log(JSON.stringify({sourceFiles:sourceNames.length,changedSources:changed.length,publicDeclarationsIdentical:true,javascriptFiles:javascript.length,changedJavascript:javascript.filter(r=>!r.byteIdentical).map(r=>r.name),addedRuntimeProperties:summary.runtimeObjectAddedProperties}));
