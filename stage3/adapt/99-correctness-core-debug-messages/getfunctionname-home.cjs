'use strict';
const assert=require('node:assert/strict'),fs=require('node:fs'),path=require('node:path');
const pins=require('./getfunctionname-home.json');
const normal=text=>text.replace(/\r\n/g,'\n');
function inspect(ts,debugText,localText='',stage='finished'){
 function parse(name,text){const source=ts.createSourceFile(name,text,ts.ScriptTarget.Latest,true);assert.equal(source.parseDiagnostics.length,0,name+': parse error');let declarations=[];function walk(n){if(ts.isFunctionDeclaration(n)&&n.name?.text==='getFunctionName')declarations.push(n);ts.forEachChild(n,walk);}walk(source);return {source,declarations};}
 const debug=parse('debug.ts',debugText),local=parse('deprecate.ts',localText);assert.equal(debug.declarations.length+local.declarations.length,1,'getFunctionName: missing or duplicate homes');
 if(local.declarations.length){const node=local.declarations[0];assert.equal(node.parent,local.source,'getFunctionName: local must be module level');assert.equal(normal(node.getText(local.source)),pins.local,'getFunctionName: drifted or half-applied moved body');assert.equal(normal(node.body.getText(local.source)),pins.body,'getFunctionName: moved body drift');return {moved:true,node,source:local.source};}
 const node=debug.declarations[0];assert(ts.isModuleBlock(node.parent)&&node.parent.parent.name.text==='Debug','getFunctionName: original home changed');
 const variants=[pins.debug];if(stage==='30')variants.push(pins.debug.replace('return func.name;','return (func as any).name;'),pins.debug.replace('return func.name;','return (func as any).name;').replace('match[1]!','match[1]'));if(stage==='41')variants.push(pins.debug.replace('return func.name;','return (func as any).name;'));
 assert(variants.includes(normal(node.getText(debug.source))),'getFunctionName: drifted or half-applied debug body');return {moved:false,node,source:debug.source};
}
function read(ts,tree,debugText,stage){const file=path.join(tree,'src/deprecatedCompat/deprecate.ts');return inspect(ts,debugText,fs.existsSync(file)?fs.readFileSync(file,'utf8'):'',stage);}
module.exports={inspect,read,pins};
