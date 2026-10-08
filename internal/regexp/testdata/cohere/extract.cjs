// TypeScript syntax, not text searches, determines the inventory. Never execute source files.
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const cp = require('node:child_process');
const ts = require('typescript');
if (ts.version !== '6.0.3') throw Error('TypeScript 6.0.3 required');
const root = path.resolve(__dirname, '../../../..');
const output = __dirname;
const inventory = {version: 1, typescript: ts.version, roots: ['cohere', 'stage1/cohere'], files: [], patterns: [], dynamic: [], parseDiagnostics: []};
function walk(dir) {
  const result = [];
  for (const entry of fs.readdirSync(dir, {withFileTypes: true}).sort((a,b) => a.name.localeCompare(b.name, 'en'))) {
    if (entry.name === '.git' || entry.name === 'node_modules') continue;
    const name = path.join(dir, entry.name);
    if (entry.isDirectory()) result.push(...walk(name));
    else if (entry.isFile() && /\.(?:ts|tsx|mts|cts)$/.test(name)) result.push(name);
  }
  return result;
}
const files = inventory.roots.flatMap(dir => walk(path.join(root, dir))).sort();
for (const file of files) {
  const relative = path.relative(root, file).split(path.sep).join('/');
  const text = fs.readFileSync(file, 'utf8');
  inventory.files.push({file: relative, sha256: crypto.createHash('sha256').update(text).digest('hex')});
  const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
  function location(node) {
    const at = source.getLineAndCharacterOfPosition(node.getStart(source));
    return {file: relative, line: at.line + 1, column: at.character + 1};
  }
  for (const d of source.parseDiagnostics) inventory.parseDiagnostics.push({...location({getStart: () => d.start || 0}), message: ts.flattenDiagnosticMessageText(d.messageText, '\n')});
  const regexNodes = [], calls = [];
  function find(node) {
    if (ts.isRegularExpressionLiteral(node) || ts.isNewExpression(node) && ts.isIdentifier(node.expression) && node.expression.text === 'RegExp') regexNodes.push(node);
    if (ts.isCallExpression(node)) calls.push(node);
  }
  const pending = [source];
  while (pending.length) {
    const node = pending.pop();
    find(node);
    ts.forEachChild(node, child => { pending.push(child); });
  }
  regexNodes.sort((a,b) => a.pos-b.pos);
  calls.sort((a,b) => a.pos-b.pos);
  if (regexNodes.length === 0) continue;
  // Bind only this file: unrelated compiler fixtures intentionally redeclare names.
  const host = ts.createCompilerHost({noLib:true, noResolve:true});
  host.getSourceFile = name => name === file ? source : undefined;
  const program = ts.createProgram([file], {noLib:true, noResolve:true}, host);
  const checker = program.getTypeChecker();
  const records = new Map();
  function constant(node, seen = new Set()) {
    if (!node || seen.has(node)) return undefined;
    seen = new Set(seen); seen.add(node);
    if (ts.isStringLiteral(node) || ts.isNoSubstitutionTemplateLiteral(node)) return node.text;
    if (ts.isNumericLiteral(node)) return Number(node.text);
    if (node.kind === ts.SyntaxKind.TrueKeyword) return true;
    if (node.kind === ts.SyntaxKind.FalseKeyword) return false;
    if (ts.isParenthesizedExpression(node) || ts.isAsExpression(node) || ts.isTypeAssertionExpression(node) || ts.isNonNullExpression(node)) return constant(node.expression, seen);
    if (ts.isBinaryExpression(node) && node.operatorToken.kind === ts.SyntaxKind.PlusToken) {
      const left = constant(node.left, seen), right = constant(node.right, seen);
      if (left !== undefined && right !== undefined) return left + right;
    }
    if (ts.isTemplateExpression(node)) {
      let value = node.head.text;
      for (const span of node.templateSpans) {
        const part = constant(span.expression, seen);
        if (part === undefined) return undefined;
        value += String(part) + span.literal.text;
      }
      return value;
    }
    if (ts.isCallExpression(node) && ts.isPropertyAccessExpression(node.expression) && ts.isIdentifier(node.expression.expression) && node.expression.expression.text === 'RegExp' && node.expression.name.text === 'escape') {
      const argument = constant(node.arguments[0], seen);
      if (typeof argument === 'string') return RegExp.escape(argument);
    }
    if (ts.isIdentifier(node)) {
      const symbol = checker.getSymbolAtLocation(node);
      const declaration = symbol?.valueDeclaration;
      if (declaration && ts.isVariableDeclaration(declaration) && declaration.parent.flags & ts.NodeFlags.Const) return constant(declaration.initializer, seen);
    }
    return undefined;
  }
  function resolveRegex(node, seen = new Set()) {
    if (!node || seen.has(node)) return undefined;
    if (records.has(node)) return records.get(node);
    seen.add(node);
    if (ts.isParenthesizedExpression(node)) return resolveRegex(node.expression, seen);
    if (ts.isIdentifier(node)) {
      const declaration = checker.getSymbolAtLocation(node)?.valueDeclaration;
      if (declaration && ts.isVariableDeclaration(declaration) && declaration.parent.flags & ts.NodeFlags.Const) return resolveRegex(declaration.initializer, seen);
    }
  }
  // Literals first, so RegExp(/literal/, flags) can be evaluated without source execution.
  const ordered = [...regexNodes.filter(ts.isRegularExpressionLiteral), ...regexNodes.filter(ts.isNewExpression)];
  for (const node of ordered) {
    let pattern, flags = '', kind, literalToken;
    if (ts.isRegularExpressionLiteral(node)) {
      const token = node.text, end = token.lastIndexOf('/');
      literalToken = token;
      pattern = end > 0 ? token.slice(1, end) : token.slice(1); flags = end > 0 ? token.slice(end + 1) : ''; kind = 'literal';
    } else {
      kind = 'constructor';
      const args = node.arguments || [];
      const literal = resolveRegex(args[0]);
      pattern = args.length === 0 ? '' : literal ? literal.pattern : constant(args[0]);
      if (typeof pattern !== 'string') pattern = undefined;
      flags = args.length < 2 ? literal?.flags || '' : constant(args[1]);
      if (typeof flags !== 'string') flags = undefined;
      if (pattern === undefined || flags === undefined) {
        inventory.dynamic.push({...location(node), expression:node.getText(source), reason:pattern === undefined ? 'nonconstant pattern' : 'nonconstant flags', constantPattern:pattern});
        continue;
      }
    }
    const record = {id: inventory.patterns.length, ...location(node), kind, literalToken, pattern, patternUnits:Array.from({length:pattern.length}, (_,i)=>pattern.charCodeAt(i)), flags, fixture: /(?:testdata|test|tests|baselines)\//.test(relative) || /\.test\.[cm]?tsx?$/.test(relative), inputs:[]};
    inventory.patterns.push(record); records.set(node, record);
  }
  for (const call of calls) {
    if (!ts.isPropertyAccessExpression(call.expression)) continue;
    const method = call.expression.name.text;
    let regex, input;
    if (method === 'exec' || method === 'test') {
      regex = resolveRegex(call.expression.expression); input = constant(call.arguments[0]);
    } else if (['match','matchAll','search','replace','replaceAll','split'].includes(method)) {
      regex = resolveRegex(call.arguments[0]); input = constant(call.expression.expression);
    }
    if (regex && typeof input === 'string') regex.inputs.push({units:Array.from({length:input.length}, (_,i)=>input.charCodeAt(i)), source:location(call), method, evidence:'constant receiver/argument at regex call; reachability not asserted'});
  }
}
inventory.revisions = {
  adamic: cp.execFileSync('git', ['rev-parse','HEAD'], {cwd:root,encoding:'utf8'}).trim(),
  cohere: cp.execFileSync('git', ['rev-parse','HEAD'], {cwd:path.join(root,'cohere'),encoding:'utf8'}).trim(),
  typescript: cp.execFileSync('git', ['rev-parse','HEAD'], {cwd:path.join(root,'cohere/TypeScript'),encoding:'utf8'}).trim(),
};
fs.writeFileSync(path.join(output, 'inventory.json'), JSON.stringify(inventory,null,2) + '\n');
console.log(JSON.stringify({files:files.length, patterns:inventory.patterns.length, literals:inventory.patterns.filter(p=>p.kind==='literal').length, constructors:inventory.patterns.filter(p=>p.kind==='constructor').length, dynamic:inventory.dynamic.length, fixtureInputs:inventory.patterns.reduce((n,p)=>n+p.inputs.length,0), parseDiagnostics:inventory.parseDiagnostics.length}));
