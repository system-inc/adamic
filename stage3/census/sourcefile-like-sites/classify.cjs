const fs=require('fs'),path=require('path');
const ts=require('../../api/node_modules/typescript');
const out=__dirname, root='/workspace/scratch/baobab-adapted/src/compiler';
const findings=JSON.parse(fs.readFileSync(path.join(out,'findings.json'),'utf8'));
const cache=new Map();
function source(name){if(!cache.has(name))cache.set(name,ts.createSourceFile(name,fs.readFileSync(path.join(root,name),'utf8'),ts.ScriptTarget.Latest,true));return cache.get(name);}
function fname(n,sf){if(n.name)return n.name.getText(sf);if(n.parent&&(ts.isVariableDeclaration(n.parent)||ts.isPropertyAssignment(n.parent)))return n.parent.name.getText(sf);return '<anonymous@'+(sf.getLineAndCharacterOfPosition(n.getStart(sf)).line+1)+'>';}
for(const f of findings){
 const m=f.where.match(/(?:src\/compiler\/)?([^:]+):(\d+):(\d+)$/);if(!m)throw Error(f.where);
 let name=m[1];if(name.includes('/src/compiler/'))name=name.split('/src/compiler/')[1];
 const sf=source(name),pos=sf.getPositionOfLineAndCharacter(+m[2]-1,+m[3]-1);
 let deepest=sf;function walk(n){if(n.getFullStart()<=pos&&pos<n.end){deepest=n;n.forEachChild(walk);}}walk(sf);
 let enclosing=deepest;while(enclosing&&!ts.isFunctionLike(enclosing))enclosing=enclosing.parent;
 f.file=name;f.enclosing_function=enclosing?fname(enclosing,sf):'<top level>';f.function_chain=[];for(let owner=enclosing;owner;owner=owner.parent){if(ts.isFunctionLike(owner))f.function_chain.unshift(fname(owner,sf));}
 let call=deepest,index=-1;while(call){if(ts.isCallExpression(call)||ts.isNewExpression(call)){index=call.arguments?.findIndex(a=>a.getFullStart()<=pos&&pos<a.end)??-1;if(index>=0)break;}call=call.parent;}
 if(call){
 f.callee=call.expression.getText(sf);f.argument_index=index;f.expression=call.getText(sf);f.view=index>=0?'call argument':'call expression (review)';}
 else{let n=deepest;while(n&&!ts.isReturnStatement(n)&&!ts.isVariableDeclaration(n)&&!ts.isBinaryExpression(n))n=n.parent;f.callee='<non-call>';f.view=n?ts.SyntaxKind[n.kind]:'other';f.expression=(n||deepest).getText(sf);}
 f.location=name+':'+m[2]+':'+m[3];
}
fs.writeFileSync(path.join(out,'sites.json'),JSON.stringify(findings,null,2)+'\n');
console.log(JSON.stringify(findings.map(f=>({where:f.location,phase:f.phase,callee:f.callee,arg:f.argument_index,fn:f.enclosing_function,expression:f.expression})),null,2));
