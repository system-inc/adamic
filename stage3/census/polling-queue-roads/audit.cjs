'use strict';
const fs=require('fs'),ts=require('../../api/node_modules/typescript'),{roads}=require('./roads.cjs');
const sf=ts.createSourceFile('owner.ts',roads.original,ts.ScriptTarget.Latest,true);
const fields=[];function walk(n){if(ts.isPropertyAccessExpression(n)&&['pollingInterval','pollIndex','pollScheduled'].includes(n.name.text)){const p=n.parent;fields.push({line:sf.getLineAndCharacterOfPosition(n.getStart(sf)).line+1,text:n.getText(sf),use:ts.isBinaryExpression(p)&&p.left===n&&p.operatorToken.kind===ts.SyntaxKind.EqualsToken?'write':'read'});}ts.forEachChild(n,walk)}walk(sf);
fs.writeFileSync(__dirname+'/evidence/field-uses.json',JSON.stringify(fields,null,2)+'\n');console.log(fields.reduce((o,f)=>{const k=f.text.split('.').at(-1)+':'+f.use;o[k]=(o[k]||0)+1;return o},{}));
