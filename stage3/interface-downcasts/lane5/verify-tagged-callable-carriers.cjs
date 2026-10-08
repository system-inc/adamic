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
  if (ts.isTypeAliasDeclaration(node) || ts.isInterfaceDeclaration(node)) nodes.set(node.name.text, node);
  ts.forEachChild(node, visit);
 }
 visit(source);
 return nodes;
}
for (const [name, node] of collect(original)) originalNodes.set(name, node);
const families = new Map([
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
 for (const name of fs.readdirSync(directory).filter(name => name.endsWith('.a'))) {
  const source = ts.createSourceFile(name, fs.readFileSync(path.join(directory,name),'utf8'), ts.ScriptTarget.Latest, true);
  const fixture = collect(source);
  for (const alias of family.aliases) {
   if (!fixture.has(alias) || normalized(fixture.get(alias).getText(source)) !== normalized(originalNodes.get(alias).getText(original))) throw Error(name+': original alias changed '+alias);
  }
  for (const carrier of family.carriers) {
   const originalKind = originalNodes.get(carrier).members.find(member => member.name?.getText(original) === 'kind');
   const fixtureKind = fixture.get(carrier)?.members.find(member => member.name?.getText(source) === 'kind');
   const expected = checker.getTypeAtLocation(originalKind.type).value;
   if (typeof expected !== 'number' || !fixtureKind || !ts.isLiteralTypeNode(fixtureKind.type) || !ts.isNumericLiteral(fixtureKind.type.literal) || Number(fixtureKind.type.literal.text) !== expected || !fixtureKind.modifiers?.some(modifier => modifier.kind === ts.SyntaxKind.ReadonlyKeyword)) throw Error(name+': original discriminator changed '+carrier);
  }
  count++;
 }
}
console.log('Verified '+count+' fixtures retain original aliases and numeric kind discriminators.');
