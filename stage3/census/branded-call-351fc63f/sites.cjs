const fs=require('fs'),path=require('path'),ts=require('typescript');
const tree=path.resolve(process.argv[2]);
const file=path.join(tree,'src/compiler/checker.ts');
const raw=fs.readFileSync(file,'utf8'),sf=ts.createSourceFile(file,raw,ts.ScriptTarget.Latest,true);
const targets=[4748,4767,4778,10911,48433,50073,50082],calls=[];
function loc(n){const p=sf.getLineAndCharacterOfPosition(n.getStart(sf));return {line:p.line+1,column:p.character+1};}
function visit(n){if(ts.isCallExpression(n))calls.push(n);ts.forEachChild(n,visit);}visit(sf);
const rows=targets.map(target=>{
 let candidates=calls.filter(n=>Math.abs(loc(n).line-target)<16 && ((n.expression.getText(sf)==='findAncestor' && n.arguments[1] && ts.isCallExpression(n.arguments[1]) && /^or(?:Predicate[23])?$/.test(n.arguments[1].expression.getText(sf))) || (target===10911 && n.expression.getText(sf).endsWith('.find') && n.arguments[0] && ts.isCallExpression(n.arguments[0]) && /^or(?:Predicate[23])?$/.test(n.arguments[0].expression.getText(sf)))));
 if(candidates.length){const distance=Math.min(...candidates.map(n=>Math.abs(loc(n).line-target)));candidates=candidates.filter(n=>Math.abs(loc(n).line-target)===distance);}
 if(candidates.length!==1)throw Error('ambiguous target '+target+': '+candidates.map(n=>n.getText(sf)));
 const n=candidates[0],owners=[];let unit;
 for(let p=n.parent;p;p=p.parent)if(ts.isFunctionDeclaration(p)){owners.unshift(p.name?.text||'<anonymous>');if(!unit)unit=p;}
 const p=loc(n),u=loc(unit);
 return {requested_line:target,site:file+':'+p.line+':'+p.column,unit:file+':'+u.line+':'+u.column,owners,text:n.getText(sf)};
});
const coreFile=path.join(tree,'src/compiler/core.ts'),coreRaw=fs.readFileSync(coreFile,'utf8'),core=ts.createSourceFile(coreFile,coreRaw,ts.ScriptTarget.Latest,true);
const functions=core.statements.filter(n=>ts.isFunctionDeclaration(n)&&['orPredicate2','orPredicate3','or'].includes(n.name?.text)).map(n=>({name:n.name.text,start:n.getStart(core),end:n.end,line:core.getLineAndCharacterOfPosition(n.getStart(core)).line+1,endLine:core.getLineAndCharacterOfPosition(n.end).line+1,text:n.getText(core)}));
console.log(JSON.stringify({tree,checker:rows,core:functions},null,2));
