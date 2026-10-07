// Run with: node count-typescript.mjs <TypeScript v6.0.3 checkout> <typescript npm package directory>
// Instrument the stock checker in memory to expose its narrowing test. No substitute definition.
import fs from 'node:fs';
import path from 'node:path';
import Module from 'node:module';
const [root, packageRoot] = process.argv.slice(2).map(value => path.resolve(value));
const library = path.join(packageRoot, 'lib/typescript.js');
const original = fs.readFileSync(library, 'utf8');
const anchor = 'getTypeOfSymbolAtLocation: (symbol, locationIn) => {';
if (original.split(anchor).length !== 2) throw Error('checker export anchor changed');
const instrumented = new Module(library);
instrumented.filename = library;
instrumented.paths = Module._nodeModulePaths(path.dirname(library));
instrumented._compile(original.replace(anchor, '__adamicIsDiscriminantProperty: isDiscriminantProperty,\n    __adamicPropertyName: getAccessedPropertyName,\n    ' + anchor), library);
const ts = instrumented.exports;
if (ts.version !== '6.0.3') throw Error(`want 6.0.3, got ${ts.version}`);
const configPath = path.join(root, 'src/compiler/tsconfig.json');
const config = ts.readConfigFile(configPath, ts.sys.readFile);
const parsed = ts.parseJsonConfigFileContent(config.config, ts.sys, path.dirname(configPath));
const program = ts.createProgram(parsed.fileNames, { ...parsed.options, configFilePath: configPath, typeRoots: [path.join(root, "node_modules/@types")], noEmit: true });
const checker = program.getTypeChecker();
const sources = program.getSourceFiles().filter(source => source.fileName.startsWith(path.join(root, 'src/compiler/')) && !source.isDeclarationFile);
const fields = new Map();
const seen = new Set();
function note(type) {
 if (!type || seen.has(type)) return;
 seen.add(type);
 if (type.flags & ts.TypeFlags.Union) {
  for (const member of type.types) {
   for (const property of checker.getPropertiesOfType(member)) {
    const literal = checker.getTypeOfSymbol(property);
    if ((literal.flags & (ts.TypeFlags.StringLiteral | ts.TypeFlags.NumberLiteral | ts.TypeFlags.BooleanLiteral)) && checker.__adamicIsDiscriminantProperty(type, property.name)) {
     if (!fields.has(property.name)) fields.set(property.name, new Set());
     fields.get(property.name).add(member);
    }
   }
   note(member);
  }
 }
}
function visit(source, callback) {
 function walk(node) { callback(node); ts.forEachChild(node, walk); }
 walk(source);
}
for (const source of sources) visit(source, node => {
 if (ts.isExpressionNode(node) || ts.isTypeAliasDeclaration(node) || ts.isTypeReferenceNode(node) || ts.isUnionTypeNode(node)) note(checker.getTypeAtLocation(node));
});
const writes = [];
for (const source of sources) visit(source, node => {
 if (!ts.isPropertyAccessExpression(node) && !ts.isElementAccessExpression(node)) return;
 if (!ts.isAssignmentTarget(node)) return;
 const name = checker.__adamicPropertyName(node);
 if (!name || !fields.has(name)) return;
 let holder = checker.getTypeAtLocation(node.expression);
 if (holder.flags & ts.TypeFlags.TypeParameter) holder = checker.getBaseConstraintOfType(holder);
 if (!holder || holder.flags & (ts.TypeFlags.Any | ts.TypeFlags.Unknown | ts.TypeFlags.Never)) return;
 if (![...fields.get(name)].some(member => checker.isTypeAssignableTo(member, holder))) return;
 let owner = node.parent;
 while (owner && !ts.isFunctionLike(owner)) owner = owner.parent;
 if (owner && ts.isConstructorDeclaration(owner)) return;
 const position = source.getLineAndCharacterOfPosition(node.getStart(source));
 writes.push({ file: path.relative(root, source.fileName), line: position.line + 1, column: position.character + 1, field: name, view: checker.typeToString(holder), text: node.parent.getText(source) });
});
writes.sort((a,b) => a.file.localeCompare(b.file) || a.line-b.line || a.column-b.column);
const diagnostics = ts.getPreEmitDiagnostics(program);
console.log(JSON.stringify({ version: ts.version, sourceFiles: sources.length, diagnostics: diagnostics.map(d => ({file: d.file && path.relative(root, d.file.fileName), line: d.file && d.file.getLineAndCharacterOfPosition(d.start).line + 1, code:d.code, message:ts.flattenDiagnosticMessageText(d.messageText, ' ')})), count: writes.length, writes }, null, 2));
