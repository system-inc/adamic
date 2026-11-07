'use strict';
const fs=require('fs'),ts=require('/workspace/adamic/stage3/api/node_modules/typescript'),assert=require('assert/strict'),root='/workspace/scratch/levenshtein';
const pin=JSON.parse(fs.readFileSync(root+'/measured-owners.json','utf8')),old=fs.readFileSync(root+'/before/src/compiler/core.ts','utf8').replaceAll('\r\n','\n'),text=old.replace(pin.before.replaceAll('\r\n','\n'),pin.after.replaceAll('\r\n','\n')),s=ts.createSourceFile('core.ts',text,ts.ScriptTarget.Latest,true),rows=[];
const roles=['case-insensitive equal substitution','case-insensitive unequal substitution','equal code-unit branch','delete candidate','insert candidate','final result'];
function visit(n){if(ts.isNonNullExpression(n)&&ts.isElementAccessExpression(n.expression)&&['previous','current'].includes(n.expression.expression.getText(s))){const p=s.getLineAndCharacterOfPosition(n.getStart(s));rows.push({site:roles[rows.length],where:'core.ts:'+(p.line+1)+':'+(p.character+1),expression:n.getText(s),index:n.expression.argumentExpression.getText(s)});}ts.forEachChild(n,visit);}visit(s);assert.equal(rows.length,6);
console.log(JSON.stringify(rows,null,2));
