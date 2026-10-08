// Classify the pinned census's nested declaration sites against this unit's
// structural restrictions. This is not a claim that any corpus file compiles.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT);
const root = path.resolve(process.argv[2]);
const configPath = path.join(root, 'src/compiler/tsconfig.json');
const configFile = ts.readConfigFile(configPath, ts.sys.readFile);
if (configFile.error) throw new Error(ts.flattenDiagnosticMessageText(configFile.error.messageText, '\n'));
const config = ts.parseJsonConfigFileContent(configFile.config, ts.sys, path.dirname(configPath));
if (config.errors.length) throw new Error('invalid upstream configuration');
const program = ts.createProgram(config.fileNames, config.options);
const checker = program.getTypeChecker();
const files = program.getSourceFiles().filter(file => file.fileName.startsWith(path.join(root, 'src/compiler') + path.sep));
const declarations = new Map();
const bySymbol = new Map();
function enclosing(node) {
 for (let parent = node.parent; parent; parent = parent.parent) {
  if (ts.isFunctionLike(parent)) return parent;
 }
 return undefined;
}
for (const file of files) {
 function collect(node) {
  if (ts.isFunctionDeclaration(node) && enclosing(node)) {
   const position = file.getLineAndCharacterOfPosition(node.getStart(file));
   const entry = {file: path.relative(root, file.fileName), line: position.line + 1, column: position.character + 1, reasons: []};
   declarations.set(node, entry);
   if (node.name) bySymbol.set(checker.getSymbolAtLocation(node.name), node);
   if (!ts.isBlock(node.parent) || !ts.isFunctionLike(node.parent.parent)) entry.reasons.push('block scoped');
   if (!node.name || node.typeParameters?.length) entry.reasons.push('generic or unnamed');
   if (node.parameters.some(parameter => !ts.isIdentifier(parameter.name) && (parameter.initializer || parameter.questionToken || parameter.dotDotDotToken))) entry.reasons.push('destructured parameter with modifiers');
   if (node.parameters.some(parameter => parameter.dotDotDotToken)) entry.reasons.push('rest parameter');
   if (node.parameters.some(parameter => ts.isIdentifier(parameter.name) && parameter.name.text === 'this')) entry.reasons.push('dynamic this');
  }
  ts.forEachChild(node, collect);
 }
 collect(file);
}
const references = [];
for (const file of files) {
 function inspect(node) {
  if (ts.isTypeNode(node)) return;
  if (node.kind === ts.SyntaxKind.ThisKeyword) {
   // An arrow inherits this from its nearest ordinary function.
   let owner = enclosing(node);
   while (owner && ts.isArrowFunction(owner)) owner = enclosing(owner);
   const entry = declarations.get(owner);
   if (entry && !entry.reasons.includes('dynamic this')) entry.reasons.push('dynamic this');
  }
  if (ts.isIdentifier(node) && !ts.isDeclarationName(node)) {
   const target = bySymbol.get(ts.isShorthandPropertyAssignment(node.parent) ? checker.getShorthandAssignmentValueSymbol(node.parent) : checker.getSymbolAtLocation(node));
   const current = enclosing(node);
   if (target && current && current !== enclosing(target)) {
    let callee = node;
    while (ts.isParenthesizedExpression(callee.parent)) callee = callee.parent;
    const direct = ts.isCallExpression(callee.parent) && callee.parent.expression === callee;
    const sibling = declarations.has(current) && enclosing(current) === enclosing(target);
    if (!(direct && sibling)) {
     const position = file.getLineAndCharacterOfPosition(node.getStart(file));
     references.push({file: path.relative(root, file.fileName), line: position.line + 1, column: position.character + 1, reason: 'first-class or ancestor-group reference'});
    }
   }
  }
  ts.forEachChild(node, inspect);
 }
 inspect(file);
}
const sites = [...declarations.values()];
// Optional independent census inventory must identify precisely the same sites.
if (process.argv[3]) {
 const expected = JSON.parse(fs.readFileSync(process.argv[3], 'utf8'))
  .filter(site => site.reason === 'a function inside a function (a closure)');
 const key = site => site.file + ':' + site.line + ':' + site.column;
 const actualKeys = sites.map(key).sort();
 const expectedKeys = expected.map(key).sort();
 if (JSON.stringify(actualKeys) !== JSON.stringify(expectedKeys)) throw new Error('nested census site mismatch');
}
const affected = [...new Set(sites.map(site => site.file))].sort();
const blocked = new Set([...sites.filter(site => site.reasons.length).map(site => site.file), ...references.map(site => site.file)]);
const cleared = affected.filter(file => !blocked.has(file));
const reasons = {};
for (const site of sites) for (const reason of site.reasons) reasons[reason] = (reasons[reason] || 0) + 1;
console.log(JSON.stringify({definition: 'Structural declaration and reference restrictions only; excludes representation, cycle, other language and checker blockers', before: {sites: sites.length, lines: new Set(sites.map(site => site.file + ':' + site.line)).size, files: affected.length}, after: {blocked_declarations: sites.filter(site => site.reasons.length).length, blocked_reference_sites: references.length, files_with_remaining_restrictions: affected.length - cleared.length, files_without_these_restrictions: cleared.length}, reasons, cleared_files: cleared, blocked_sites: sites.filter(site => site.reasons.length), blocked_references: references}, null, 2));
