// Preserve full member declarations and read expressions in refusal probes.
const fs = require('fs');
const path = require('path');
const ts = require('/tmp/lane5-b-original/lib/typescript.js');
const evidence = require('./batch-02-original.json');
const ranks = new Set([16, 25, 28, 40, 55, 130, 172, 181, 190, 226, 244, 250, 253, 319, 322]);
const probes = [];
for (const member of evidence.members.filter(m => ranks.has(m.rank))) {
 const source = ts.createSourceFile('probe.a', 'interface Target {'+member.declaration+'}', ts.ScriptTarget.Latest, true);
 const names = new Map(), parameters = new Set(), namespaces = new Map();
 function visit(node) {
  if (ts.isTypeParameterDeclaration(node)) parameters.add(node.name.text);
  if (ts.isTypeReferenceNode(node)) {
   const name = node.typeName.getText(source);
   if (name.includes('.')) {
    const [namespace, item] = name.split('.');
    namespaces.set(namespace, [...(namespaces.get(namespace)||[]),item]);
   } else names.set(name, Math.max(names.get(name)||0,node.typeArguments?.length||0));
  }
  ts.forEachChild(node, visit);
 }
 visit(source);
 const builtin = new Set(['ArrayLike','NonNullable','PropertyDescriptorMap','ThisType']);
 let carriers = [...names].filter(([name])=>!parameters.has(name)&&!builtin.has(name)&&!namespaces.has(name)).map(([name,arity])=>`interface ${name}${arity?'<'+Array.from({length:arity},(_,i)=>'T'+i).join(',')+'>':''} {readonly value:number;}`).join('\n');
 for (const [name,items] of namespaces) carriers += '\nenum '+name+' {'+[...new Set(items)].map((item,i)=>item+'='+i).join(',')+'}';
 // Enum values are unused: these probes only read the original callable field.
 let receiver = member.read.slice(0, member.read.lastIndexOf('.')).replace(/\?$/, '');
 let scaffold = `const ${receiver}=value as Target;`;
 if (receiver.includes('.')) {const parts=receiver.split('.'); scaffold='const '+parts[0]+'={'+parts[1]+':value as Target};';}
 if (member.rank===319) scaffold='const type={types:value as Target};';
 if (member.rank===319) carriers+='\ninterface UnionType {readonly types:Target;}';
 if (member.rank!==319 && !/^[A-Za-z_$][\w$]*(\.[A-Za-z_$][\w$]*)?$/.test(receiver)) throw Error('unsupported scaffold '+member.rank+': '+receiver);
 const signature=source.statements[0].members[0];
 const args=signature.parameters.map(p=>{
  const text=p.type.getText(source);
  if (p.questionToken || text.includes('undefined')) return 'undefined';
  if (text==='string') return '"key"';
  if (text==='number' || text==='any') return '3';
  if (text==='boolean') return 'true';
  if (ts.isFunctionTypeNode(p.type)) return '()=>3';
  if (text.startsWith('SyntaxKind.')) return text;
  if (p.dotDotDotToken) return text==='number[]' ? '3' : '{value:3}';
  if (text.includes('[]')) return '[]';
  if (text==='PropertyDescriptorMap & ThisType<any>') return '{}';
  return '{value:3}';
 });
 const filename = 'rank-'+member.rank+'/contract-probe.a';
 fs.mkdirSync(path.join(__dirname,'rank-'+member.rank),{recursive:true});
 fs.writeFileSync(path.join(__dirname,filename),'// Original member/read; adjacent carriers reduced for signature admission only.\n'+carriers+'\ninterface Base {readonly '+member.field+':unknown;}\ninterface Target {'+member.declaration+'}\nfunction probe(value:Base):void {'+scaffold+member.read+'('+args.join(',')+');}\nprobe({'+member.field+':7});\n');
 probes.push({...member,filename,refusal:member.rank===181 ? 'adamic/no-type-predicate' : 'checked view read of field '+member.field+' with unsupported callable contract'});
}
fs.writeFileSync(path.join(__dirname,'batch-02-probes.json'),JSON.stringify(probes,null,2)+'\n');
console.log('Prepared '+probes.length+' original-signature admission probes.');
