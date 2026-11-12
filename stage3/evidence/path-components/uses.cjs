const fs=require('fs'),path=require('path'),ts=require('/workspace/adamic/stage3/api/node_modules/typescript');
const [tree,out]=process.argv.slice(2),uses=[];
for(const f of fs.readdirSync(tree+'/src/compiler',{recursive:true}).filter(f=>f.endsWith('.ts')).sort()){
 const text=fs.readFileSync(tree+'/src/compiler/'+f,'utf8'),sf=ts.createSourceFile(f,text,99,true);
 function visit(n){if(ts.isIdentifier(n)&&n.text==='PathPathComponents'){
 const p=sf.getLineAndCharacterOfPosition(n.getStart(sf));let kind=ts.SyntaxKind[n.parent.kind];
 if(ts.isTypeReferenceNode(n.parent)&&ts.isTypeReferenceNode(n.parent.parent)&&n.parent.parent.typeName.getText(sf)==='Readonly')kind='Readonly<PathPathComponents>';
 if(ts.isTypeReferenceNode(n.parent)&&(ts.isAsExpression(n.parent.parent)||ts.isTypeAssertionExpression(n.parent.parent)))kind='cast';
 uses.push({file:'src/compiler/'+f,line:p.line+1,column:p.character+1,kind,source:text.split(/\r?\n/)[p.line]});
 }ts.forEachChild(n,visit);}visit(sf);
}
const counts={};for(const u of uses)counts[u.kind]=(counts[u.kind]||0)+1;
fs.writeFileSync(out,JSON.stringify({total:uses.length,counts,uses},null,2)+'\n');console.log(JSON.stringify({total:uses.length,counts}));
