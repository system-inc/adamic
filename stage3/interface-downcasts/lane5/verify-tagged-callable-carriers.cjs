const fs = require('fs');
const path = require('path');
const child = require('child_process');
const pin = process.argv[2];
const ts = require(path.join(pin, 'lib/typescript.js'));
if (child.execFileSync('git', ['-C', pin, 'rev-parse', 'HEAD'], {encoding:'utf8'}).trim() !== '050880ce59e30b356b686bd3144efe24f875ebc8') throw Error('wrong TypeScript pin');
const file = path.join(pin, 'src/compiler/types.ts');
const program = ts.createProgram([file], {noResolve:true});
const checker = program.getTypeChecker();
const original = program.getSourceFile(file);
const normalized = text => text.replace(/^export\s+/, '').replace(/\s+/g, '');
const originalNodes = new Map();
function collect(source) {
 const nodes = new Map();
 function visit(node) {
  if (ts.isTypeAliasDeclaration(node) || ts.isInterfaceDeclaration(node) || ts.isEnumDeclaration(node)) nodes.set(node.name.text, node);
  ts.forEachChild(node, visit);
 }
 visit(source);
 return nodes;
}
for (const [name, node] of collect(original)) originalNodes.set(name, node);
const families = new Map([
 [278, {directory:'info-canonical-name', aliases:['GetCanonicalFileName'], aliasFile:'src/compiler/core.ts', carriers:[]}],
 [249, {directory:'tagged-template-update', aliases:['TemplateLiteral'], carriers:['TemplateExpression','NoSubstitutionTemplateLiteral']}],
 [241, {directory:'import-specifier', aliases:['ModuleExportName'], carriers:['Identifier','StringLiteral']}],
 [217, {directory:'canonical-file-name', aliases:['GetCanonicalFileName'], aliasFile:'src/compiler/core.ts', carriers:[]}],
 [212, {directory:'type-operator', aliases:[], carriers:[], enumConstants:[['SyntaxKind','KeyOfKeyword'],['SyntaxKind','ReadonlyKeyword'],['SyntaxKind','UniqueKeyword']]}],
 [192, {directory:'get-accessor-update', aliases:['PropertyName'], carriers:['Identifier','StringLiteral','NoSubstitutionTemplateLiteral','NumericLiteral','ComputedPropertyName','PrivateIdentifier','BigIntLiteral']}],
 [194, {directory:'set-accessor-update', aliases:['PropertyName'], carriers:['Identifier','StringLiteral','NoSubstitutionTemplateLiteral','NumericLiteral','ComputedPropertyName','PrivateIdentifier','BigIntLiteral']}],
 [176, {directory:'property-declaration', aliases:['PropertyName'], carriers:['Identifier','StringLiteral','NoSubstitutionTemplateLiteral','NumericLiteral','ComputedPropertyName','PrivateIdentifier','BigIntLiteral'], tokens:['QuestionToken','ExclamationToken']}],
 [177, {directory:'qualified-name', aliases:['EntityName'], carriers:['Identifier','QualifiedName']}],
 [159, {directory:'parameter-update', aliases:['BindingName','BindingPattern'], carriers:['Identifier','ObjectBindingPattern','ArrayBindingPattern'], tokens:['DotDotDotToken','QuestionToken']}],
 [160, {directory:'property-update', aliases:['PropertyName'], carriers:['Identifier','StringLiteral','NoSubstitutionTemplateLiteral','NumericLiteral','ComputedPropertyName','PrivateIdentifier','BigIntLiteral'], tokens:['QuestionToken','ExclamationToken']}],
 [122, {directory:'variable-update-tagged', aliases:['BindingName','BindingPattern'], carriers:['Identifier','ObjectBindingPattern','ArrayBindingPattern']}],
 [141, {directory:'export-specifier', aliases:['ModuleExportName'], carriers:['Identifier','StringLiteral']}],
 [144, {directory:'property-signature', aliases:['PropertyName'], carriers:['Identifier','StringLiteral','NoSubstitutionTemplateLiteral','NumericLiteral','ComputedPropertyName','PrivateIdentifier','BigIntLiteral']}],
 [145, {directory:'type-check', aliases:['TypeOfTag'], carriers:[]}],
]);
let count = 0;
for (const rank of process.argv[3].split(',').map(Number)) {
 const family = families.get(rank);
 if (!family) throw Error('missing requested tagged family');
 const directory = path.join(__dirname, 'later-ranked-callables', family.directory);
 const aliasSource = family.aliasFile ? ts.createSourceFile(family.aliasFile, fs.readFileSync(path.join(pin,family.aliasFile),'utf8'),ts.ScriptTarget.Latest,true) : original;
 const aliasNodes = family.aliasFile ? collect(aliasSource) : originalNodes;
 for (const name of fs.readdirSync(directory).filter(name => (name.endsWith('.a') || name.endsWith('.ts')))) {
  const source = ts.createSourceFile(name, fs.readFileSync(path.join(directory,name),'utf8'), ts.ScriptTarget.Latest, true);
  const fixture = collect(source);
  for (const alias of family.aliases) {
   if (!fixture.has(alias) || normalized(fixture.get(alias).getText(source)) !== normalized(aliasNodes.get(alias).getText(aliasSource))) throw Error(name+': original alias changed '+alias);
  }
  for (const carrier of family.carriers) {
   const originalKind = originalNodes.get(carrier).members.find(member => member.name?.getText(original) === 'kind');
   const fixtureKind = fixture.get(carrier)?.members.find(member => member.name?.getText(source) === 'kind');
   const expected = checker.getTypeAtLocation(originalKind.type).value;
   if (typeof expected !== 'number' || !fixtureKind || !ts.isLiteralTypeNode(fixtureKind.type) || !ts.isNumericLiteral(fixtureKind.type.literal) || Number(fixtureKind.type.literal.text) !== expected || !fixtureKind.modifiers?.some(modifier => modifier.kind === ts.SyntaxKind.ReadonlyKeyword)) throw Error(name+': original discriminator changed '+carrier);
  }
  for (const token of family.tokens || []) {
   const originalToken = originalNodes.get(token);
   const kind = checker.getTypeAtLocation(originalToken).getProperty('kind');
   const expected = checker.getTypeOfSymbolAtLocation(kind, originalToken).value;
   const actual = fixture.get(token)?.members.find(member => member.name?.getText(source) === 'kind');
   if (typeof expected !== 'number' || !actual || !ts.isLiteralTypeNode(actual.type) || !ts.isNumericLiteral(actual.type.literal) || Number(actual.type.literal.text) !== expected) throw Error(name+': original token discriminator changed '+token);
  }
  for (const [enumeration, member] of family.enumConstants || []) {
   const originalMember = originalNodes.get(enumeration).members.find(node => node.name?.getText(original) === member);
   const expected = checker.getTypeAtLocation(originalMember).value;
   const actual = fixture.get(enumeration)?.members.find(node => node.name?.getText(source) === member)?.initializer;
   if (typeof expected !== 'number' || !actual || !ts.isNumericLiteral(actual) || Number(actual.text) !== expected) throw Error(name+': original enum value changed '+enumeration+'.'+member);
  }
  count++;
 }
}
console.log('Verified '+count+' fixtures retain requested original aliases, discriminators and enum values.');
