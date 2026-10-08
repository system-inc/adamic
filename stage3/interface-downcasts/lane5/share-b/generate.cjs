// Generate independent per-rank fixtures, retaining original declarations and reads.
const fs = require('fs');
const path = require('path');
const ts = require(path.resolve(__dirname, '../../../fixtures/assertions/api/node_modules/typescript'));
const evidence = JSON.parse(fs.readFileSync(path.join(__dirname, process.argv[2] || 'original-witnesses.json')));
const ranks = new Set(process.argv[3] ? process.argv[3].split(',').map(Number) : [331,334,349,352,355,358,361,364,367,376,379,394,397,400,412,481,622,634,748,778]);
const ledger = JSON.parse(fs.readFileSync(path.join(__dirname, "../unknown-callable-pairs-ranked.json")));
const families = [];
for (const member of evidence.members.filter(m => ranks.has(m.rank))) {
 const source = ts.createSourceFile('member.a', 'interface Target {' + member.declaration + '}', ts.ScriptTarget.Latest, true);
 const declaration = source.statements[0].members[0];
 const signature = ts.isPropertySignature(declaration) ? declaration.type : declaration;
 const names = new Map();
 function references(node) {
  if (ts.isTypeReferenceNode(node)) names.set(node.typeName.getText(source), Math.max(names.get(node.typeName.getText(source)) || 0, node.typeArguments?.length || 0));
  ts.forEachChild(node, references);
 }
 references(signature);
 const carriers = [...names].map(([n, arity]) => 'interface ' + n + (arity ? '<' + Array.from({length:arity},(_,i)=>'T'+i).join(',') + '>' : '') + ' { readonly value: number; }').join('\n');
 const parameters = signature.parameters.map(p => p.getText(source)).join(', ');
 const parameterTypes = signature.parameters.map(p => p.type.getText(source));
 const resultType = signature.type.getText(source);
 const args = parameterTypes.map(t => t.includes('[]') ? '[{value:3}]' : t === 'string' ? '"value3"' : t === 'boolean' ? 'true' : t === 'number' ? '3' : '{value:3}');
 const observations = signature.parameters.map((p,i) => {
  const n = p.name.getText(source), t = parameterTypes[i];
  if (t.includes('[]')) return t.includes('undefined') ? `if (${n} !== undefined) {for (const element of ${n}) {console.log(\`\${element.value}\`);}}` : `for (const element of ${n}) {console.log(\`\${element.value}\`);}`;
  if (t === 'string') return `console.log(${n});`;
  if (t === 'boolean' || t === 'number') return `console.log(\`\${${n}}\`);`;
  return t.includes('undefined') ? `if (${n} !== undefined) {console.log(\`\${${n}.value}\`);}` : `console.log(\`\${${n}.value}\`);`;
 }).join('');
 const result = resultType === 'void' ? '' : resultType === 'string' ? 'return "answer3";' : resultType === 'boolean' ? 'return true;' : resultType === 'number' ? 'return 3;' : 'return {value:3};';
 const invocation = member.read + '(' + args.join(',') + ')';
 let call = resultType === 'void' ? invocation + ';' : resultType === 'string' ? 'console.log(' + invocation + ');' : resultType === 'boolean' || resultType === 'number' ? 'console.log(`${' + invocation + '}`);' : 'console.log(`${' + invocation + '.value}`);';
 if (resultType.includes('undefined')) call = 'const result=' + invocation + ';if (result !== undefined) {console.log(`${result.value}`);}' ;
 let receiver;
 if (member.read.startsWith('context.factory.converters.')) receiver = 'const context = {factory:{converters:value as Target}};';
 else if (member.read.startsWith('context.factory.')) receiver = 'const context = {factory:value as Target};';
 else if (member.read.startsWith('emitHelpers().')) receiver = 'function emitHelpers():Target {return value as Target;}';
 else if (member.read.startsWith('(host as Program).')) receiver = 'const host = value;';
 else receiver = 'const ' + member.read.slice(0, member.read.lastIndexOf('.')).replace(/\?$/, '') + '=value as Target;';
 const prefix = '// Original declaration and read; adjacent data carriers reduced.\n' + carriers + '\ninterface Base {readonly ' + member.field + ':unknown;}\ninterface Target {' + member.declaration + '}\n' + (member.rank === 622 ? 'interface Program extends Target {}\n' : '') + 'function probe(value:Base):void {' + receiver + call + '}\n';
 const variants = {
  good: `(${parameters}):${resultType}=>{${observations}${result}}`,
  'wrong-value':'7',
  'wrong-arity': signature.parameters.length ? `():${resultType}=>{${result}}` : `(extra:number):${resultType}=>{${result}}`
 };
 if (resultType !== 'void') variants['wrong-result'] = `(${parameters}):number=>9`;
 if (resultType === 'number') variants['wrong-result'] = `(${parameters}):string=>"wrong"`;
 if (signature.parameters.length) variants['wrong-members'] = '(' + signature.parameters.map(p => p.name.getText(source) + (p.type.getText(source) === 'number' ? ':string' : ':number')).join(',') + '):' + resultType + '=>{' + result + '}';
 const directory = 'rank-' + member.rank;
 const destination = path.join(__dirname, directory);
 fs.mkdirSync(destination, {recursive:true});
 for (const [variant, producer] of Object.entries(variants)) fs.writeFileSync(path.join(destination, variant + '.a'), prefix + 'probe({' + member.field + ':' + producer + '});\n');
 families.push({...member, expected:ledger.find(p=>p.rank===member.rank).declared_types[0], directory, resultType, arity:signature.parameters.length ? '0':'1', variants:Object.keys(variants)});
}
fs.writeFileSync(path.join(__dirname, process.argv[4] || 'families.json'), JSON.stringify(families,null,2)+'\n');
console.log('Generated ' + families.length + ' families.');
