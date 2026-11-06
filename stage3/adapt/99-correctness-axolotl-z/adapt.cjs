'use strict';
const fs=require('node:fs'),path=require('node:path'),crypto=require('node:crypto');
const pins=require('./owners.json');
const hash=s=>crypto.createHash('sha256').update(s).digest('hex');
function plan(text,ts){
 if(ts.version!=='6.0.3')throw Error('want stock TypeScript 6.0.3');
 if(pins.length!==1)throw Error('want one owner');const pin=pins[0];
 for(const side of ['before','after'])if(hash(pin[side])!==pin[side+'Sha256'])throw Error('corrupt owner pin');
 const source=ts.createSourceFile('core.ts',text,ts.ScriptTarget.Latest,true);if(source.parseDiagnostics.length)throw Error('cannot parse core.ts');
 const found=[];function visit(n){if(ts.isVariableDeclaration(n)&&ts.isIdentifier(n.name)&&n.name.text===pin.owner)found.push(n);ts.forEachChild(n,visit);}visit(source);
 if(found.length!==1)throw Error('missing or duplicate elementAt owner');
 const d=found[0],list=d.parent,stmt=list.parent;
 if(!ts.isVariableDeclarationList(list)||!ts.isVariableStatement(stmt)||stmt.parent!==source||!(list.flags&ts.NodeFlags.Const)||!stmt.modifiers?.some(m=>m.kind===ts.SyntaxKind.ExportKeyword)||list.declarations.length!==1)throw Error('elementAt is not the pinned top-level exported const');
 const current=stmt.getText(source).replaceAll('\r\n','\n');
 if(current===pin.after)return {text,edits:0};
 if(current!==pin.before)throw Error('changed or half-adapted elementAt owner');
 const newline=text.includes('\r\n')?'\r\n':'\n';
 const next=text.slice(0,stmt.getStart(source))+pin.after.replaceAll('\n',newline)+text.slice(stmt.end);
 if(ts.createSourceFile('core.ts',next,ts.ScriptTarget.Latest,true).parseDiagnostics.length)throw Error('adapted source cannot parse');
 return {text:next,edits:1};
}
function main(){
 if(process.argv.length!==3)throw Error('usage: node adapt.cjs <tree>');
 const name=path.resolve(process.argv[2],'src/compiler/core.ts'),before=fs.readFileSync(name,'utf8'),result=plan(before,require('typescript'));
 if(fs.readFileSync(name,'utf8')!==before)throw Error('source changed before write');
 if(result.edits)fs.writeFileSync(name,result.text);
 console.log(JSON.stringify({files:result.edits?1:0,edits:result.edits,javascriptChanging:true}));
}
module.exports={plan};
if(require.main===module){try{main();}catch(error){console.error(error.message);process.exitCode=1;}}
