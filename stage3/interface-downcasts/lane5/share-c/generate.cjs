// Generate member-contract witnesses; surrounding carriers are explicitly reduced.
const fs=require('fs'), path=require('path');
const ts=require(path.join(process.argv[2],'lib/typescript.js'));
const carriers=JSON.parse(fs.readFileSync('/tmp/lane5-c-carrier-declarations.json'));
const evidence=JSON.parse(fs.readFileSync(path.join(__dirname,'original-members.json')));
const maximum=Number(process.argv[3] || 500);
const manifest=[];
const preserved=new Map(JSON.parse(fs.readFileSync(path.join(__dirname,'certified.json'))).map(r=>[r.rank,r]));
function parse(text) {return ts.createSourceFile('declaration.a',text,ts.ScriptTarget.Latest,true);}
for (const member of evidence.members.filter(m=>m.rank<=maximum)) {
 if(preserved.has(member.rank)){manifest.push({...preserved.get(member.rank),status:'generated'});continue;}
 const row={rank:member.rank, reads:member.reads, type:member.type, field:member.field, read:member.read};
 manifest.push(row);
 if (/Set</.test(member.type)) {row.status='delegated Set intrinsic';continue;}
 if (/\[\]|Map<|Map$|SymbolTable|Array<|RegExp|Math|ObjectConstructor|Buffer/.test(member.type)) {row.status='original intrinsic receiver requires dedicated witness';continue;}
 if (member.declarations.length!==1) {row.status='overloaded original member';continue;}
 const declaration=member.declarations[0].text;
 if (/^static\s/.test(declaration)||member.read.startsWith('this.')){row.status='original class or this receiver requires separate context';continue;}
 if (/\{/.test(declaration) && !declaration.trim().endsWith(';')) {row.status='original class receiver requires method witness';continue;}
 const target=parse('interface Target {'+declaration+'}').statements[0];
 const field=target.members?.[0];
 if (!field || !(ts.isMethodSignature(field)||ts.isPropertySignature(field))) {row.status='original namespace function or non-callable declaration';continue;}
 if (field.questionToken) {row.status='delegated optional host method';continue;}
 let signature=ts.isMethodSignature(field)?field:field.type;
 if (ts.isTypeReferenceNode(signature)) {
  const alias=carriers[signature.typeName.getText()];
  if(alias?.kind==='TypeAliasDeclaration') signature=parse(alias.text).statements[0].type;
 }
 if (!signature.parameters || !signature.type) {row.status='callable alias requires separate original context';continue;}
 if (signature.typeParameters?.length || signature.parameters.some(p=>p.dotDotDotToken)) {row.status='generic or rest signature';continue;}
 if (!/^\w+(?:\.\w+)+$/.test(member.read)) {row.status='binding or effectful original read requires separate context';continue;}
 const definitions=new Map(), visiting=new Set();
 function named(name,args) {
  if (['ReadonlyArray','Array'].includes(name)) return;
  if (/^(Map|Set|ReadonlyMap|ReadonlySet|Date|RegExp|MapIterator|SetIterator|IterableIterator)$/.test(name)) throw Error('nominal or intrinsic payload '+name);
  if (definitions.has(name)||visiting.has(name))return;
  visiting.add(name);
  const original=carriers[name];
  if (original?.kind==='TypeAliasDeclaration') {
   const alias=parse(original.text).statements[0];
   if(alias.typeParameters?.length) throw Error('generic alias '+name);
   definitions.set(name,original.text); visitType(alias.type);
  } else if (original?.kind==='EnumDeclaration') {
   definitions.set(name,'enum '+name+' {'+Object.entries(original.values).map(([k,v])=>k+'='+JSON.stringify(v)).join(',')+'}');
  } else {
   const generic=args?.length?'<' + args.map((_,i)=>'T'+i).join(',')+'>':'';
   definitions.set(name,'interface '+name+generic+' {readonly value:number;'+(original?.kindValue!==undefined?'readonly kind:'+original.kindValue+';':'')+'}');
  }
  visiting.delete(name);
 }
 function visitType(t) {
  if(ts.isTypeReferenceNode(t)) {
   const name=t.typeName.getText();
   if(t.typeName.kind!==ts.SyntaxKind.Identifier){const base=t.typeName.left.getText();if(carriers[base]?.kind!=='EnumDeclaration')throw Error('qualified type '+name);named(base);return;}
   named(name,t.typeArguments);for(const a of t.typeArguments||[])visitType(a);return;
  }
  if (ts.isFunctionTypeNode(t)) {if(t.typeParameters?.length)throw Error('generic callback');for(const p of t.parameters)visitType(p.type);visitType(t.type);return;}
  ts.forEachChild(t,n=>{if(ts.isTypeNode(n))visitType(n);});
 }
 function value(t, seen=new Set()) {
  if(ts.isParenthesizedTypeNode(t))return value(t.type,seen);
  if(ts.isUnionTypeNode(t))return value(t.types.find(x=>x.kind!==ts.SyntaxKind.UndefinedKeyword&&x.kind!==ts.SyntaxKind.NullKeyword)||t.types[0],seen);
  if(ts.isArrayTypeNode(t))return '['+value(t.elementType,seen)+']';
  if(ts.isTypeOperatorNode(t))return value(t.type,seen);
  if(ts.isFunctionTypeNode(t))return '('+t.parameters.map(p=>p.getText()).join(',')+'):'+t.type.getText()+'=>'+(t.type.kind===ts.SyntaxKind.VoidKeyword?'{}':'('+value(t.type,seen)+')');
  if(ts.isTypeReferenceNode(t)) {
   const name=t.typeName.getText();
   if(t.typeName.kind!==ts.SyntaxKind.Identifier){const d=carriers[t.typeName.left.getText()];if(d?.kind==='EnumDeclaration')return String(d.values[t.typeName.right.getText()]);}
   if(carriers[name]?.kind==='EnumDeclaration')return String(Object.values(carriers[name].values)[0]);
   if(['ReadonlyArray','Array'].includes(name))return '['+value(t.typeArguments[0],seen)+']';
   if(seen.has(name))throw Error('recursive alias value '+name);
   const d=carriers[name];
   if(d?.kind==='TypeAliasDeclaration')return value(parse(d.text).statements[0].type,new Set([...seen,name]));
   return '{value:3'+(d?.kindValue!==undefined?',kind:'+d.kindValue:'')+'}';
  }
  if(ts.isLiteralTypeNode(t))return t.literal.getText();
  switch(t.kind) {
   case ts.SyntaxKind.StringKeyword:return '"abc"';
   case ts.SyntaxKind.NumberKeyword:return '3';
   case ts.SyntaxKind.BooleanKeyword:return 'true';
   case ts.SyntaxKind.UndefinedKeyword:case ts.SyntaxKind.VoidKeyword:return 'undefined';
   default:throw Error('unsupported value '+t.getText());
  }
 }
 try {
  for(const p of signature.parameters)visitType(p.type);visitType(signature.type);
  if(ts.isPropertySignature(field)&&ts.isTypeReferenceNode(field.type))named(field.type.typeName.getText());
  const args=signature.parameters.map(p=>value(p.type));
  const result=value(signature.type);
  const body=signature.type.kind===ts.SyntaxKind.VoidKeyword?'{}':'('+result+')';
  // The exact original member read is retained, including its receiver path.
  const parts=member.read.split('.'), leaf=parts.pop();let owner='value as Target';
  for(let i=parts.length-1;i>0;i--)owner='{'+parts[i]+':'+owner+'}';
  const prefix='// Original member declaration and read; adjacent carriers and bodies reduced.\n'+[...definitions.values()].join('\n')+'\ninterface Base {readonly '+leaf+':unknown;}\ninterface Target {'+declaration+'}\n';
  const probe='function probe(value:Base):void {const '+parts[0]+'='+owner+';'+(signature.type.kind===ts.SyntaxKind.VoidKeyword?'':'const witnessResult=')+member.read+'('+args.join(',')+');console.log("completed");}\n';
  const good='('+signature.parameters.map(p=>p.getText()).join(',')+'):'+signature.type.getText()+'=>' + body;
  const variants={good, 'wrong-arity':signature.parameters.length?'(): '+signature.type.getText()+'=>' + body:'(extra:number): '+signature.type.getText()+'=>' + body};
  const directory=path.join(__dirname,'families','rank-'+member.rank);
  fs.mkdirSync(directory,{recursive:true});
  for(const [variant,implementation] of Object.entries(variants))fs.writeFileSync(path.join(directory,variant+'.a'),prefix+probe+'probe({'+leaf+':'+implementation+'});\n');
  row.status='generated';row.arity=signature.parameters.length;row.foundArity=signature.parameters.length?0:1;row.declaration=declaration;
 } catch(error) {row.status='witness preparation boundary: '+error.message;}
}
fs.writeFileSync(path.join(__dirname,'manifest.json'),JSON.stringify(manifest,null,2)+'\n');
console.log('Generated '+manifest.filter(m=>m.status==='generated').length+' families from '+manifest.length+' ranked members.');
