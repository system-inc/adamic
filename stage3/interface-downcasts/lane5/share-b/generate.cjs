// Generate independent per-rank fixtures, retaining original declarations and reads.
const fs = require('fs');
const path = require('path');
const ts = require(path.resolve(__dirname, '../../../fixtures/assertions/api/node_modules/typescript'));
const evidence = JSON.parse(fs.readFileSync(path.join(__dirname, process.argv[2] || 'original-witnesses.json')));
const ranks = new Set(process.argv[3] ? process.argv[3].split(',').map(Number) : [331,334,349,352,355,358,361,364,367,376,379,394,397,400,412,481,622,634,748,778]);
const ledger = JSON.parse(fs.readFileSync(path.join(__dirname, "../unknown-callable-pairs-ranked.json")));
const families = [];
const preserveArrayCarriers = process.env.ADAMIC_CALLABLE_SHARE_B_ARRAY_CARRIERS === '1';
const enumDeclarations = new Map();
const enumMembers = new Map();
const enumSources = new Map();
const stringEnums = new Set();
let originalTypes = ts.createSourceFile('src/compiler/types.ts',fs.readFileSync('/tmp/lane5-b-original/src/compiler/types.ts','utf8'),ts.ScriptTarget.Latest,true);
function enums(node) {
 if (ts.isEnumDeclaration(node) && node.members[0].initializer && ts.isStringLiteral(node.members[0].initializer)) stringEnums.add(node.name.text);
 if (ts.isEnumDeclaration(node)) enumMembers.set(node.name.text,node.members[0].name.getText(originalTypes));
 if (ts.isEnumDeclaration(node)) enumSources.set(node.name.text, {file:originalTypes.fileName,sha256:require('crypto').createHash('sha256').update(originalTypes.text).digest('hex')});
 if (ts.isEnumDeclaration(node)) enumDeclarations.set(node.name.text,node.getText(originalTypes).replace(/^export\s+/,'').replace(/^const\s+enum/,'enum').replace(/\r\n/g,'\n').replace(/[ \t]+$/gm,''));
 ts.forEachChild(node,enums);
}
enums(originalTypes);
for (const file of ['src/compiler/corePublic.ts','src/compiler/factory/emitHelpers.ts']) {
 originalTypes=ts.createSourceFile(file,fs.readFileSync('/tmp/lane5-b-original/'+file,'utf8'),ts.ScriptTarget.Latest,true);enums(originalTypes);
}

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
 if (member.rank===658) names.set('ModuleKind',0);
 const carriers = [...names].map(([n, arity]) => preserveArrayCarriers && n === 'NodeArray' ? 'interface NodeArray<T> extends ReadonlyArray<T> {}' : (n==='ResolutionMode' && member.rank===658 ? 'type ResolutionMode = ModuleKind.ESNext | ModuleKind.CommonJS | undefined;' : enumDeclarations.get(n)) || 'interface ' + n + (arity ? '<' + Array.from({length:arity},(_,i)=>'T'+i).join(',') + '>' : '') + ' { readonly value: number; }').join('\n');
 const parameters = signature.parameters.map(p => p.getText(source)).join(', ');
 const parameterTypes = signature.parameters.map(p => p.type.getText(source));
 const resultType = signature.type.getText(source);
 const args = parameterTypes.map(type => {const t=type.replace(/\s*\|\s*undefined$/, '');return (t.includes('[]') || preserveArrayCarriers && t.includes('NodeArray<')) ? '[{value:3}]' : enumDeclarations.has(t) ? t + '.' + enumMembers.get(t) : t === 'string' ? '"value3"' : t === 'boolean' ? 'true' : t === 'number' ? '3' : '{value:3}';});
 const observations = signature.parameters.map((p,i) => {
  const n = p.name.getText(source), t = parameterTypes[i];
  if (t.includes('[]') || preserveArrayCarriers && t.includes('NodeArray<')) return t.includes('undefined') || p.questionToken ? `if (${n} !== undefined) {for (const element of ${n}) {console.log(\`\${element.value}\`);}}` : `for (const element of ${n}) {console.log(\`\${element.value}\`);}`;
  if (t === 'string') return `console.log(${n});`;
  if (/^(boolean|number)(\s*\|\s*undefined)?$/.test(t) || enumDeclarations.has(t)) return `console.log(\`\${${n}}\`);`;
  if (t.includes('string')) {const observation=`if (typeof ${n} === \"string\") {console.log(${n});} else {console.log(\`\${${n}.value}\`);}`;return p.questionToken || t.includes('undefined') ? `if (${n} !== undefined) {${observation}}` : observation;}
  return t.includes('undefined') || p.questionToken ? `if (${n} !== undefined) {console.log(\`\${${n}.value}\`);}` : `console.log(\`\${${n}.value}\`);`;
 }).join('');
 const arrayResult = resultType.includes('[]') || preserveArrayCarriers && resultType.includes('NodeArray<');
 const result = arrayResult ? (resultType.includes('| undefined)[]') ? 'return [{value:3},undefined];' : 'return [{value:3}];') : resultType === 'void' ? '' : resultType === 'string' ? 'return "answer3";' : resultType === 'boolean' ? 'return true;' : resultType === 'number' ? 'return 3;' : enumDeclarations.has(resultType) ? 'return '+resultType+'.'+enumMembers.get(resultType)+';' : member.rank===658 ? 'return ModuleKind.ESNext;' : 'return {value:3};';
 const invocation = member.read + '(' + args.join(',') + ')';
 let call = resultType === 'void' ? invocation + ';' : resultType === 'string' ? 'console.log(' + invocation + ');' : resultType === 'boolean' || resultType === 'number' ? 'console.log(`${' + invocation + '}`);' : 'console.log(`${' + invocation + '.value}`);';
 if (resultType.includes('undefined')) call = 'const result=' + invocation + ';if (result !== undefined) {console.log(`${result.value}`);}' ;
 if (enumDeclarations.has(resultType) || member.rank===658) call = 'console.log(`${'+invocation+'}`);';
 if (arrayResult) call = 'const result=' + invocation + ';if (result !== undefined) {for (const element of result) {if (element !== undefined) {console.log(`${element.value}`);}}}';
 let receiver;
 if (member.read.startsWith('context.factory.converters.')) receiver = 'const context = {factory:{converters:value as Target}};';
 else if (member.read.startsWith('context.factory.')) receiver = 'const context = {factory:value as Target};';
 else if (member.read.startsWith('context.getEmitResolver().')) receiver = 'const context={getEmitResolver:():Target=>value as Target};';
 else if (member.read.startsWith('emitHelpers().')) receiver = 'function emitHelpers():Target {return value as Target;}';
 else if (member.read.startsWith('(host as Program).')) receiver = 'const host = value;';
 else if (/^[A-Za-z_$][\w$]*\(\)\./.test(member.read)) receiver = 'function ' + member.read.slice(0,member.read.indexOf('(')) + '():Target {return value as Target;}';
 else {const parts=member.read.slice(0,member.read.lastIndexOf('.')).replace(/[?!]$/,'').split('.');let holder='value as Target';for(const part of parts.slice(1).reverse()) holder='{'+part+':'+holder+'}';receiver='const '+parts[0]+'='+holder+';'}
 let prefix = '// Original declaration and read; adjacent data carriers reduced.\n' + carriers + '\ninterface Base {readonly ' + member.field + ':unknown;}\ninterface Target {' + member.declaration + '}\n' + (member.rank === 622 ? 'interface Program extends Target {}\n' : '') + 'function probe(value:Base):void {' + receiver + call + '}\n';
 if ([268,739].includes(member.rank)) {
  const slot=member.rank===268?'source':'mapper1', type=member.rank===268?'DebugType':'DebugTypeMapper';
  prefix='// Original declaration and read; adjacent data carriers reduced.\n'+carriers+'\ninterface Base {readonly '+member.field+':unknown;}\ninterface Target {'+member.declaration+'}\ninterface '+type+' extends Target {}\nclass Witness {readonly '+slot+':Base;constructor(value:Base){this.'+slot+'=value;}check():void {'+call+'}}\nfunction probe(value:Base):void {new Witness(value).check();}\n';
 }
 const variants = {
  good: `(${parameters}):${resultType}=>{${observations}${result}}`,
  'wrong-value':'7',
  'wrong-arity': signature.parameters.length ? `():${resultType}=>{${result}}` : `(extra:number):${resultType}=>{${result}}`
 };
 if (resultType !== 'void') variants['wrong-result'] = `(${parameters}):number=>9`;
 if (resultType === 'number' || enumDeclarations.has(resultType) || member.rank===658) variants['wrong-result'] = `(${parameters}):string=>"wrong"`;
 if (signature.parameters.length) variants['wrong-members'] = '(' + signature.parameters.map(p => p.name.getText(source) + (p.type.getText(source) === 'number' || enumDeclarations.has(p.type.getText(source)) && !stringEnums.has(p.type.getText(source)) ? ':string' : ':number')).join(',') + '):' + resultType + '=>{' + result + '}';
 const directory = 'rank-' + member.rank + (preserveArrayCarriers ? '-array-carrier' : '');
 const destination = path.join(__dirname, directory);
 fs.mkdirSync(destination, {recursive:true});
 for (const [variant, producer] of Object.entries(variants)) fs.writeFileSync(path.join(destination, variant + '.a'), prefix + 'probe({' + member.field + ':' + producer + '});\n');
 const carrierEnums = [...names].filter(([name])=>enumDeclarations.has(name)).map(([name])=>({name,declaration:enumDeclarations.get(name),sourceFile:enumSources.get(name).file,sha256:enumSources.get(name).sha256}));
 families.push({...member, carrierEnums, nodeErrors:arrayResult?['wrong-result']:[], carrierSource:member.rank===658?{filename:'src/compiler/types.ts',sha256:member.declarationSha256}:undefined, carrierDeclarations:[...(preserveArrayCarriers?['interface NodeArray<T> extends ReadonlyArray<T> {}']:[]),...(member.rank===658?['type ResolutionMode = ModuleKind.ESNext | ModuleKind.CommonJS | undefined;']:[])], expected:ledger.find(p=>p.rank===member.rank).declared_types[0], directory, resultType, arity:signature.parameters.length ? '0':'1', variants:Object.keys(variants)});
}
fs.writeFileSync(path.join(__dirname, process.argv[4] || 'families.json'), JSON.stringify(families,null,2)+'\n');
console.log('Generated ' + families.length + ' families.');
