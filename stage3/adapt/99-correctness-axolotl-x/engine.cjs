'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const crypto = require('node:crypto');
const ts = require('typescript');
const normalize = text => text.replace(/\r\n/g, '\n');
const namedFunction = node => (ts.isFunctionDeclaration(node) || ts.isFunctionExpression(node)) && node.name && ts.isIdentifier(node.name);
const digest = text => crypto.createHash('sha256').update(text).digest('hex');
function sources(tree) {
    const directory = path.join(tree, 'src/compiler');
    return ts.sys.readDirectory(directory, ['.ts'], undefined, ['**/*.ts']).sort();
}
function program(tree, replacements = new Map()) {
    assert.equal(ts.version, '6.0.3');
    const options = {target: ts.ScriptTarget.ESNext, module: ts.ModuleKind.NodeNext, moduleResolution: ts.ModuleResolutionKind.NodeNext, skipLibCheck: true, noEmit: true};
    const host = ts.createCompilerHost(options);
    const readFile = host.readFile.bind(host);
    host.readFile = name => replacements.get(path.resolve(name)) ?? readFile(name);
    return ts.createProgram(sources(tree), options, host);
}
function ownerMap(source) {
    const result = new Map();
    function visit(node, parents) {
        const named = namedFunction(node);
        const scoped = named || ((ts.isModuleDeclaration(node) || ts.isClassDeclaration(node)) && node.name);
        const chain = scoped ? [...parents, node.name.text] : parents;
        if (named && node.body) {
            const key = chain.join('.');
            assert(!result.has(key), `${source.fileName}: duplicate owner ${key}`);
            result.set(key, {node, start: node.getStart(source), end: node.end, text: node.getText(source)});
        }
        ts.forEachChild(node, child => visit(child, chain));
    }
    visit(source, []);
    return result;
}
function parse(file, text) {
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, `${file}: parse error`);
    return source;
}
function scan(tree, p = program(tree)) {
    const checker = p.getTypeChecker();
    const core = p.getSourceFile(path.join(tree, 'src/compiler/core.ts'));
    assert(core, 'core.ts missing');
    const declarations = core.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === 'maybeBind');
    assert.equal(declarations.length, 1, 'core maybeBind declaration drift');
    const target = checker.getSymbolAtLocation(declarations[0].name);
    function symbol(node) {
        let s = checker.getSymbolAtLocation(node);
        const seen = new Set();
        while (s && (s.flags & ts.SymbolFlags.Alias)) {
            assert(!seen.has(s), 'cyclic symbol alias'); seen.add(s);
            s = checker.getAliasedSymbol(s);
        }
        return s;
    }
    const sites = [], refused = [], foreign = [];
    for (const name of sources(tree)) {
        const source = p.getSourceFile(name);
        assert(source, `${name}: source missing from program`);
        assert.equal(source.parseDiagnostics.length, 0, `${name}: parse error`);
        const file = path.relative(path.join(tree, 'src/compiler'), name);
        function visit(node, owner, ancestry) {
            const scoped = namedFunction(node) || ((ts.isModuleDeclaration(node) || ts.isClassDeclaration(node)) && node.name);
            const chain = scoped ? [...ancestry, node.name.text] : ancestry;
            if (namedFunction(node) && node.body) owner = chain.join(".");
            if (ts.isCallExpression(node) && ts.isIdentifier(node.expression) && node.expression.text === 'maybeBind') {
                const where = source.getLineAndCharacterOfPosition(node.getStart(source));
                const row = {file, owner, line: where.line + 1, column: where.character + 1, before: node.getText(source), start: node.getStart(source), end: node.end};
                if (symbol(node.expression) !== target) foreign.push({...row, reason: 'callee does not resolve to core.ts export'});
                else {
                    const [x, member] = node.arguments;
                    let reason;
                    if (node.arguments.length !== 2) reason = 'not exactly two arguments';
                    else if (!ts.isIdentifier(x)) reason = 'receiver is not one identifier';
                    else if (!ts.isPropertyAccessExpression(member) || member.questionDotToken) reason = 'method is not a non-optional property access';
                    else if (!ts.isIdentifier(member.expression)) reason = 'method receiver is not one identifier';
                    else if (x.text !== member.expression.text || !symbol(x) || symbol(x) !== symbol(member.expression)) reason = 'receiver identifiers do not name the same symbol';
                    else if (!owner) reason = 'no named function owner';
                    if (reason) refused.push({...row, reason});
                    else sites.push({...row, after: `${member.getText(source)}?.bind(${x.getText(source)})`});
                }
            }
            ts.forEachChild(node, child => visit(child, owner, chain));
        }
        visit(source, undefined, []);
    }
    return {sites, refused, foreign};
}
function edit(text, edits) {
    for (const e of [...edits].sort((a,b) => b.start - a.start)) text = text.slice(0, e.start) + e.after + text.slice(e.end);
    return text;
}
function removableImport(file, text) {
    const source = parse(file, text);
    const specs = [];
    function find(node) { if (ts.isImportSpecifier(node) && node.name.text === 'maybeBind') specs.push(node); ts.forEachChild(node, find); }
    find(source);
    assert.equal(specs.length, 1, `${file}: missing or duplicate maybeBind import`);
    const spec = specs[0];
    let uses = 0;
    function count(node) {
        if (node === spec) return;
        if (ts.isIdentifier(node) && node.text === 'maybeBind') uses++;
        ts.forEachChild(node, count);
    }
    count(source);
    if (uses) return undefined;
    const start = text.lastIndexOf('\n', spec.getStart(source)) + 1;
    const end = text.indexOf('\n', spec.end) + 1;
    assert.equal(text.slice(start,end), '    maybeBind,\n', `${file}: changed import spelling`);
    return {start, end, before: text.slice(start,end), after: ''};
}
module.exports = {fs, path, assert, ts, normalize, digest, sources, program, ownerMap, parse, scan, edit, removableImport};
