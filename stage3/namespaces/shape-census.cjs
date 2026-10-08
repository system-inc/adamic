// Isolate namespace declaration structure with the independent stock TypeScript AST.
// Bodies, external types and enum expressions are replaced, not claimed to compile.
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const ts = require(process.env.NAMESPACE_TYPESCRIPT);
const root = process.argv[2];
const destination = process.argv[3];
fs.mkdirSync(destination, { recursive: true });
const rows = [];
const exported = node => node.modifiers?.some(m => m.kind === ts.SyntaxKind.ExportKeyword) ? 'export ' : '';
const generics = node => node.typeParameters?.length ? '<' + node.typeParameters.map(p => p.name.text).join(',') + '>' : '';
function eagerCall(node) {
 if (ts.isFunctionLike(node)) return false;
 if (ts.isCallExpression(node) || ts.isNewExpression(node)) return true;
 let found = false;
 ts.forEachChild(node, child => { if (eagerCall(child)) found = true; });
 return found;
}
function bindingValue(name) {
 if (ts.isObjectBindingPattern(name)) return '{' + name.elements.map(e => (e.propertyName || e.name).getText() + ':1').join(',') + '}';
 if (ts.isArrayBindingPattern(name)) return '[' + name.elements.map(() => '1').join(',') + ']';
 return '1';
}
function shape(node) {
 const prefix = exported(node);
 if (ts.isModuleDeclaration(node)) return prefix + 'namespace ' + node.name.text + ' {\n' + node.body.statements.map(shape).join('\n') + '\n}';
 if (ts.isFunctionDeclaration(node)) return prefix + 'function ' + node.name.text + generics(node) + '(): number' + (node.body ? ' {return 0;}' : ';');
 if (ts.isInterfaceDeclaration(node)) return prefix + 'interface ' + node.name.text + generics(node) + ' {}';
 if (ts.isTypeAliasDeclaration(node)) return prefix + 'type ' + node.name.text + generics(node) + ' = number;';
 if (ts.isClassDeclaration(node)) return prefix + 'class ' + node.name.text + generics(node) + ' {}';
 if (ts.isEnumDeclaration(node)) {
  const constant = node.modifiers?.some(m => m.kind === ts.SyntaxKind.ConstKeyword) ? 'const ' : '';
  return prefix + constant + 'enum ' + node.name.text + ' {' + node.members.map((m, i) => m.name.getText() + '=' + i).join(',') + '}';
 }
 if (ts.isVariableStatement(node)) {
  const kind = node.declarationList.flags & ts.NodeFlags.Const ? 'const' : node.declarationList.flags & ts.NodeFlags.Let ? 'let' : 'var';
  return prefix + kind + ' ' + node.declarationList.declarations.map(d => {
   const name = d.name.getText();
   if (!ts.isIdentifier(d.name)) return name + '=' + bindingValue(d.name);
   if (!d.initializer) return name + ':number';
   let value = d.initializer;
   while (ts.isAsExpression(value) || ts.isParenthesizedExpression(value) || ts.isTypeAssertionExpression(value)) value = value.expression;
   if (ts.isStringLiteral(value)) return name + ':string=' + JSON.stringify(value.text);
   if (value.kind === ts.SyntaxKind.TrueKeyword || value.kind === ts.SyntaxKind.FalseKeyword) return name + ':boolean=' + value.getText();
   return name + ':number=' + (eagerCall(d.initializer) ? 'shapeInitialize()' : '1');
  }).join(',') + ';';
 }
 throw new Error('Unrepresented declaration: ' + ts.SyntaxKind[node.kind]);
}
for (const file of ['builderState.ts', 'checker.ts', 'debug.ts', 'factory/utilities.ts', 'parser.ts', 'tracing.ts']) {
 const text = fs.readFileSync(path.join(root, file), 'utf8');
 const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
 function visit(node, ancestors = []) {
  if (ts.isModuleDeclaration(node) && node.body && ts.isModuleBlock(node.body)) {
   const names = [...ancestors, node.name.text];
   const name = names.join('.');
   let isolated = shape(node);
   for (let i = ancestors.length - 1; i >= 0; i--) {
    const merged = name === 'Debug.log' && i === 0 ? 'export function log():number{return 0;}\n' : '';
    isolated = 'namespace ' + ancestors[i] + ' {\n' + merged + isolated + '\n}';
   }
   if (name === 'BuilderState') isolated = 'interface BuilderState {}\n' + isolated;
   if (name === 'BinaryExpressionState') isolated = 'type BinaryExpressionState<T> = (state:T)=>void;\n' + isolated;
   if (name === 'tracingEnabled') isolated += '\nconst tracing=tracingEnabled;';
   isolated = '// Declaration shape only: external types, bodies and enum expressions are normalized.\nfunction shapeInitialize():number{return 1;}\n' + isolated + '\nconsole.log(' + JSON.stringify('shape:' + name) + ');\n';
   fs.writeFileSync(path.join(destination, name + '.a'), isolated);
   const counts = {};
   for (const member of node.body.statements) {
    const kind = ts.isVariableStatement(member) ? 'VariableStatement' : ts.SyntaxKind[member.kind];
    counts[kind] = (counts[kind] || 0) + 1;
   }
   rows.push({ name, file, line: source.getLineAndCharacterOfPosition(node.getStart(source)).line + 1, sha256: crypto.createHash('sha256').update(text).digest('hex'), counts,
    overloads: node.body.statements.filter(m => ts.isFunctionDeclaration(m) && !m.body).length,
    genericFunctions: node.body.statements.filter(m => ts.isFunctionDeclaration(m) && m.typeParameters?.length).length });
   for (const member of node.body.statements) if (ts.isModuleDeclaration(member)) visit(member, names);
   return;
  }
  ts.forEachChild(node, child => visit(child, ancestors));
 }
 visit(source);
}
if (rows.length !== 10) throw new Error('Expected the ten runtime namespace declarations, got ' + rows.length);
fs.writeFileSync(path.join(destination, 'census.json'), JSON.stringify({ typescript: ts.version, sourceCommit: '050880ce59e30b356b686bd3144efe24f875ebc8', normalization: 'Names, export flags, declaration kinds, var/let/const, absence of initializers, eager initializer calls, binding patterns, generic arity, overload signatures and nesting retained. External types, executable function bodies and enum expressions replaced. Tracing alias retained.', rows }, null, 2) + '\n');
