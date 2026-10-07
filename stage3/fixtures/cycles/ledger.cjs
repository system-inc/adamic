// Trace imported value reads during ESM module evaluation, using stock TypeScript's API.
// Usage: NODE_PATH=<typescript 6.0.3 node_modules> node ledger.cjs <upstream> <scratch>
const ts = require('typescript');
const fs = require('node:fs');
const path = require('node:path');
const cp = require('node:child_process');
const assert = require('node:assert/strict');
assert.equal(ts.version, '6.0.3');
const root = path.resolve(process.argv[2]);
const out = path.resolve(process.argv[3]);
assert(!fs.existsSync(out), 'use a new scratch directory');
const entry = path.join(root, 'src/tsc/tsc.ts');
const program = ts.createProgram([entry], {
    target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext,
    moduleResolution: ts.ModuleResolutionKind.Bundler,
    rootDir: root, outDir: out, noEmitOnError: false, skipLibCheck: true,
});
const checker = program.getTypeChecker();
const metadata = [];
const relative = f => path.relative(root, f).replaceAll('\\', '/');
function location(n) {
    const sf = n.getSourceFile();
    const p = sf.getLineAndCharacterOfPosition(n.getStart(sf));
    return `${relative(sf.fileName)}:${p.line + 1}:${p.character + 1}`;
}
function ultimate(s) {
    return s && s.flags & ts.SymbolFlags.Alias ? checker.getAliasedSymbol(s) : s;
}
function constEnum(s) {
    return s && s.flags & ts.SymbolFlags.ConstEnum;
}
function provider(s) {
    s = ultimate(s);
    const d = s && (s.valueDeclaration || s.declarations?.[0]);
    return d ? { provider: relative(d.getSourceFile().fileName), declaration: location(d),
        hoisted: ts.isFunctionDeclaration(d), name: s.name } : undefined;
}
function imported(n) {
    const s = checker.getSymbolAtLocation(n);
    return s?.declarations?.some(d => ts.isImportSpecifier(d) || ts.isNamespaceImport(d) || ts.isImportClause(d))
        && !s.declarations.some(d => ts.isVariableDeclaration(d) || ts.isParameter(d) || ts.isFunctionDeclaration(d));
}
function valueReference(n) {
    const p = n.parent;
    if (ts.isBinaryExpression(p) && p.left === n && p.operatorToken.kind >= ts.SyntaxKind.FirstAssignment
        && p.operatorToken.kind <= ts.SyntaxKind.LastAssignment) return false;
    if ((ts.isPrefixUnaryExpression(p) || ts.isPostfixUnaryExpression(p))
        && (p.operator === ts.SyntaxKind.PlusPlusToken || p.operator === ts.SyntaxKind.MinusMinusToken)) return false;
    if (ts.isPropertyAccessExpression(p) && p.name === n) return false;
    if ((ts.isPropertyAssignment(p) || ts.isMethodDeclaration(p)) && p.name === n) return false;
    if (ts.isExportSpecifier(p) || ts.isImportSpecifier(p) || ts.isNamespaceImport(p)) return false;
    if (p.name === n && !ts.isShorthandPropertyAssignment(p)) return false;
    return true;
}
function instrument(context) {
    const f = context.factory;
    const global = name => f.createPropertyAccessExpression(f.createIdentifier('globalThis'), name);
    const call = (name, args) => f.createCallExpression(global(name), undefined, args);
    const statement = (name, value) => f.createExpressionStatement(call(name, [f.createStringLiteral(value)]));
    const read = (node, binding, s) => {
        const info = provider(s);
        if (!info || !info.provider.startsWith('src/compiler/')) return node;
        const id = metadata.length;
        let ancestor = node;
        let direct = true;
        while (ancestor.parent && !ts.isSourceFile(ancestor.parent)) {
            if (ts.isFunctionLike(ancestor)) direct = false;
            ancestor = ancestor.parent;
        }
        if (ts.isFunctionLike(ancestor)) direct = false;
        metadata.push({ location: location(node), binding, directStatement: direct ? location(ancestor) : null, ...info });
        return f.createParenthesizedExpression(f.createCommaListExpression([
            call('__cycleRead', [f.createNumericLiteral(id)]), node,
        ]));
    };
    function visit(n) {
        if (ts.isTypeNode(n) || ts.isInterfaceDeclaration(n) || ts.isTypeAliasDeclaration(n)
            || ts.isImportDeclaration(n) || ts.isExportDeclaration(n)) return n;
        // Keep stock compiler constant folding intact. Const enums have no load-time binding read.
        if (ts.isPropertyAccessExpression(n)) {
            const member = ultimate(checker.getSymbolAtLocation(n.name));
            const receiver = ultimate(checker.getSymbolAtLocation(n.expression));
            if (constEnum(receiver) || member?.declarations?.some(d => ts.isEnumMember(d) &&
                d.parent.modifiers?.some(m => m.kind === ts.SyntaxKind.ConstKeyword))) return n;
            if (ts.isIdentifier(n.expression) && imported(n.expression)) {
                const sym = checker.getSymbolAtLocation(n.expression);
                const namespace = sym?.declarations?.some(d => ts.isNamespaceImport(d));
                // Preserve member-call receivers and writable property references.
                const receiver = read(n.expression, namespace ? n.getText() : n.expression.text,
                    namespace ? checker.getSymbolAtLocation(n.name) : sym);
                return f.updatePropertyAccessExpression(n, receiver, n.name);
            }
        }
        if (ts.isShorthandPropertyAssignment(n)) {
            const s = checker.getShorthandAssignmentValueSymbol(n);
            if (s?.declarations?.some(d => ts.isImportSpecifier(d))) {
                return f.createPropertyAssignment(n.name, read(n.name, n.name.text, s));
            }
        }
        if (ts.isIdentifier(n) && imported(n) && valueReference(n)) {
            const s = ultimate(checker.getSymbolAtLocation(n));
            if (!constEnum(s)) return read(n, n.text, s);
        }
        return ts.visitEachChild(n, visit, context);
    }
    return sf => {
        if (!relative(sf.fileName).startsWith('src/')) return sf;
        const file = relative(sf.fileName);
        const statements = [statement('__cycleBegin', file)];
        for (const s of sf.statements) {
            const changed = ts.visitNode(s, visit);
            if (ts.isImportDeclaration(s) || ts.isExportDeclaration(s) || ts.isFunctionDeclaration(s)
                || ts.isInterfaceDeclaration(s) || ts.isTypeAliasDeclaration(s)) {
                statements.push(changed);
            } else {
                statements.push(statement('__cyclePush', location(s)), changed,
                    f.createExpressionStatement(call('__cyclePop', [])));
            }
        }
        statements.push(statement('__cycleEnd', file));
        return f.updateSourceFile(sf, statements);
    };
}
fs.mkdirSync(out, { recursive: true });
program.emit(undefined, undefined, undefined, undefined, { before: [instrument] });
fs.writeFileSync(path.join(out, 'package.json'), '{"type":"module"}\n');
fs.writeFileSync(path.join(out, 'metadata.json'), JSON.stringify(metadata, null, 2));
fs.writeFileSync(path.join(out, 'trace.mjs'), `
import fs from 'node:fs';
import { createRequire } from 'node:module';
import { fileURLToPath } from 'node:url';
// The upstream bundle supplies CommonJS host globals; keep its Node host active in ESM.
globalThis.require = createRequire(new URL('./src/compiler/sys.js', import.meta.url));
globalThis.__filename = fileURLToPath(new URL('./src/compiler/sys.js', import.meta.url));
globalThis.__dirname = fileURLToPath(new URL('./src/compiler/', import.meta.url));
const events = [], reads = [], stack = [], begun = new Set(), ended = new Set();
let active = true;
globalThis.__cycleBegin = file => {
  events.push({event:'begin', file}); begun.add(file);
  if(file === 'src/tsc/tsc.ts') active = false;
};
globalThis.__cycleEnd = file => {events.push({event:'end', file}); ended.add(file);};
globalThis.__cyclePush = loc => stack.push(loc);
globalThis.__cyclePop = () => stack.pop();
const metadata = JSON.parse(fs.readFileSync(new URL('./metadata.json', import.meta.url)));
globalThis.__cycleRead = id => {
  if (!active) return;
  const m = metadata[id];
  reads.push({...m, statement:stack.at(-1), sequence:events.length,
    providerBegun:begun.has(m.provider), providerEnded:ended.has(m.provider)});
};
process.on('exit', () => fs.writeFileSync(new URL('./trace.json', import.meta.url), JSON.stringify({events,reads}, null, 2)));
process.argv = [process.argv[0], new URL('./src/tsc/tsc.js', import.meta.url).pathname, '--version'];
await import('./src/tsc/tsc.js');
`);
const run = cp.spawnSync(process.execPath, [path.join(out, 'trace.mjs')], { encoding: 'utf8' });
fs.writeFileSync(path.join(out, 'node.json'), JSON.stringify({stdout:run.stdout,stderr:run.stderr,exit:run.status}, null, 2));
assert.equal(run.status, 0, run.stderr);
assert.equal(run.stdout, 'Version 6.0.3\n');
// Independently emit the original program and check that tracing kept its dependency graph.
program.emit(undefined, (file, text) => {
    const target = path.join(out, 'baseline', path.relative(out, file));
    fs.mkdirSync(path.dirname(target), { recursive: true });
    fs.writeFileSync(target, text);
});
const bootstrap = `import {createRequire} from 'node:module';
import {fileURLToPath} from 'node:url';
globalThis.require=createRequire(new URL('./src/compiler/sys.js',import.meta.url));
globalThis.__filename=fileURLToPath(new URL('./src/compiler/sys.js',import.meta.url));
globalThis.__dirname=fileURLToPath(new URL('./src/compiler/',import.meta.url));
process.argv=[process.argv[0],new URL('./src/tsc/tsc.js',import.meta.url).pathname,'--version'];
await import('./src/tsc/tsc.js');`;
fs.writeFileSync(path.join(out, 'baseline', 'boot.mjs'), bootstrap);
const baseline = cp.spawnSync(process.execPath, [path.join(out, 'baseline', 'boot.mjs')], { encoding: 'utf8' });
assert.equal(baseline.status, run.status, baseline.stderr);
assert.equal(baseline.stdout, run.stdout);
assert.equal(baseline.stderr, run.stderr);
function dependencies(file) {
    const sf = ts.createSourceFile(file, fs.readFileSync(file, 'utf8'), ts.ScriptTarget.Latest, true, ts.ScriptKind.JS);
    return sf.statements.filter(s => (ts.isImportDeclaration(s) || ts.isExportDeclaration(s)) && s.moduleSpecifier)
        .map(s => s.moduleSpecifier.text).filter(s => s.startsWith('.'));
}
const visited = new Set(), order = [], graph = {};
function dfs(file) {
    if (visited.has(file)) return;
    visited.add(file);
    const emitted = path.join(out, file.replace(/\.ts$/, '.js'));
    const deps = dependencies(emitted);
    assert.deepEqual(deps, dependencies(path.join(out, 'baseline', file.replace(/\.ts$/, '.js'))));
    graph[file] = deps.map(spec => path.posix.normalize(path.posix.join(path.posix.dirname(file), spec)).replace(/\.js$/, '.ts'));
    for (const target of graph[file]) dfs(target);
    order.push(file);
}
dfs('src/tsc/tsc.ts');
const trace = JSON.parse(fs.readFileSync(path.join(out, 'trace.json')));
const observed = trace.events.filter(e => e.event === 'begin').map(e => e.file);
assert.deepEqual(order, observed);
assert.throws(() => assert.deepEqual(order.slice().reverse(), observed));
fs.writeFileSync(path.join(out, 'order.json'), JSON.stringify({order, graph, baseline: {stdout:baseline.stdout,stderr:baseline.stderr,exit:baseline.status},
    reversedOrderMutant:'caught by ESM DFS versus observed body order'}, null, 2));
console.log(`trace: ${out}/trace.json; ${order.length} module bodies match independent ESM DFS; uninstrumented output matches; reversed-order mutant caught`);
