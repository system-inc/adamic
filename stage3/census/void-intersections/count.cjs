#!/usr/bin/env node
// Usage: node count.cjs ADAPTED_TREE OUTPUT.jsonl.gz
// Requires the repository's lockfile-pinned stage3/api install.
const fs = require('node:fs');
const path = require('node:path');
const zlib = require('node:zlib');
const ts = require('../../api/node_modules/typescript');
if (ts.version !== '6.0.3') throw new Error(`Expected TypeScript 6.0.3, got ${ts.version}`);
const [treeArg, output] = process.argv.slice(2);
if (!treeArg || !output) throw new Error('Usage: node count.cjs ADAPTED_TREE OUTPUT.jsonl.gz');
const tree = path.resolve(treeArg);
const configPath = path.join(tree, 'src/compiler/tsconfig.json');
const config = ts.readConfigFile(configPath, ts.sys.readFile);
if (config.error) throw new Error(ts.flattenDiagnosticMessageText(config.error.messageText, '\n'));
// stage3/census/tool -> load.Load -> internal/load/load.go:compilerOptions.
const overrides = {
  strict: true, strictBindCallApply: true, useUnknownInCatchVariables: true, noUncheckedIndexedAccess: true, exactOptionalPropertyTypes: true,
  erasableSyntaxOnly: false, verbatimModuleSyntax: true, allowImportingTsExtensions: true,
  noEmit: true, emitDeclarationOnly: false, module: ts.ModuleKind.ESNext, moduleDetection: ts.ModuleDetectionKind.Force,
  moduleResolution: ts.ModuleResolutionKind.Bundler, target: ts.ScriptTarget.ES2024,
  lib: ['lib.es2024.d.ts'], types: [],
};
const parsed = ts.parseJsonConfigFileContent(config.config, ts.sys, path.dirname(configPath), overrides, configPath);
if (parsed.errors.length) throw new Error(ts.formatDiagnosticsWithColorAndContext(parsed.errors, {
  getCurrentDirectory: () => tree, getCanonicalFileName: f => f, getNewLine: () => '\n',
}));
const program = ts.createProgram({rootNames: parsed.fileNames, options: parsed.options});
const checker = program.getTypeChecker();
function matches(type) {
  if (type.flags & ts.TypeFlags.Union) return type.types.some(matches);
  return !!(type.flags & ts.TypeFlags.Intersection) && type.types.some(t => !!(t.flags & ts.TypeFlags.Void));
}
function voidIntersections(type) {
  if (type.flags & ts.TypeFlags.Union) return type.types.flatMap(voidIntersections);
  return matches(type) ? [type] : [];
}
function unparen(node) {
  while (node.parent && ts.isParenthesizedExpression(node.parent)) node = node.parent;
  return node;
}
function discarded(node) {
  const n = unparen(node), p = n.parent;
  return !!p && (ts.isExpressionStatement(p) && p.expression === n || ts.isVoidExpression(p) && p.expression === n);
}
function read(node) {
  if (discarded(node)) return false;
  const n = unparen(node), p = n.parent;
  if (p && ts.isBinaryExpression(p) && p.left === n &&
      p.operatorToken.kind >= ts.SyntaxKind.FirstAssignment && p.operatorToken.kind <= ts.SyntaxKind.LastAssignment) return false;
  if (p && ts.isCallExpression(p) && p.expression === n && discarded(p)) return false;
  return true;
}
function role(n) {
  const site = unparen(n), p = site.parent;
  let use;
  if (p && (ts.isCallExpression(p) || ts.isNewExpression(p)) && p.arguments?.includes(site)) use = 'argument';
  else if (p && (ts.isReturnStatement(p) || ts.isArrowFunction(p) && p.body === site)) use = 'return value';
  else if (p && (ts.isIfStatement(p) || ts.isWhileStatement(p) || ts.isDoStatement(p) || ts.isConditionalExpression(p)) && (p.expression === site || p.condition === site)) use = 'condition';
  else if (p && ts.isForStatement(p) && p.condition === site) use = 'condition';
  else if (p && ts.isVariableDeclaration(p) && p.initializer === site) use = 'initializer';
  else if (p && ts.isBinaryExpression(p)) use = `operand of ${ts.tokenToString(p.operatorToken.kind)}`;
  else if (p && ts.isCallExpression(p) && p.expression === site) use = 'callee';
  const shape = ts.isPropertyAccessExpression(n) ? 'property access result' : ts.isElementAccessExpression(n) ? 'element access result' : ts.isCallExpression(n) ? 'call result' : ts.isIdentifier(n) ? 'identifier' : ts.SyntaxKind[n.kind];
  return use ? `${shape}; ${use}` : shape;
}
const rows = [];
for (const source of program.getSourceFiles().filter(f => f.fileName.startsWith(path.join(tree, 'src/compiler') + path.sep) && f.fileName.endsWith('.ts')).sort((a,b) => a.fileName.localeCompare(b.fileName))) {
  function visit(n) {
    if (ts.isTypeNode(n)) return;
    // Declaration names, type syntax and the name token of a.b are not separate reads.
    const expression = ts.isExpressionNode(n) || ts.isIdentifier(n) && ts.isShorthandPropertyAssignment(n.parent) && n.parent.name === n;
    if (expression && !(ts.isPropertyAccessExpression(n.parent) && n.parent.name === n)) {
      const type = checker.getTypeAtLocation(n);
      // Parentheses are transparent to the lowering rule, not another read.
      if (matches(type) && !ts.isParenthesizedExpression(n) && read(n)) {
        const pos = source.getLineAndCharacterOfPosition(n.getStart(source));
        const symbol = checker.getSymbolAtLocation(ts.isPropertyAccessExpression(n) ? n.name : n);
        const declarations = symbol?.declarations || [];
        const brandDeclarations = declarations.filter(d => (ts.isPropertySignature(d) || ts.isPropertyDeclaration(d)) && d.name?.getText().startsWith('__'));
        rows.push({file: path.relative(tree, source.fileName).split(path.sep).join('/'), line: pos.line + 1, column: pos.character + 1,
          expression: n.getText(source).replace(/\s+/g, ' ').slice(0,80), type: checker.typeToString(type, n, ts.TypeFormatFlags.NoTruncation), role: role(n),
          brandMember: brandDeclarations.length > 0,
          voidIntersectionOrigins: voidIntersections(type).map(t => ({
            type: checker.typeToString(t, n, ts.TypeFormatFlags.NoTruncation),
            brandProperties: checker.getPropertiesOfType(t).filter(s => s.name.startsWith('__')).map(s => s.name),
          })),
          brandDeclarations: brandDeclarations.map(d => {const s = d.getSourceFile(); const p = s.getLineAndCharacterOfPosition(d.getStart()); return `${path.relative(tree,s.fileName)}:${p.line+1}:${p.character+1}`;})});
      }
    }
    ts.forEachChild(n, visit);
  }
  visit(source);
}
fs.writeFileSync(output, zlib.gzipSync(rows.map(r => JSON.stringify(r)).join('\n') + (rows.length ? '\n' : ''), {mtime: 0}));
const groups = new Map();
for (const r of rows) groups.set(r.type, (groups.get(r.type) || 0) + 1);
console.log(JSON.stringify({typescript:ts.version, sites:rows.length, files:new Set(rows.map(r=>r.file)).size, brandMembers:rows.filter(r=>r.brandMember).length, shapes:[...groups].sort((a,b)=>b[1]-a[1])},null,2));
