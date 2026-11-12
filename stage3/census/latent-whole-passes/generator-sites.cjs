// Locate the five direct generator-fed elaborateElementwise call sites, without claiming
// they are independent first-error census units or that their bodies successfully lower.
const fs=require('fs'),ts=require('typescript');
const path=process.argv[2],text=fs.readFileSync(path,'utf8');
const sf=ts.createSourceFile(path,text,ts.ScriptTarget.Latest,true);
const rows=[];
function where(n){const p=sf.getLineAndCharacterOfPosition(n.getStart(sf));return `${p.line+1}:${p.character+1}`;}
function visit(n){
 if(ts.isCallExpression(n)&&n.expression.getText(sf)==='elaborateElementwise'){
  let owner=n.parent;while(owner&&!ts.isFunctionDeclaration(owner))owner=owner.parent;
  let top=owner;for(let p=owner?.parent;p;p=p.parent)if(ts.isFunctionDeclaration(p))top=p;
  rows.push({where:where(n),owner:owner?.name?.text,owner_where:owner?where(owner):null,top:top?.name?.text,top_where:top?where(top):null,argument:n.arguments[0]?.getText(sf)});
 }
 ts.forEachChild(n,visit);
}
visit(sf);console.log(JSON.stringify(rows,null,2));
