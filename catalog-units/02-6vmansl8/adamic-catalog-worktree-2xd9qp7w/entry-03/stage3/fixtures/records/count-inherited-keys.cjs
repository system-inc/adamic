// Census of original compiler code, using stock TypeScript as the parser/type oracle.
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const ts = require(process.env.CENSUS_TYPESCRIPT);
const source = path.resolve(process.argv[2]);
const output = path.resolve(process.argv[3]);
if (ts.version !== '6.0.3') throw Error('Expected TypeScript 6.0.3');
const proto = new Set(['constructor', 'toString', 'valueOf', 'hasOwnProperty', 'isPrototypeOf', 'propertyIsEnumerable', 'toLocaleString', '__proto__', '__defineGetter__', '__defineSetter__', '__lookupGetter__', '__lookupSetter__']);
const configFile = path.join(source, 'src/compiler/tsconfig.json');
const config = ts.readConfigFile(configFile, ts.sys.readFile);
if (config.error) throw Error('Cannot read upstream config');
const parsed = ts.parseJsonConfigFileContent(config.config, ts.sys, path.dirname(configFile), undefined, configFile);
if (parsed.errors.length) throw Error('Cannot parse upstream config');
parsed.options.typeRoots = [path.join(path.dirname(path.dirname(path.dirname(process.env.CENSUS_TYPESCRIPT))), '@types')];
const program = ts.createProgram(parsed.fileNames, parsed.options);
const checker = program.getTypeChecker();
const compiler = path.join(source, 'src/compiler/');
const files = program.getSourceFiles().filter(f => f.fileName.startsWith(compiler) && !f.fileName.includes('.generated.') && f.fileName.endsWith('.ts')).sort((a, b) => a.fileName.localeCompare(b.fileName));
const rows = [];
const loops = [];
const census = JSON.parse(fs.readFileSync(path.join(__dirname, 'inherited-key-ledger.json'), 'utf8'));
const censusByLocation = new Map(census.map((row, index) => [row.file + ':' + row.start + ':' + row.end, index]));
function loc(file, node) {
    const start = node.getStart(file);
    const p = file.getLineAndCharacterOfPosition(start);
    return {file: path.relative(source, file.fileName), line: p.line + 1, column: p.character + 1, start, end: node.end};
}
function unwrap(node) {
    while (ts.isParenthesizedExpression(node) || ts.isAsExpression(node) || ts.isTypeAssertionExpression(node) || ts.isNonNullExpression(node)) node = node.expression;
    return node;
}
function literal(node) {
    node = unwrap(node);
    return ts.isStringLiteralLike(node) || ts.isNumericLiteral(node) || (ts.isPrefixUnaryExpression(node) && ts.isNumericLiteral(node.operand));
}
function role(node) {
    let outer = node;
    while (ts.isParenthesizedExpression(outer.parent) || ts.isNonNullExpression(outer.parent) || ts.isAsExpression(outer.parent)) outer = outer.parent;
    const parent = outer.parent;
    if (ts.isDeleteExpression(parent)) return 'delete';
    if (ts.isBinaryExpression(parent) && parent.left === outer && parent.operatorToken.kind >= ts.SyntaxKind.FirstAssignment && parent.operatorToken.kind <= ts.SyntaxKind.LastAssignment) return parent.operatorToken.kind === ts.SyntaxKind.EqualsToken ? 'write' : 'read_write';
    if ((ts.isPrefixUnaryExpression(parent) || ts.isPostfixUnaryExpression(parent)) && [ts.SyntaxKind.PlusPlusToken, ts.SyntaxKind.MinusMinusToken].includes(parent.operator)) return 'read_write';
    if (ts.isForInStatement(parent) && parent.initializer === outer || ts.isForOfStatement(parent) && parent.initializer === outer) return 'write';
    // Destructuring targets are writes, but receiver/key subexpressions remain reads.
    let pattern = outer;
    while (ts.isArrayLiteralExpression(pattern.parent) || ts.isSpreadElement(pattern.parent) || ts.isPropertyAssignment(pattern.parent) && pattern.parent.initializer === pattern || ts.isObjectLiteralExpression(pattern.parent)) pattern = pattern.parent;
    if (ts.isBinaryExpression(pattern.parent) && pattern.parent.left === pattern && pattern.parent.operatorToken.kind === ts.SyntaxKind.EqualsToken) return 'write';
    return 'read';
}
function keyDomain(node) {
    let t = checker.getTypeAtLocation(node);
    const constraint = checker.getBaseConstraintOfType(t);
    if (constraint) t = constraint;
    const members = t.isUnion() ? t.types : [t];
    if (members.every(m => m.flags & (ts.TypeFlags.NumberLike | ts.TypeFlags.BigIntLike | ts.TypeFlags.ESSymbolLike))) return {classification: 'fixed_set_excludes_prototype', evidence: 'Non-string key domain cannot stringify to an Object.prototype member.', string_capable: false};
    if (members.every(m => m.flags & (ts.TypeFlags.StringLiteral | ts.TypeFlags.NumberLiteral))) {
        const values = members.map(m => String(m.value));
        if (values.every(v => !proto.has(v))) return {classification: 'fixed_set_excludes_prototype', evidence: 'Finite literal key set excludes all 12 prototype names: ' + values.join(', '), string_capable: members.some(m => m.flags & ts.TypeFlags.StringLiteral)};
    }
    return {classification: 'unknown', evidence: 'Unrestricted or unresolved key provenance; no claim that an own-key guard excludes the name.', string_capable: true};
}
function helperName(node) {
    let expression = unwrap(node.expression);
    let symbol = checker.getSymbolAtLocation(ts.isPropertyAccessExpression(expression) ? expression.name : expression);
    if (symbol?.flags & ts.SymbolFlags.Alias) symbol = checker.getAliasedSymbol(symbol);
    const declaration = symbol?.declarations?.find(d => ts.isFunctionDeclaration(d) && d.getSourceFile().fileName === path.join(compiler, 'core.ts') && ['hasProperty', 'getProperty'].includes(d.name?.text));
    return declaration?.name.text;
}
function add(file, node, kind, key, receiver, accessRole, loopStack) {
    const location = loc(file, node);
    const locationId = location.file + ':' + location.line + ':' + location.column;
    const id = locationId + '@' + node.end;
    const domain = keyDomain(key);
    const row = {...location, id, location_id: locationId, kind, role: accessRole, dynamic: !literal(key), expression: node.getText(file), key: key.getText(file), receiver: receiver.getText(file), key_type: checker.typeToString(checker.getTypeAtLocation(key)), receiver_type: checker.typeToString(checker.getTypeAtLocation(receiver)), ...domain, for_in: loopStack.map(l => l.id), census_index: censusByLocation.get(location.file + ':' + location.start + ':' + location.end) ?? null};
    row.counted = (!['element', 'in'].includes(kind) || row.dynamic) && ['read', 'read_write', 'membership'].includes(accessRole);
    rows.push(row);
}
for (const file of files) {
    function visit(node, loopStack = []) {
        if (ts.isForInStatement(node)) {
            const location = loc(file, node);
            const loop = {...location, id: location.file + ':' + location.line + ':' + location.column, expression: node.expression.getText(file), initializer: node.initializer.getText(file)};
            loops.push(loop);
            visit(node.initializer, loopStack);
            visit(node.expression, loopStack);
            visit(node.statement, [...loopStack, loop]);
            return;
        }
        if (ts.isElementAccessExpression(node)) add(file, node, 'element', node.argumentExpression, node.expression, role(node), loopStack);
        if (ts.isCallExpression(node)) {
            const name = helperName(node);
            if (name && node.arguments.length >= 2) add(file, node, name, node.arguments[1], node.arguments[0], name === 'hasProperty' ? 'membership' : 'read', loopStack);
            // Include direct own-property calls too, so a for-in guard is not hidden.
            if (ts.isPropertyAccessExpression(node.expression) && node.expression.name.text === 'call' && node.arguments.length >= 2 && /hasOwnProperty$/.test(node.expression.expression.getText(file))) add(file, node, 'hasOwnProperty.call', node.arguments[1], node.arguments[0], 'membership', loopStack);
            if (ts.isPropertyAccessExpression(node.expression) && node.expression.expression.getText(file) === 'Object' && node.expression.name.text === 'hasOwn' && node.arguments.length >= 2) add(file, node, 'Object.hasOwn', node.arguments[1], node.arguments[0], 'membership', loopStack);
        }
        if (ts.isBinaryExpression(node) && node.operatorToken.kind === ts.SyntaxKind.InKeyword) add(file, node, 'in', node.left, node.right, 'membership', loopStack);
        ts.forEachChild(node, child => visit(child, loopStack));
    }
    visit(file);
}
const matched = new Set(rows.filter(r => r.census_index !== null).map(r => r.census_index));
const missing = census.map((_, i) => i).filter(i => !matched.has(i));
if (missing.length) throw Error('Missing original string-index census sites: ' + missing);
fs.mkdirSync(output, {recursive: true});
fs.writeFileSync(path.join(output, 'raw.json'), JSON.stringify({typescript: ts.version, files: files.map(f => ({file: path.relative(source, f.fileName), sha256: crypto.createHash('sha256').update(fs.readFileSync(f.fileName)).digest('hex')})), sites: rows, loops, diagnostics: ts.getPreEmitDiagnostics(program).map(d => ({code: d.code, file: d.file && path.relative(source, d.file.fileName), message: ts.flattenDiagnosticMessageText(d.messageText, '\n')}))}, null, 2) + '\n');
console.log(JSON.stringify({files: files.length, sites: rows.length, counted: rows.filter(r => r.counted).length, string_capable: rows.filter(r => r.counted && r.string_capable).length, loops: loops.length, census_sites: matched.size}));
