#!/usr/bin/env node
'use strict';
// Record enclosing conditions and typed function boundaries without claiming they prove an alias.
const fs = require('node:fs'), path = require('node:path'), zlib = require('node:zlib');
const ts = require('typescript'), assert = require('node:assert/strict');
const [tree, output] = process.argv.slice(2);
const inventory = JSON.parse(zlib.gunzipSync(fs.readFileSync(path.join(__dirname, '../evidence/ruling/analysis.json.gz'))));
const files = new Map();
function file(name) {
 if (!files.has(name)) {
  const f = ts.createSourceFile(name, fs.readFileSync(path.join(tree, name), 'utf8'), 99, true);
  const nodes = []; function visit(n) { nodes.push(n); ts.forEachChild(n, visit); } visit(f);
  files.set(name, {f, nodes});
 }
 return files.get(name);
}
function where(n, f) { const p = f.getLineAndCharacterOfPosition(n.getStart(f)); return `${f.fileName}:${p.line + 1}:${p.character + 1}`; }
function context(n, f) {
 const conditions = [], boundaries = [];
 for (let child = n, p = n.parent; p; child = p, p = p.parent) {
  if (ts.isIfStatement(p) && (child === p.thenStatement || child === p.elseStatement)) conditions.push({where:where(p.expression,f), branch:child === p.thenStatement ? 'true' : 'false', expression:p.expression.getText(f)});
  if (ts.isConditionalExpression(p) && (child === p.whenTrue || child === p.whenFalse)) conditions.push({where:where(p.condition,f), branch:child === p.whenTrue ? 'true' : 'false', expression:p.condition.getText(f)});
  if (ts.isBinaryExpression(p) && [ts.SyntaxKind.AmpersandAmpersandToken, ts.SyntaxKind.BarBarToken].includes(p.operatorToken.kind) && child === p.right) conditions.push({where:where(p.left,f), branch:p.operatorToken.kind === ts.SyntaxKind.AmpersandAmpersandToken ? 'true' : 'false', expression:p.left.getText(f)});
  if (ts.isFunctionDeclaration(p) || ts.isArrowFunction(p) || ts.isFunctionExpression(p)) boundaries.push({where:where(p,f), name:p.name?.getText(f), parameters:p.parameters.map(x => x.getText(f)), return_type:p.type?.getText(f)});
 }
 return {conditions, boundaries};
}
function locate(name, line, column, text, kind) {
 const {f, nodes} = file(name); const pos = f.getPositionOfLineAndCharacter(line - 1, column - 1);
 const candidates = nodes.filter(n => n.getStart(f) === pos && (!text || n.getText(f) === text) && (!kind || 'Kind' + ts.SyntaxKind[n.kind] === kind));
 if (text) assert.equal(candidates.length, 1, name + ':' + line + ':' + column);
 const n = text ? candidates[0] : candidates.find(ts.isBinaryExpression) || candidates.find(ts.isCallExpression) || candidates[0];
 assert(n); return {where:where(n,f), text:n.getText(f), ...context(n,f)};
}
const rows = inventory.rows.filter(r => ['b','rest'].includes(r.bucket)).map(r => ({id:`${r.File}:${r.Line}:${r.Column}`, site:locate('src/compiler/' + r.File,r.Line,r.Column,r.Text,r.Kind), writes:r.writes.map(w => {const [name,line,column]=w.where.split(':');return {property:w.property,...locate(name,+line,+column)};})}));
fs.writeFileSync(output, JSON.stringify(rows, null, 2) + '\n');
console.log(JSON.stringify({sites:rows.length, writes:rows.reduce((n,r)=>n+r.writes.length,0)}));
