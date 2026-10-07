#!/usr/bin/env node
// Gather declarations through stock TypeScript's checker. Never print declaration ASTs.
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const ts = require(process.env.SCANNER_TYPESCRIPT || process.env.SLICE_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3') throw new Error(`stock TypeScript 6.0.3 required, got ${ts.version}`);
const [treeArgument, outputArgument, ...entries] = process.argv.slice(2);
if (!treeArgument || !outputArgument || !entries.length) throw new Error('usage: slice.cjs <adapted-tree> <new-output> <relative-file>:<exported-symbol> [...]');
const tree = path.resolve(treeArgument), output = path.resolve(outputArgument);
if (fs.existsSync(output)) throw new Error('refusing existing output: ' + output);
const entrySpecs = entries.map(entry => {
    const colon = entry.lastIndexOf(':');
    if (colon < 1) throw new Error('entry must be file:symbol: ' + entry);
    return {file: path.resolve(tree, entry.slice(0, colon)), name: entry.slice(colon + 1)};
});
const program = ts.createProgram(entrySpecs.map(e => e.file), {
    target: ts.ScriptTarget.ESNext, module: ts.ModuleKind.ESNext,
    moduleResolution: ts.ModuleResolutionKind.Bundler, strict: true,
    verbatimModuleSyntax: true, skipLibCheck: true, allowImportingTsExtensions: true, noEmit: true,
});
const checker = program.getTypeChecker();
const reached = new Set(), queue = [], references = new Map(), namespaceFiles = new Set();
function ownFile(source) {
    return source && !source.isDeclarationFile && path.relative(tree, source.fileName).startsWith('src/') && !path.relative(tree, source.fileName).startsWith('../');
}
function rootDeclaration(node) {
    while (node.parent && !ts.isSourceFile(node.parent)) node = node.parent;
    if (ts.isImportDeclaration(node) || ts.isExportDeclaration(node) || ts.isSourceFile(node)) return undefined;
    return node;
}
function actual(symbol) {
    const seen = new Set();
    while (symbol && symbol.flags & ts.SymbolFlags.Alias) {
        if (seen.has(symbol)) throw new Error('alias cycle: ' + symbol.name);
        seen.add(symbol); symbol = checker.getAliasedSymbol(symbol);
    }
    return symbol;
}
function add(symbol) {
    symbol = actual(symbol);
    const moduleSource = symbol?.declarations?.find(ts.isSourceFile);
    if (moduleSource && ownFile(moduleSource)) {
        if (namespaceFiles.has(moduleSource)) return;
        namespaceFiles.add(moduleSource);
        for (const statement of moduleSource.statements) {
            if (ts.isExportDeclaration(statement) && !reached.has(statement)) { reached.add(statement); queue.push(statement); }
        }
        for (const exported of checker.getExportsOfModule(symbol)) add(exported);
    }
    for (const declaration of symbol?.declarations || []) {
        const source = declaration.getSourceFile();
        if (!ownFile(source)) continue;
        const root = rootDeclaration(declaration);
        if (!root) continue;
        if (!reached.has(root)) { reached.add(root); queue.push(root); }
    }
}
for (const entry of entrySpecs) {
    const source = program.getSourceFile(entry.file);
    if (!source) throw new Error('missing entry file: ' + entry.file);
    const symbol = checker.getExportsOfModule(checker.getSymbolAtLocation(source)).find(s => s.name === entry.name);
    if (!symbol) throw new Error('missing entry symbol: ' + entry.name);
    add(symbol);
}
function importBinding(symbol) {
    return symbol?.declarations?.find(d => ts.isImportSpecifier(d) || ts.isNamespaceImport(d) || ts.isImportClause(d));
}
function typePosition(node) {
    for (let current = node.parent; current && !ts.isStatement(current); current = current.parent) {
        if (ts.isTypeNode(current)) return true;
        if (ts.isHeritageClause(current) && (current.token === ts.SyntaxKind.ImplementsKeyword || ts.isInterfaceDeclaration(current.parent))) return true;
    }
    return false;
}
for (let index = 0; index < queue.length; index++) {
    const root = queue[index], source = root.getSourceFile();
    const imports = references.get(source) || new Map(); references.set(source, imports);
    function visit(node) {
        if (ts.isIdentifier(node)) {
            let symbol = ts.isExportSpecifier(node.parent) ? checker.getExportSpecifierLocalTargetSymbol(node.parent) : checker.getSymbolAtLocation(node);
            if (ts.isShorthandPropertyAssignment(node.parent) && node.parent.name === node) {
                symbol = checker.getShorthandAssignmentValueSymbol(node.parent) || symbol;
            }
            // A type-only import can share its spelling with a private value.
            // Ask for value meaning at reference sites, rather than following the type alias.
            if (!typePosition(node) && symbol && !(actual(symbol)?.flags & ts.SymbolFlags.Value)
                && !(ts.isPropertyAccessExpression(node.parent) && node.parent.name === node)) {
                symbol = checker.resolveName(node.text, node, ts.SymbolFlags.Value, false) || symbol;
            }
            const binding = importBinding(symbol);
            if (binding) {
                const resolved = actual(symbol);
                const target = resolved?.declarations?.find(d => ownFile(d.getSourceFile()));
                if (!target) throw new Error('unresolved source import: ' + node.text + ' in ' + source.fileName);
                const previous = imports.get(binding);
                imports.set(binding, {binding, symbol: resolved, target: target.getSourceFile(),
                    value: (previous?.value || false) || (!!(resolved.flags & ts.SymbolFlags.Value) && !typePosition(node))});
            }
            add(symbol);
        }
        ts.forEachChild(node, visit);
    }
    visit(root);
}
const byFile = new Map();
for (const node of reached) {
    const source = node.getSourceFile();
    if (!byFile.has(source)) byFile.set(source, []);
    byFile.get(source).push(node);
}
// A module used as a namespace value has its own source-file declaration.
// Keep its original export facade; every exported declaration is reached above.
for (const source of namespaceFiles) {
    if (!byFile.has(source)) byFile.set(source, []);
    for (const statement of source.statements) {
        if (ts.isExportDeclaration(statement) && !byFile.get(source).includes(statement)) byFile.get(source).push(statement);
    }
}
fs.mkdirSync(output, {recursive: true});
const ledger = [];
let lines = 0, bytes = 0, importLines = 0;
for (const [source, nodes] of [...byFile].sort((a,b) => a[0].fileName.localeCompare(b[0].fileName))) {
    const relative = path.relative(tree, source.fileName), destination = path.join(output, relative);
    const newline = source.text.includes('\r\n') ? '\r\n' : '\n';
    const groups = new Map();
    for (const ref of references.get(source)?.values() || []) {
        let specifier = path.relative(path.dirname(destination), path.join(output, path.relative(tree, ref.target.fileName))).replaceAll(path.sep, '/');
        if (!specifier.startsWith('.')) specifier = './' + specifier;
        // Direct .ts imports are accepted by both Adamic and the Node oracle loader.
        const binding = ref.binding;
        const exported = ref.symbol.name;
        const local = binding.name.text;
        const key = specifier + ':' + (ref.value ? 'value' : 'type');
        if (!groups.has(key)) groups.set(key, {specifier, value: ref.value, names: []});
        if (ts.isNamespaceImport(binding)) groups.get(key).namespace = local;
        else groups.get(key).names.push(exported === local ? exported : `${exported} as ${local}`);
    }
    let imports = '';
    for (const group of groups.values()) {
        imports += group.namespace ? `import ${group.value ? '' : 'type '}* as ${group.namespace} from ${JSON.stringify(group.specifier)};` + newline : `import ${group.value ? '' : 'type '}{ ${group.names.join(', ')} } from ${JSON.stringify(group.specifier)};` + newline;
        importLines++;
    }
    let result = imports;
    for (const node of nodes.sort((a,b) => a.pos-b.pos)) {
        // Full-start includes the original comments and whitespace attached to the statement.
        const text = source.text.slice(node.getFullStart(), node.end);
        const start = source.getLineAndCharacterOfPosition(node.getStart(source));
        const record = {file: relative, kind: ts.SyntaxKind[node.kind], start: node.getFullStart(), end: node.end,
            line: start.line+1, names: node.name ? [node.name.text] : ts.isVariableStatement(node) ? node.declarationList.declarations.map(d => d.name.getText(source)) : [],
            bytes: Buffer.byteLength(text), lines: text.split('\n').length,
            sha256: crypto.createHash('sha256').update(text).digest('hex'), output_start: result.length};
        result += text; record.output_end = result.length; ledger.push(record);
        lines += record.lines; bytes += record.bytes;
    }
    fs.mkdirSync(path.dirname(destination), {recursive: true}); fs.writeFileSync(destination, result);
}
const summary = {typescript: ts.version, tree, entries, declarations: ledger.length,
    value_declarations: ledger.filter(r => !['InterfaceDeclaration','TypeAliasDeclaration'].includes(r.kind)).length,
    type_declarations: ledger.filter(r => ['InterfaceDeclaration','TypeAliasDeclaration'].includes(r.kind)).length,
    files: byFile.size, declaration_lines: lines, declaration_bytes: bytes, rewritten_import_lines: importLines};
fs.writeFileSync(path.join(output, 'slice.json'), JSON.stringify({summary, declarations: ledger}, null, 2)+'\n');
console.log(JSON.stringify(summary, null, 2));
