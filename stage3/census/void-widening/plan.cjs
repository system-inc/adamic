'use strict';
// Explore an annotation closure in memory. The stock checker validates the batch.
const fs = require('node:fs'), path = require('node:path'), crypto = require('node:crypto'), assert = require('node:assert/strict');
const ts = require('../../api/node_modules/typescript');
const {compiler, nodes} = require('../../adapt/47-host-errors/adapt.cjs');
const [treeArg, inventory, output] = process.argv.slice(2), tree = path.resolve(treeArg);
const program = compiler(tree, ts), baseline = ts.getPreEmitDiagnostics(program), checker = program.getTypeChecker();
const raw = JSON.parse(fs.readFileSync(inventory)), rows = raw.sites || raw;
for (const row of rows) if (!row.form) row.form = row.reason.startsWith('a void value') ? 'return-of-void' : 'relation';
const candidates = new Map(), roots = new Map();
const hash = text => crypto.createHash('sha256').update(text).digest('hex');
function ownReturns(d) {
    const found = [];
    function visit(n) {
        if (n !== d.body && ts.isFunctionLike(n)) return;
        if (ts.isReturnStatement(n) && n.expression) found.push(n.expression);
        ts.forEachChild(n, visit);
    }
    if (d.body && ts.isBlock(d.body)) visit(d.body);
    else if (d.body) found.push(d.body);
    return found;
}
function local(d) {
    const sf = d.getSourceFile();
    if (!sf.fileName.startsWith(tree + '/src/compiler/')) return false;
    for (let p = d.parent; p && p !== sf; p = p.parent) if (ts.isFunctionLike(p) && p.body) return true;
    return /@internal\b/.test(d.getFullText(sf).slice(0, d.getStart(sf) - d.getFullStart()));
}
function add(d) {
    if (!d || !local(d)) return null;
    // A full upstream build rejects the global noop retype at four harness consumers.
    if (d.getSourceFile().fileName.endsWith('/src/compiler/core.ts') && d.name?.getText() === 'noop') return null;
    if (candidates.has(d)) return candidates.get(d);
    if (!d.body && !ts.isFunctionTypeNode(d)) return null;
    if (!d.type && !d.body || d.type && d.type.kind !== ts.SyntaxKind.VoidKeyword) return null;
    const sf = d.getSourceFile(), close = d.getChildren(sf).find(n => n.kind === ts.SyntaxKind.CloseParenToken);
    if (!d.type && !close && !ownReturns(d).length) return null;
    const c = {d, dependencies: new Set(), valid: true, reason: null};
    candidates.set(d, c);
    for (const e of ownReturns(d)) dependency(e, c);
    return c;
}
function need(d, c) {
    const other = add(d);
    if (other) c.dependencies.add(other);
    else { c.valid = false; c.reason = 'A returned void value has no closed internal implementation that can be retyped.'; }
}
function callable(e, c) {
    while (ts.isParenthesizedExpression(e) || ts.isAsExpression(e)) {
        if (ts.isAsExpression(e) && ts.isFunctionTypeNode(e.type) && e.type.type.kind === ts.SyntaxKind.VoidKeyword) need(e.type, c);
        e = e.expression;
    }
    if (ts.isArrowFunction(e) || ts.isFunctionExpression(e)) { need(e, c); return; }
    const sigs = checker.getSignaturesOfType(checker.getTypeAtLocation(e), ts.SignatureKind.Call);
    if (!sigs.length) { c.valid = false; c.reason = 'No concrete callable signature.'; }
    for (const sig of sigs) need(sig.declaration, c);
}
function dependency(e, c) {
    const t = checker.getTypeAtLocation(e);
    if (t.flags === ts.TypeFlags.Undefined || t.flags === ts.TypeFlags.Never) return;
    if (ts.isCallExpression(e)) {
        if (t.flags === ts.TypeFlags.Void) { need(checker.getResolvedSignature(e)?.declaration, c); return; }
        // Generic iterator returns can become undefined after their callbacks do.
        if (t.isUnion() && t.types.every(x => x.flags & (ts.TypeFlags.Void | ts.TypeFlags.Undefined))) {
            for (const a of e.arguments) if (checker.getSignaturesOfType(checker.getTypeAtLocation(a), ts.SignatureKind.Call).length) callable(a, c);
            return;
        }
    }
    if (ts.isConditionalExpression(e)) { dependency(e.whenTrue, c); dependency(e.whenFalse, c); return; }
    if (ts.isBinaryExpression(e) && [ts.SyntaxKind.AmpersandAmpersandToken, ts.SyntaxKind.BarBarToken, ts.SyntaxKind.QuestionQuestionToken].includes(e.operatorToken.kind)) {
        dependency(e.left, c); dependency(e.right, c); return;
    }
    if (ts.isParenthesizedExpression(e)) { dependency(e.expression, c); return; }
    c.valid = false; c.reason = 'Return expression is not proven undefined: ' + e.getText().slice(0, 160);
}
for (const row of rows) {
    const sf = program.getSourceFile(path.join(tree, row.file)), at = sf.getPositionOfLineAndCharacter(row.line - 1, row.column - 1);
    const n = nodes(ts, sf, n => n.getStart(sf) === at && ts.isExpressionNode(n))[0];
    assert(n, JSON.stringify(row));
    const root = {dependencies: new Set(), valid: true};
    if (row.form === 'return-of-void') dependency(n, root); else callable(n, root);
    roots.set(row, root);
}
let changed;
do {
    changed = false;
    for (const c of [...candidates.values(), ...roots.values()]) if (c.valid && [...c.dependencies].some(d => !d.valid)) {
        c.valid = false; c.reason = 'A required return declaration cannot be retyped.'; changed = true;
    }
} while (changed);
const needed = new Set();
function include(c) { for (const dep of c.dependencies) if (!needed.has(dep)) { needed.add(dep); include(dep); } }
for (const root of roots.values()) if (root.valid) include(root);
const edits = new Map(), manifest = [];
for (const c of needed) {
    const d = c.d, sf = d.getSourceFile(), close = d.getChildren(sf).find(n => n.kind === ts.SyntaxKind.CloseParenToken);
    // Concise arrows retain their parameter spelling and infer undefined from the refined callees.
    if (!d.type && !close) continue;
    const start = d.type ? d.type.getStart(sf) : close.end, end = d.type ? d.type.end : close.end;
    const edit = {file: path.relative(tree, sf.fileName), line: sf.getLineAndCharacterOfPosition(d.getStart(sf)).line + 1,
        declarationStart: d.getStart(sf), declarationText: d.getText(sf), start, end, before: sf.text.slice(start, end), after: d.type ? 'undefined' : ': undefined'};
    manifest.push(edit);
    if (!edits.has(sf.fileName)) edits.set(sf.fileName, []);
    edits.get(sf.fileName).push(edit);
}
const overlay = new Map(), erasure = [];
for (const [file, changes] of edits) {
    const original = program.getSourceFile(file).text;
    let text = original;
    for (const e of changes.sort((a,b) => b.start - a.start)) text = text.slice(0, e.start) + e.after + text.slice(e.end);
    for (const removeComments of [false, true]) {
        const options = {fileName: file, compilerOptions: {target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, removeComments}};
        const a = ts.transpileModule(original, options).outputText, b = ts.transpileModule(text, options).outputText;
        assert.equal(a, b, file);
        erasure.push({file: path.relative(tree, file), removeComments, bytes: Buffer.byteLength(a), sha256: hash(a)});
    }
    overlay.set(file, text);
}
const host = ts.createCompilerHost(program.getCompilerOptions()), read = host.readFile;
host.readFile = name => overlay.get(name) ?? read(name);
const afterProgram = ts.createProgram(program.getRootFileNames(), program.getCompilerOptions(), host);
const key = d => JSON.stringify([d.file ? path.relative(tree, d.file.fileName) : null, d.code, ts.flattenDiagnosticMessageText(d.messageText, '\n')]);
const counts = new Map(); for (const d of baseline) counts.set(key(d), (counts.get(key(d)) || 0) + 1);
const after = ts.getPreEmitDiagnostics(afterProgram), introduced = [];
for (const d of after) { const k = key(d), n = counts.get(k) || 0; if (n) counts.set(k, n - 1); else introduced.push({key: k, start: d.start}); }
function editIds(root) {
    const all = new Set();
    function visit(c) { for (const dep of c.dependencies) if (!all.has(dep)) { all.add(dep); visit(dep); } }
    visit(root);
    return manifest.filter(e => [...all].some(c => path.relative(tree, c.d.getSourceFile().fileName) === e.file && c.d.getStart() === e.declarationStart)).map(e => e.file + ':' + e.start);
}
const afterChecker = afterProgram.getTypeChecker();
for (const [row, root] of roots) if (root.valid) {
    const sf = afterProgram.getSourceFile(path.join(tree, row.file));
    const originalSf = program.getSourceFile(path.join(tree, row.file));
    const originalAt = originalSf.getPositionOfLineAndCharacter(row.line - 1, row.column - 1);
    const delta = manifest.filter(e => e.file === row.file && e.end <= originalAt).reduce((n,e) => n + e.after.length - (e.end - e.start), 0);
    const expression = nodes(ts, sf, n => n.getStart(sf) === originalAt + delta && ts.isExpressionNode(n))[0];
    assert(expression, 'rewritten site missing');
    const type = afterChecker.getTypeAtLocation(expression);
    row.rewrittenType = afterChecker.typeToString(type);
    if (row.form === 'return-of-void') assert.equal(type.flags, ts.TypeFlags.Undefined, 'return did not become undefined');
    else assert(afterChecker.getSignaturesOfType(type, ts.SignatureKind.Call).every(s => afterChecker.getReturnTypeOfSignature(s).flags === ts.TypeFlags.Undefined), 'callback did not become undefined');
}
const result = {baselineDiagnostics: baseline.length, afterDiagnostics: after.length, introduced,
    edits: manifest, erasure, sites: [...roots].map(([row, root]) => ({...row, eligible: root.valid, reason: root.reason, editIds: editIds(root),
        edits: [...root.dependencies].map(c => ({file: path.relative(tree, c.d.getSourceFile().fileName), line: c.d.getSourceFile().getLineAndCharacterOfPosition(c.d.getStart()).line + 1}))}))};
fs.writeFileSync(output, JSON.stringify(result, null, 2) + '\n');
console.log(JSON.stringify({sites: rows.length, eligible: result.sites.filter(s => s.eligible).length, declarations: manifest.length, files: overlay.size, introduced}));
