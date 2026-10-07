#!/usr/bin/env node
'use strict';
// Recheck the saved write inventory with the stock checker. No source is rewritten.
const fs = require('node:fs');
const zlib = require('node:zlib');
const path = require('node:path');
const assert = require('node:assert/strict');
const ts = require('typescript');
assert.equal(ts.version, '6.0.3');
const [tree, input, output] = process.argv.slice(2);
const report = JSON.parse(zlib.gunzipSync(fs.readFileSync(input)));
const configFile = path.join(tree, 'src/compiler/tsconfig.json');
const config = ts.readConfigFile(configFile, ts.sys.readFile);
const parsed = ts.parseJsonConfigFileContent(config.config, ts.sys, path.dirname(configFile), {}, configFile);
const program = ts.createProgram(parsed.fileNames, {...parsed.options, strict:true, exactOptionalPropertyTypes:true, noUncheckedIndexedAccess:true, verbatimModuleSyntax:true, erasableSyntaxOnly:true});
const checker = program.getTypeChecker();
function unwrap(n) { while (ts.isParenthesizedExpression(n) || ts.isAsExpression(n) || ts.isNonNullExpression(n)) n = n.expression; return n; }
function symbol(n) { return checker.getSymbolAtLocation(n); }
function same(a,b) { a=unwrap(a); b=unwrap(b); return ts.isIdentifier(a) && ts.isIdentifier(b) && symbol(a) === symbol(b); }
function location(n) { const f=n.getSourceFile(), p=f.getLineAndCharacterOfPosition(n.getStart()); return `${path.relative(tree,f.fileName)}:${p.line+1}:${p.character+1}`; }
function find(r) {
 const f=program.getSourceFile(path.join(tree,'src/compiler',r.File)); assert(f);
 const pos=f.getPositionOfLineAndCharacter(r.Line-1,r.Column-1); let found;
 function visit(n) { if(n.getStart(f)===pos && n.getText(f)===r.Text && [ts.SyntaxKind[n.kind],'Kind'+ts.SyntaxKind[n.kind]].includes(r.Kind)) found=n; if(n.pos<=pos && pos<n.end) ts.forEachChild(n,visit); }
 visit(f); assert(found,`moved site ${r.File}:${r.Line}`); return found;
}
// Accept only a bound flags member, or an immutable local initialized from it.
function flagReceiver(n) {
 n=unwrap(n);
 if(ts.isPropertyAccessExpression(n) && ['flags','objectFlags'].includes(n.name.text)) return unwrap(n.expression);
 if(ts.isIdentifier(n)) for(const d of symbol(n)?.declarations || []) {
  if(ts.isVariableDeclaration(d) && d.initializer && (d.parent.flags & ts.NodeFlags.Const)) {
   const init=unwrap(d.initializer);
   if(ts.isPropertyAccessExpression(init) && ['flags','objectFlags'].includes(init.name.text)) return unwrap(init.expression);
  }
 }
}
function enumMember(n,name,enumName='TypeFlags') {
 n=unwrap(n); if(!ts.isPropertyAccessExpression(n)) return false;
 const s=symbol(n.name);
 return s?.declarations?.some(d=>ts.isEnumMember(d) && d.name.getText()===name && ts.isEnumDeclaration(d.parent) && d.parent.name.text===enumName);
}
function guarantees(condition,value,tag,enumName='TypeFlags') {
 condition=unwrap(condition);
 if(ts.isBinaryExpression(condition)) {
  if(condition.operatorToken.kind===ts.SyntaxKind.AmpersandAmpersandToken) return guarantees(condition.left,value,tag,enumName) || guarantees(condition.right,value,tag,enumName);
  if(condition.operatorToken.kind===ts.SyntaxKind.AmpersandToken) {
   const receiver=flagReceiver(condition.left);
   return receiver && same(receiver,value) && enumMember(condition.right,tag,enumName);
  }
 }
 return false;
}
function guards(n,value,tag,enumName='TypeFlags') {
 const result=[];
 for(let child=n,parent=n.parent;parent;child=parent,parent=parent.parent) {
  let condition;
  if(ts.isIfStatement(parent) && child===parent.thenStatement) condition=parent.expression;
  if(ts.isConditionalExpression(parent) && child===parent.whenTrue) condition=parent.condition;
  if(ts.isBinaryExpression(parent) && parent.operatorToken.kind===ts.SyntaxKind.AmpersandAmpersandToken && child===parent.right) condition=parent.left;
  if(condition && guarantees(condition,value,tag,enumName)) result.push({where:location(condition),condition:condition.getText(),receiver_symbol:symbol(unwrap(value))?.getName(),tag,enum:enumName});
 }
 return result;
}
function freshProof(n) {
 // The ruling expressly sanctions the allocation arm of (v || (v = {}))[k] = x.
 // Only that immediate initialization is fresh; the old-v arm and subsequent
 // writes have a target-typed source and are not this literal's widening site.
 if(!ts.isBinaryExpression(n) || n.operatorToken.kind!==ts.SyntaxKind.EqualsToken || !ts.isIdentifier(n.left) || !ts.isObjectLiteralExpression(n.right) || n.right.properties.length) return;
 const assign=unwrap(n), disjunction=assign.parent;
 let p=disjunction; while(ts.isParenthesizedExpression(p)) p=p.parent;
 if(!ts.isBinaryExpression(p) || p.operatorToken.kind!==ts.SyntaxKind.BarBarToken || unwrap(p.right)!==assign || !same(p.left,n.left)) return;
 let receiver=p.parent; while(ts.isParenthesizedExpression(receiver)) receiver=receiver.parent;
 if(!ts.isElementAccessExpression(receiver) || unwrap(receiver.expression)!==p) return;
 const write=receiver.parent;
 if(!ts.isBinaryExpression(write) || write.left!==receiver || write.operatorToken.kind!==ts.SyntaxKind.EqualsToken) return;
 let escapes=false;
 function inspect(x) { if(ts.isIdentifier(x) && symbol(x)===symbol(n.left)) escapes=true; ts.forEachChild(x,inspect); }
 inspect(write.right); inspect(receiver.argumentExpression);
 if(escapes) return;
 const declarations=symbol(n.left)?.declarations || [];
 assert(declarations.some(d=>ts.isVariableDeclaration(d) || ts.isParameter(d)));
 return {allocation:location(n.right),owning_binding:declarations.map(location),immediate_write:location(write),expression:write.getText(),escape_analysis:'Empty literal has no initializer callbacks, spreads or references. Its only initial binding is the wider-view receiver binding; the same evaluation immediately writes through that receiver. No other alias to this allocation can be stored, passed or captured before this initial write. Freshness ends at that binding/initialization; no later write through an escaped value is licensed. This narrowly follows the expressly sanctioned sys.ts:155 construction pattern, not a general exemption for stored literals.'};
}
const results=[];
for(const r of report.rows.filter(r=>r.classification==='write')) {
 const n=find(r), value=unwrap(n), stock=checker.getTypeAtLocation(ts.isAsExpression(n)?n.expression:n);
 let bucket='rest', proof={reason:'No fresh literal or dominating tag check was proved for this relation.'};
 const fresh=freshProof(n);
 if(fresh) {bucket='a';proof=fresh;}
 else if(r.Source==='never') {bucket='b';proof={adamic_source:r.Source,stock_source:checker.typeToString(stock),stock_never:!!(stock.flags & ts.TypeFlags.Never),reason:'The refusal census records the source of this relation as never. Case (b) applies to that Adamic static relation. This is not independent evidence that the original stock-TypeScript expression is unreachable; inhabited stock sources are flagged for compiler reconciliation.'};}
 else if(['TypeParameter','UnionType','AnonymousType'].includes(r.Target) && ts.isIdentifier(value)) {
  let evidence=guards(n,value,r.Target==='TypeParameter'?'TypeParameter':r.Target==='UnionType'?'Union':'Object');
  if(r.Target==='AnonymousType') { const objectGuards=guards(n,value,'Anonymous','ObjectFlags'); evidence=evidence.length && objectGuards.length ? [...evidence,...objectGuards] : []; }
  if(evidence.length) {bucket='c';proof={guards:evidence,target_declarations:r.target_property_declarations,reason:'The positive tag tests dominate this view and resolve to the same checker-bound source value. The proven runtime class declares the member; these are guarded downcasts under case (c).'};}
 }
 results.push({...r,bucket,proof});
}
assert.equal(results.length,100);
assert.equal(new Set(results.map(r=>`${r.File}:${r.Line}:${r.Column}:${r.Kind}`)).size,100);
const counts=Object.fromEntries(['a','b','c','rest'].map(b=>[b,results.filter(r=>r.bucket===b).length]));
const out={typescript:ts.version,ruling:'October 7 05:30',counts,never_stock_disagreements:results.filter(r=>r.bucket==='b'&&!r.proof.stock_never).length,rows:results};
fs.writeFileSync(output,JSON.stringify(out,null,2)+'\n');
console.log(JSON.stringify({counts,never_stock_disagreements:out.never_stock_disagreements}));
