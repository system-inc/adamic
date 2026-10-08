// Observe initializers and interface contracts without equating a named candidate
// with a proven runtime origin. Read the compiler through its cohere submodule.
const ts = require(process.env.METHOD_VALUES_TYPESCRIPT || 'typescript');
const fs = require('fs'), path = require('path');
const root = path.resolve(__dirname, '../../../cohere/TypeScript/tsc/testdata/fixtures/compiler');
const files = [];
function collect(directory) {
  for (const entry of fs.readdirSync(directory, {withFileTypes:true})) {
    const file = path.join(directory, entry.name);
    if (entry.isDirectory()) collect(file);
    else if (file.endsWith('.ts')) files.push(file);
  }
}
collect(root);
const program = ts.createProgram(files, {target:ts.ScriptTarget.ESNext, module:ts.ModuleKind.NodeNext, moduleResolution:ts.ModuleResolutionKind.NodeNext, skipLibCheck:true});
const checker = program.getTypeChecker();
function location(node) {
  const file = node.getSourceFile();
  return {file:path.relative(root, file.fileName), line:file.getLineAndCharacterOfPosition(node.getStart(file)).line+1, kind:ts.SyntaxKind[node.kind]};
}
function readsThis(body) {
  let result = false;
  function visit(node) {
    if (node.kind === ts.SyntaxKind.ThisKeyword) result = true;
    if (node !== body && ts.isFunctionLike(node) && !ts.isArrowFunction(node)) return;
    ts.forEachChild(node, visit);
  }
  visit(body);
  return result;
}
function initialBody(node, seen = new Set()) {
  if (!node || seen.has(node)) return [];
  seen.add(node);
  if (node.body) return [{...location(node), readsThis:readsThis(node.body)}];
  if (ts.isShorthandPropertyAssignment(node)) {
    const symbol = checker.getShorthandAssignmentValueSymbol(node);
    return (symbol?.declarations || []).flatMap(value => initialBody(value, seen));
  }
  if (node.initializer) return initialBody(node.initializer, seen);
  if (ts.isIdentifier(node)) {
    return (checker.getSymbolAtLocation(node)?.declarations || []).flatMap(value => initialBody(value, seen));
  }
  return [];
}
const census = JSON.parse(fs.readFileSync(path.join(__dirname, 'census.json')));
const rows = census.filter(row => row.classification === 'unresolved runtime origin').map(row => {
  const file = program.getSourceFile(path.join(root, row.file));
  const position = file.getPositionOfLineAndCharacter(row.line-1, row.column-1);
  let access;
  function find(node) {
    if (ts.isPropertyAccessExpression(node) && node.getStart(file) === position && node.name.text === row.name) access = node;
    ts.forEachChild(node, find);
  }
  find(file);
  if (!access) throw Error(`missing site ${row.file}:${row.line}`);
  const declarations = checker.getSymbolAtLocation(access.name)?.declarations || [];
  const receiver = access.expression;
  const origins = ts.isIdentifier(receiver) ? checker.getSymbolAtLocation(receiver)?.declarations || [] : [];
  let bodies = [];
  for (const origin of origins) {
    const initializer = origin.initializer;
    if (initializer && ts.isObjectLiteralExpression(initializer)) {
      const member = initializer.properties.find(property => property.name?.getText(file) === row.name);
      bodies.push(...initialBody(member));
    }
  }
  return {
    file:row.file, line:row.line, column:row.column, expression:row.expression,
    observation:bodies.length ? 'local object initializer bodies' : 'runtime contract without a closed initializer',
    propertyDeclarations:declarations.map(location),
    receiverDeclarations:origins.map(origin => ({...location(origin), initializer:origin.initializer ? location(origin.initializer) : undefined})),
    initialBodies:bodies, namedCandidates:row.candidates,
    provenRuntimeThisFree:false,
    limitation:bodies.length ? 'Initial bodies are this-free, but the program object escapes; no whole-program nonreplacement proof is claimed.' : 'Interface contracts and named candidates do not determine every supplied implementation. Preserve the .a refusal unless the actual program closes the origin.'
  };
});
if (rows.length !== 43) throw Error(`expected 43 sites, got ${rows.length}`);
const local = rows.filter(row => row.initialBodies.length);
if (local.length !== 6 || local.some(row => row.initialBodies.some(body => body.readsThis))) throw Error('local program initializer observations changed');
fs.writeFileSync(path.join(__dirname, 'unresolved-audit.json'), JSON.stringify(rows,null,2)+'\n');
console.log(`43 sites audited: ${local.length} this-free local initializations; ${rows.length-local.length} runtime contracts. Proven census remains 60 this-free, 5 this-reading, 43 unresolved runtime origins.`);
