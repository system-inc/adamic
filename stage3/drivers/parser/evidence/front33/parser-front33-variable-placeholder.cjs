const fs=require('fs'),ts=require('/home/agent/.cache/adamic-stage3/api/node_modules/typescript');
const [file,line,col]=process.argv.slice(2),text=fs.readFileSync(file,'utf8'),sf=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true),pos=sf.getPositionOfLineAndCharacter(+line-1,+col-1);let variable;
function find(n){if(n.pos<=pos&&pos<n.end){if(ts.isVariableDeclaration(n))variable=n;ts.forEachChild(n,find);}}find(sf);
if(!variable?.type)throw Error('expected typed discovery initializer');
const statement=variable.parent.parent;if(!ts.isVariableStatement(statement)||statement.declarationList.declarations.length!==1)throw Error('expected singleton declaration');
const start=statement.getStart(sf),end=statement.end,name=variable.name.getText(sf),type=variable.type.getText(sf),exported=statement.modifiers?.some(m=>m.kind===ts.SyntaxKind.ExportKeyword);
const replacement=(exported?'export ':'')+'let '+name+': '+type+'; if (1 === 1) { throw "discovery-only module initializer"; }'+'\n'.repeat((text.slice(start,end).match(/\n/g)||[]).length);
fs.writeFileSync(file,text.slice(0,start)+replacement+text.slice(end));console.log(JSON.stringify({file,start,end,name,type,replacement}));
