const fs=require('fs'),ts=require('/home/agent/.cache/adamic-stage3/api/node_modules/typescript');
const [file,line,column]=process.argv.slice(2),text=fs.readFileSync(file,'utf8'),sf=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true),pos=sf.getPositionOfLineAndCharacter(+line-1,+column-1);
const selected=sf.statements.find(n=>ts.isFunctionDeclaration(n)&&n.pos<=pos&&pos<n.end);if(!selected)throw Error('no overload owner');const name=selected.name.text,edits=[];
for(const n of sf.statements)if(ts.isFunctionDeclaration(n)&&n.name?.text===name){
 for(const p of n.parameters)if(p.type){function walk(x){if(ts.isTypePredicateNode(x))edits.push({start:x.getStart(sf),end:x.end,value:'boolean'});else ts.forEachChild(x,walk);}walk(p.type);}
 if(n.type&&ts.isTypePredicateNode(n.type))edits.push({start:n.type.getStart(sf),end:n.type.end,value:'boolean'});
 if(n.body)edits.push({start:n.body.getStart(sf),end:n.body.end,value:'{ throw "discovery-only overload group";'+ '\n'.repeat((n.body.getText(sf).match(/\n/g)||[]).length)+'}'});
}
let result=text;for(const e of edits.sort((a,b)=>b.start-a.start))result=result.slice(0,e.start)+e.value+result.slice(e.end);fs.writeFileSync(file,result);console.log(JSON.stringify({name,edits}));
