const fs=require('fs'),ts=require('/workspace/adamic/stage3/api/node_modules/typescript');
const file=process.argv[2]+'/src/compiler/core.ts',text=fs.readFileSync(file,'utf8'),source=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true);
let result=text;const pins=[];
for(const n of source.statements.filter(n=>ts.isFunctionDeclaration(n)&&['extend','copyProperties'].includes(n.name?.text)).reverse()){
const body=text.slice(n.body.getStart(source)+1,n.body.end-1);
const wrapped='{ return (globalThis as any).__coreCopies('+JSON.stringify(n.name.text)+', first, second, () => {'+body+'}); }';
pins.push({owner:n.name.text,before:n.getText(source),body});result=result.slice(0,n.body.getStart(source))+wrapped+result.slice(n.body.end);
}if(pins.length!==2)throw Error('owner count');fs.writeFileSync(file,result);console.log(JSON.stringify(pins,null,2));
