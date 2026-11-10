'use strict';
const fs=require('node:fs'),path=require('node:path'),crypto=require('node:crypto'),assert=require('node:assert/strict');
const ts=require('/workspace/adamic/stage3/api/node_modules/typescript');
const file=process.argv[2],out=process.argv[3],leftPath='/workspace/scratch/debug-display-before/built/local/'+file,rightPath='/workspace/scratch/debug-display-candidate/built/local/'+file;
const left=fs.readFileSync(leftPath,'utf8'),right=fs.readFileSync(rightPath,'utf8'),names=['attachFlowNodeDebugInfoWorker','attachFlowNodeDebugInfo','attachNodeArrayDebugInfoWorker','attachNodeArrayDebugInfo','enableDebugInfo'];
const hash=x=>crypto.createHash('sha256').update(x).digest('hex');
function parse(text){return ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true,ts.ScriptKind.JS);}
function normalize(text){const sf=parse(text),edits=[];function visit(n){if(ts.isIdentifier(n)&&/[0-9]$/.test(n.text))edits.push({start:n.getStart(sf),end:n.end,text:n.text.replace(/[0-9]+$/,'')});ts.forEachChild(n,visit);}visit(sf);const parts=[];let cursor=0;for(const e of edits.sort((a,b)=>a.start-b.start)){parts.push(text.slice(cursor,e.start),e.text);cursor=e.end;}parts.push(text.slice(cursor));return parts.join('');}
function owners(sf){const found=new Map();function visit(n){if(ts.isFunctionDeclaration(n)&&n.name){const key=names.find(k=>new RegExp('^'+k+'[0-9]*$').test(n.name.text));if(key){assert(!found.has(key));found.set(key,n);}}ts.forEachChild(n,visit);}visit(sf);assert.equal(found.size,5);return found;}
const lsf=parse(left),rsf=parse(right),lo=owners(lsf),ro=owners(rsf);let residual=left;
for(const[name,n]of [...lo].sort((a,b)=>b[1].getStart(lsf)-a[1].getStart(lsf)))residual=residual.slice(0,n.getStart(lsf))+ro.get(name).getText(rsf)+residual.slice(n.end);
fs.writeFileSync(out+'.expected.js.txt',residual);fs.writeFileSync(out+'.expected.normalized.txt',normalize(residual));fs.writeFileSync(out+'.after.normalized.txt',normalize(right));
const expectedSf=parse(residual);
function declarations(sf){const rows=[];function visit(n){if(ts.isFunctionDeclaration(n)&&n.name||ts.isVariableStatement(n)){rows.push({node:n,text:normalize(n.getText(sf)),names:ts.isFunctionDeclaration(n)?[n.name.text]:n.declarationList.declarations.map(d=>d.name.getText(sf))});}ts.forEachChild(n,visit);}visit(sf);return rows;}
const afterCounts=new Map();for(const d of declarations(rsf))afterCounts.set(d.text,(afterCounts.get(d.text)||0)+1);
const removed=[];for(const d of declarations(expectedSf)){const count=afterCounts.get(d.text)||0;if(count)afterCounts.set(d.text,count-1);else removed.push(d);}
// This is a declaration inventory, not a textual occurrence scan. Resolve reference symbols in the original bundle.
const program=ts.createProgram([leftPath],{allowJs:true,noLib:true,noResolve:true,checkJs:false,target:ts.ScriptTarget.ES2024});const checker=program.getTypeChecker(),original=program.getSourceFile(leftPath),originalOwners=owners(original);
const deletionRows=[];
for(const d of removed){
 assert(ts.isFunctionDeclaration(d.node),'unreviewed removed binding shape: '+d.names.join(','));const canonical=d.names[0].replace(/[0-9]+$/,'');const candidates=[];
 function find(n){if(ts.isFunctionDeclaration(n)&&n.name?.text.replace(/[0-9]+$/,'')===canonical&&normalize(n.getText(original))===d.text)candidates.push(n);ts.forEachChild(n,find);}find(original);assert.equal(candidates.length,1,'removed declaration identity '+canonical);
 const decl=candidates[0],symbol=checker.getSymbolAtLocation(decl.name);assert(symbol);const refs=[];
 function visit(n){if(ts.isIdentifier(n)&&n!==decl.name&&n.text===decl.name.text&&checker.getSymbolAtLocation(n)===symbol){const pos=original.getLineAndCharacterOfPosition(n.getStart(original));let enclosing=n.parent;while(enclosing&&!ts.isFunctionDeclaration(enclosing))enclosing=enclosing.parent;const owner=[...originalOwners].find(([k,o])=>n.getStart(original)>=o.getStart(original)&&n.end<=o.end);refs.push({line:pos.line+1,column:pos.character+1,text:n.getText(original),enclosing_function:enclosing?.name?.text||null,inside_five_owner:owner?.[0]||null});}ts.forEachChild(n,visit);}visit(original);
 deletionRows.push({name:canonical,kind:'function',start_line:expectedSf.getLineAndCharacterOfPosition(d.node.getStart(expectedSf)).line,end_line:expectedSf.getLineAndCharacterOfPosition(d.node.end).line,normalized_declaration_sha256:hash(d.text),original_declaration_line:original.getLineAndCharacterOfPosition(decl.getStart(original)).line+1,references:refs,every_reference_inside_five_owners:refs.length>0&&refs.every(r=>r.inside_five_owner!==null)});
}
fs.writeFileSync(out+'.json',JSON.stringify({file,before_sha256:hash(left),after_sha256:hash(right),expected_with_five_owners_replaced_sha256:hash(residual),normalization:'Only AST Identifier nodes lose a terminal decimal suffix; strings, comments, numeric literals, whitespace and punctuation are untouched.',removed:deletionRows},null,2)+'\n');
console.log(JSON.stringify({file,removed:deletionRows.map(r=>({name:r.name,direct_owner_references:r.every_reference_inside_five_owners,references:r.references}))}));
