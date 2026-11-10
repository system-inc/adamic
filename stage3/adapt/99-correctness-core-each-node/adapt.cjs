'use strict';
// Task 8n7eet3, site 2, Adamic Nov 10 22:01Z. Only the ruled loop and unused import removal.
const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),crypto=require('node:crypto');
const ts=require('typescript');
const pins=require('./pins.json');
const hash=text=>crypto.createHash('sha256').update(text).digest('hex');
function inspect(text){
 assert.equal(ts.version,'6.0.3');assert.equal(text.split('\r\n').length,text.split('\n').length,'debug.ts must remain CRLF');
 const source=ts.createSourceFile('debug.ts',text,ts.ScriptTarget.Latest,true);assert.equal(source.parseDiagnostics.length,0,'debug.ts parse error');
 const spaces=source.statements.filter(n=>ts.isModuleDeclaration(n)&&n.name.text==='Debug');assert.equal(spaces.length,1,'namespace Debug changed');
 const owners=spaces[0].body.statements.filter(n=>ts.isFunctionDeclaration(n)&&n.name?.text==='assertEachNode'&&n.body);assert.equal(owners.length,1,'assertEachNode implementation changed');const owner=owners[0];
 const actual=owner.getText(source),state=actual===pins.before?'before':actual===pins.after?'after':undefined;assert(state,'unreviewed assertEachNode body');assert.equal(hash(actual),pins[state+'Sha256']);
 const imports=source.statements.filter(n=>ts.isImportDeclaration(n)&&n.moduleSpecifier.text==='./_namespaces/ts.js'&&n.importClause?.namedBindings&&ts.isNamedImports(n.importClause.namedBindings));assert.equal(imports.length,1,'namespace named import changed');
 const specifiers=imports[0].importClause.namedBindings.elements.filter(n=>n.name.text==='every');assert.equal(specifiers.length,state==='before'?1:0,'half-adapted every import');
 const references=[];function visit(n){if(ts.isIdentifier(n)&&n.text==='every')references.push(n);ts.forEachChild(n,visit);}visit(source);
 if(state==='before'){
  const specifier=specifiers[0];assert(!specifier.isTypeOnly&&!specifier.propertyName,'unreviewed every import');assert.equal(text.slice(specifier.getFullStart(),specifier.end+1),pins.importBefore,'unreviewed every import text');
  const uses=references.filter(n=>n!==specifier.name);assert.equal(uses.length,1,'every has another use');assert(uses[0].getStart(source)>=owner.getStart(source)&&uses[0].end<=owner.end,'every used outside assertEachNode');
 }else assert.equal(references.length,0,'every still referenced');
 return {source,owner,state,specifier:specifiers[0]};
}
function plan(text){
 const found=inspect(text);if(found.state==='after')return {after:text,edits:0};
 const {owner,specifier,source}=found;const edits=[{start:owner.getStart(source),end:owner.end,text:pins.after},{start:specifier.getFullStart(),end:specifier.end+1,text:''}];let after=text;
 for(const edit of edits.sort((a,b)=>b.start-a.start))after=after.slice(0,edit.start)+edit.text+after.slice(edit.end);
 assert.equal(after,text.replace(pins.before,pins.after).replace(pins.importBefore,''),'only the whole owner and import may change');assert.equal(inspect(after).state,'after');
 for(const stripInternal of [false,true]){const emit=t=>ts.transpileDeclaration(t,{fileName:'debug.ts',compilerOptions:{stripInternal}}).outputText;assert.equal(emit(after),emit(text),'declarations changed');}
 return {after,edits:2};
}
function adapt(tree){const file=path.join(tree,'src/compiler/debug.ts'),before=fs.readFileSync(file,'utf8'),result=plan(before);assert.equal(fs.readFileSync(file,'utf8'),before,'concurrent edit');if(result.edits)fs.writeFileSync(file,result.after);return {files:result.edits?1:0,edits:result.edits,javascriptChanging:true};}
module.exports={inspect,plan,adapt,pins};
if(require.main===module){assert.equal(process.argv.length,3,'usage: node adapt.cjs TREE');console.log(JSON.stringify(adapt(path.resolve(process.argv[2])),null,2));}
