/* Distinguish scalar unions that the original probe's Object|Union test admits. */
const ts=require('/workspace/view-pairs/stage3/api/node_modules/typescript');
const fs=require('fs'),path=require('path');
const labels=JSON.parse(fs.readFileSync(process.argv[2],'utf8'));
const files=JSON.parse(fs.readFileSync(process.argv[3],'utf8'));
const enums=new Map(),aliases=new Map();
for(const file of files){
 const sf=ts.createSourceFile(file,fs.readFileSync(file,'utf8'),ts.ScriptTarget.Latest,true);
 function visit(n){
  if(ts.isEnumDeclaration(n)){
   const item={file,line:sf.getLineAndCharacterOfPosition(n.getStart(sf)).line+1,members:n.members.map(m=>m.name.getText(sf))};
   if(!enums.has(n.name.text))enums.set(n.name.text,[]);enums.get(n.name.text).push(item);
  }
  if(ts.isTypeAliasDeclaration(n)&&!n.typeParameters?.length){
   if(!aliases.has(n.name.text))aliases.set(n.name.text,[]);aliases.get(n.name.text).push({node:n.type,file,line:sf.getLineAndCharacterOfPosition(n.getStart(sf)).line+1});
  }
  ts.forEachChild(n,visit);
 }
 visit(sf);
}
const scalarKeywords=new Set([ts.SyntaxKind.StringKeyword,ts.SyntaxKind.NumberKeyword,ts.SyntaxKind.BooleanKeyword,ts.SyntaxKind.BigIntKeyword,ts.SyntaxKind.SymbolKeyword,ts.SyntaxKind.UndefinedKeyword,ts.SyntaxKind.VoidKeyword,ts.SyntaxKind.NeverKeyword]);
function scalar(n,seen=new Set()){
 if(scalarKeywords.has(n.kind)||ts.isLiteralTypeNode(n)||ts.isTemplateLiteralTypeNode(n))return true;
 if(ts.isParenthesizedTypeNode(n))return scalar(n.type,seen);
 if(ts.isUnionTypeNode(n))return n.types.every(t=>scalar(t,seen));
 // keyof always denotes property keys, never an object value.
 if(ts.isTypeOperatorNode(n)&&n.operator===ts.SyntaxKind.KeyOfKeyword)return true;
 if(!ts.isTypeReferenceNode(n)||n.typeArguments?.length)return false;
 const text=n.typeName.getText();const parts=text.split('.');
 if(parts.length>1){const name=parts.at(-2),member=parts.at(-1);return enums.has(name)&&enums.get(name).every(e=>e.members.includes(member));}
 if(enums.has(text))return true;
 if(seen.has(text)||!aliases.has(text))return false;
 const next=new Set(seen);next.add(text);return aliases.get(text).every(a=>scalar(a.node,next));
}
const result={typescript_version:ts.version,method:'Source enum declarations and syntactically proven scalar aliases/unions; ambiguous labels remain unclassified. Intersection targets are handled as objects by the census, regardless of this scalar registry.',enums:Object.fromEntries(enums),scalar_labels:{}};
for(const label of labels){
 const sf=ts.createSourceFile('__view_pair_type.ts','type __Probe = '+label+';',ts.ScriptTarget.Latest,true);
 const n=sf.statements[0];
 if(!sf.parseDiagnostics.length&&n&&ts.isTypeAliasDeclaration(n)&&scalar(n.type))result.scalar_labels[label]={scalar:true,basis:'Primitive/literal/property-key type, source enum, or union/alias made entirely from such scalar types.'};
}
fs.writeFileSync(process.argv[4],JSON.stringify(result,null,2)+'\n');
console.log(JSON.stringify({labels:labels.length,proven_scalar_labels:Object.keys(result.scalar_labels).length}));
