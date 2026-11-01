#!/usr/bin/env node
'use strict';
// Proof inputs only. Never used by the adapter or independent verifier.
const fs=require('node:fs'), path=require('node:path'), assert=require('node:assert/strict'), ts=require('typescript');
const [tree,index,direction]=process.argv.slice(2), row=require('./sites.json')[Number(index)];
assert(row && ['forward','undo'].includes(direction));
const file=path.join(tree,row.file), text=fs.readFileSync(file,'utf8');
const source=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true), owners=[];
function visit(node){if((ts.isFunctionDeclaration(node)||ts.isInterfaceDeclaration(node))&&node.name?.text===row.owner)owners.push(node);ts.forEachChild(node,visit);}
visit(source);
assert.equal(owners.length,row.ownerCount);
const owner=owners[row.ownerOccurrence], before=direction==='forward'?row.before:row.after, after=direction==='forward'?row.after:row.before;
const declarations=[];
function select(node){if([row.kind,row.kindAfter].includes(ts.SyntaxKind[node.kind])&&node.name?.getText(source)===row.name)declarations.push(node);ts.forEachChild(node,select);}
select(owner);
assert.equal(declarations.length,1);
const declaration=declarations[0];
assert.equal(declaration.getText(source),before,'exact proof input declaration');
fs.writeFileSync(file,text.slice(0,declaration.getStart(source))+after+text.slice(declaration.end));
