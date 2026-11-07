'use strict';
const fs=require('node:fs'),cp=require('node:child_process'),ts=require('/workspace/adamic/stage3/api/node_modules/typescript');
const out='/workspace/scratch/elementat-declaration';
const core=cp.execFileSync('git',['show','ba90529b3:stage3/census/core-sites-af42d3dc/core.ts.txt'],{encoding:'utf8'}).replaceAll('\r\n','\n');
const sf=ts.createSourceFile('core.ts',core,ts.ScriptTarget.Latest,true);
const helper=sf.statements.find(n=>ts.isFunctionDeclaration(n)&&n.name?.text==='toOffset').getText(sf);
const pin=require('../../adapt/99-correctness-axolotl-z/owners.json')[0];
const owner=pin.after;
const body=ts.createSourceFile('owner.ts',owner,ts.ScriptTarget.Latest,true).statements[0].declarationList.declarations[0].initializer.body.getText();
const declaration='export function elementAt<T>(array: readonly T[] | undefined, offset: number): T | undefined '+body;
fs.writeFileSync(out+'/declaration.txt',declaration+'\n');
let rows='interface Item { id: number; }\nconst items: readonly Item[] = [{id: 10}, {id: 20}];\nconst numbers: readonly number[] = [10,20];\nconst strings: readonly string[] = ["ten","twenty"];\n';
let count=0;
for(const [type,arr,field] of [['Item','items','.id'],['number','numbers',''],['string','strings','']]) for(const undefinedArray of [false,true]) for(const offset of [-1,0,1,5,-5]){
 const n='r'+count++; const expected=undefinedArray||offset===5||offset===-5?'undefined':arr+'['+(offset===-1?1:offset)+']'; rows+=`const ${n} = elementAt<${type}>(${undefinedArray?'undefined':arr}, ${offset});\nconsole.log("${type}/${undefinedArray?'undefined':'array'}/${offset}"); if (${n} === undefined) { console.log("undefined"); } else { console.log("" + ${n}${field}); }\nconsole.log(${n} === ${expected} ? "same" : "different");\n`;
}
for(const [name,fn] of [['p1',owner],['p2',declaration]])fs.writeFileSync(out+'/'+name+'.ts',helper+'\n'+fn+'\n'+rows);
fs.writeFileSync(out+'/probe-manifest.json',JSON.stringify({rows:count,helper,owner,declaration},null,2)+'\n');
