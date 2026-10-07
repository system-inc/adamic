#!/usr/bin/env node
// Gather declarations through stock TypeScript's checker. Never print declaration ASTs.
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const ts = require(process.env.SCANNER_TYPESCRIPT || process.env.SLICE_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3') throw new Error(`stock TypeScript 6.0.3 required, got ${ts.version}`);
const [treeArgument, outputArgument, ...arguments] = process.argv.slice(2);
const entries = [], whySelectors = [];
let preserveEvaluation = true, conservativeNamespaces = false, applyAdaptations = true;
for (let index = 0; index < arguments.length; index++) {
    if (arguments[index] === '--why') {
        if (!arguments[index + 1]) throw new Error('--why needs file:declaration');
        whySelectors.push(arguments[++index]);
    } else if (arguments[index] === '--no-adapt') applyAdaptations = false;
    else if (arguments[index] === '--reference-only') preserveEvaluation = false;
    else if (arguments[index] === '--conservative-namespaces') conservativeNamespaces = true;
    else entries.push(arguments[index]);
}
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
const parents = new Map(), expandedNamespaceFiles = new Set();
let currentCause;
function reach(node, cause = currentCause) {
    if (!reached.has(node)) { reached.add(node); queue.push(node); parents.set(node, cause); }
}
function unwrap(node) {
    while (ts.isParenthesizedExpression(node) || ts.isAsExpression(node) || ts.isTypeAssertionExpression(node) || ts.isNonNullExpression(node) || ts.isSatisfiesExpression(node)) node = node.expression;
    return node;
}
function namespaceSymbol(node) {
    node = unwrap(node);
    if (ts.isPropertyAccessExpression(node)) {
        const container = namespaceSymbol(node.expression);
        if (container?.flags & (ts.SymbolFlags.Module | ts.SymbolFlags.Enum)) return actual(checker.getExportsOfModule(container).find(s => s.name === node.name.text));
    }
    return actual(checker.getSymbolAtLocation(node));
}
function qualifiedReference(node) {
    let expression = node;
    while (expression.parent && (ts.isParenthesizedExpression(expression.parent)
        || ts.isAsExpression(expression.parent) || ts.isTypeAssertionExpression(expression.parent)
        || ts.isNonNullExpression(expression.parent) || ts.isSatisfiesExpression(expression.parent))) expression = expression.parent;
    return (ts.isPropertyAccessExpression(expression.parent) && expression.parent.expression === expression)
        || (ts.isQualifiedName(expression.parent) && expression.parent.left === expression);
}
function ownFile(source) {
    return source && !source.isDeclarationFile && path.relative(tree, source.fileName).startsWith('src/') && !path.relative(tree, source.fileName).startsWith('../');
}
function rootDeclaration(node) {
    while (node.parent && !ts.isSourceFile(node.parent) && !ts.isModuleBlock(node.parent)) node = node.parent;
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
function add(symbol, qualified = false, cause = currentCause) {
    symbol = actual(symbol);
    const moduleSource = symbol?.declarations?.find(ts.isSourceFile);
    if (moduleSource && ownFile(moduleSource)) {
        namespaceFiles.add(moduleSource);
        if ((!qualified || conservativeNamespaces) && !expandedNamespaceFiles.has(moduleSource)) {
            expandedNamespaceFiles.add(moduleSource);
            for (const exported of checker.getExportsOfModule(symbol)) add(exported, false, {...cause, expansion: 'module namespace value expands ' + path.relative(tree, moduleSource.fileName)});
        }
    }
    for (const declaration of symbol?.declarations || []) {
        const source = declaration.getSourceFile();
        if (!ownFile(source)) continue;
        if (ts.isModuleDeclaration(declaration)) {
            if (!qualified) for (const member of checker.getExportsOfModule(symbol)) add(member, false, {...cause, expansion: 'namespace value expands ' + symbol.name});
            continue;
        }
        const root = rootDeclaration(declaration);
        if (!root) continue;
        reach(root, cause);
    }
}
for (const entry of entrySpecs) {
    const source = program.getSourceFile(entry.file);
    if (!source) throw new Error('missing entry file: ' + entry.file);
    const symbol = checker.getExportsOfModule(checker.getSymbolAtLocation(source)).find(s => s.name === entry.name);
    if (!symbol) throw new Error('missing entry symbol: ' + entry.name);
    add(symbol, false, {entry: entry.name});
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
        if (ts.isPropertyAccessExpression(node)) {
            const container = namespaceSymbol(node.expression);
            if (container?.flags & (ts.SymbolFlags.Module | ts.SymbolFlags.Enum)) {
                const member = checker.getExportsOfModule(container).find(s => s.name === node.name.text);
                const position = source.getLineAndCharacterOfPosition(node.getStart(source));
                add(member, false, {from: root, reference: {file: path.relative(tree, source.fileName), line: position.line + 1, column: position.character + 1, text: node.getText(source).slice(0,180), type: typePosition(node)}});
            }
        }
        if (ts.isIdentifier(node)) {
            const position = source.getLineAndCharacterOfPosition(node.getStart(source));
            currentCause = {from: root, reference: {file: path.relative(tree, source.fileName), line: position.line + 1, column: position.character + 1, text: node.parent.getText(source).slice(0, 180), type: typePosition(node)}};
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
            if (!typePosition(node) && importBinding(symbol) && !(actual(symbol)?.flags & ts.SymbolFlags.Value)) {
                for (const statement of source.statements) {
                    const names = ts.isVariableStatement(statement) ? statement.declarationList.declarations.map(d => d.name) : [statement.name];
                    if (names.some(name => name && ts.isIdentifier(name) && name.text === node.text)
                        && !ts.isInterfaceDeclaration(statement) && !ts.isTypeAliasDeclaration(statement)
                        && !reached.has(statement)) { reach(statement); }
                }
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
            const qualified = qualifiedReference(node);
            add(symbol, qualified);
        }
        ts.forEachChild(node, visit);
    }
    visit(root);
}
const byFile = new Map();
for (const node of reached) {
    const source = node.getSourceFile();
    if (!byFile.has(source)) byFile.set(source, []);
    let outer = node;
    while (outer.parent && !ts.isSourceFile(outer.parent)) outer = outer.parent;
    if (!byFile.get(source).includes(outer)) byFile.get(source).push(outer);
}
// A module used as a namespace value has its own source-file declaration.
// Keep its original export facade. Only actually referenced members expand.
for (const source of namespaceFiles) {
    if (!byFile.has(source)) byFile.set(source, []);
    for (const statement of source.statements) {
        if (ts.isExportDeclaration(statement) && !byFile.get(source).includes(statement)) byFile.get(source).push(statement);
    }
}
// Preserve the original ordered runtime dependency graph, including empty modules.
// Named bindings are rewritten separately; these side-effect edges retain ESM DFS order.
const evaluationEdges = new Map(), evaluationVisited = new Set();
function evaluate(source) {
    if (!source || !ownFile(source) || evaluationVisited.has(source)) return;
    evaluationVisited.add(source);
    if (!byFile.has(source)) byFile.set(source, []);
    const edges = [];
    for (const statement of source.statements) if (ts.isExportDeclaration(statement)) {
        if (!byFile.get(source).includes(statement)) byFile.get(source).push(statement);
        if (!statement.moduleSpecifier && statement.exportClause && ts.isNamedExports(statement.exportClause)) {
            for (const element of statement.exportClause.elements) {
                const symbol = checker.getExportSpecifierLocalTargetSymbol(element), binding = importBinding(symbol);
                const resolved = actual(symbol), target = resolved?.declarations?.find(ts.isSourceFile);
                if (binding && target) {
                    const imports = references.get(source) || new Map(); references.set(source, imports);
                    imports.set(binding, {binding, symbol: resolved, target, value: !statement.isTypeOnly && !element.isTypeOnly});
                }
            }
        }
    }
    for (const statement of source.statements) {
        if (!(ts.isImportDeclaration(statement) || ts.isExportDeclaration(statement)) || !statement.moduleSpecifier) continue;
        if (statement.isTypeOnly || statement.importClause?.isTypeOnly) continue;
        if (ts.isImportDeclaration(statement) && statement.importClause?.namedBindings && ts.isNamedImports(statement.importClause.namedBindings) && statement.importClause.namedBindings.elements.length && statement.importClause.namedBindings.elements.every(element => element.isTypeOnly)) continue;
        if (ts.isExportDeclaration(statement) && statement.exportClause && ts.isNamedExports(statement.exportClause) && statement.exportClause.elements.length && statement.exportClause.elements.every(element => element.isTypeOnly)) continue;
        const symbol = checker.getSymbolAtLocation(statement.moduleSpecifier);
        const target = symbol?.declarations?.find(ts.isSourceFile);
        if (!target || !ownFile(target)) continue;
        edges.push(target);
        evaluate(target);
    }
    evaluationEdges.set(source, edges);
}
const evaluationRoot = program.getSourceFile(path.join(tree, 'src/compiler/_namespaces/ts.ts'));
if (preserveEvaluation) {
    if (evaluationRoot) evaluate(evaluationRoot);
    for (const source of [...byFile.keys()]) evaluate(source);
}
function namesOf(node) {
    let names = node.name ? [node.name.getText(node.getSourceFile())] : ts.isVariableStatement(node) ? node.declarationList.declarations.map(d => d.name.getText(node.getSourceFile())) : [];
    let ancestor = node.parent;
    while (ancestor && !ts.isSourceFile(ancestor)) {
        if (ts.isModuleDeclaration(ancestor)) names = names.map(name => ancestor.name.text + '.' + name);
        ancestor = ancestor.parent;
    }
    return names;
}
function identity(node) {
    const source = node.getSourceFile(), position = source.getLineAndCharacterOfPosition(node.getStart(source));
    return {file: path.relative(tree, source.fileName), line: position.line + 1, names: namesOf(node)};
}
const explanations = [];
for (const selector of whySelectors) {
    const colon = selector.lastIndexOf(':'), file = selector.slice(0, colon), name = selector.slice(colon + 1);
    const candidates = [...reached].filter(node => path.relative(tree, node.getSourceFile().fileName) === file && namesOf(node).some(n => n === name || n.endsWith('.' + name)));
    if (!candidates.length) { explanations.push({selector, reached: false}); continue; }
    const chains = candidates.map(node => {
        const chain = [];
        while (node) { const cause = parents.get(node); chain.push({...identity(node), via: cause?.reference, expansion: cause?.expansion, entry: cause?.entry}); node = cause?.from; }
        return chain.reverse();
    });
    chains.sort((a,b) => a.length - b.length);
    explanations.push({selector, reached: true, chain: chains[0]});
}
fs.mkdirSync(output, {recursive: true});
const ledger = [];
let lines = 0, bytes = 0, importLines = 0;
for (const [source, nodes] of [...byFile].sort((a,b) => a[0].fileName.localeCompare(b[0].fileName))) {
    const relative = path.relative(tree, source.fileName), destination = path.join(output, relative);
    const newline = source.text.includes('\r\n') ? '\r\n' : '\n';
    const groups = new Map();
    for (const ref of references.get(source)?.values() || []) {
        let target = ref.target;
        const binding = ref.binding;
        if (preserveEvaluation && ref.value) {
            let declaration = binding;
            while (declaration && !ts.isImportDeclaration(declaration)) declaration = declaration.parent;
            const module = declaration && checker.getSymbolAtLocation(declaration.moduleSpecifier);
            target = module?.declarations?.find(ts.isSourceFile) || target;
        }
        let specifier = path.relative(path.dirname(destination), path.join(output, path.relative(tree, target.fileName))).replaceAll(path.sep, '/');
        if (!specifier.startsWith('.')) specifier = './' + specifier;
        // Direct .ts imports are accepted by both Adamic and the Node oracle loader.
        const exported = preserveEvaluation && ref.value ? binding.propertyName?.text || binding.name.text : ref.symbol.name;
        const local = binding.name.text;
        const key = specifier + ':' + (ref.value ? 'value' : 'type') + ':' + (ts.isNamespaceImport(binding) ? 'namespace:' + local : 'named');
        if (!groups.has(key)) groups.set(key, {specifier, value: ref.value, names: []});
        if (ts.isNamespaceImport(binding)) groups.get(key).namespace = local;
        else groups.get(key).names.push(exported === local ? exported : `${exported} as ${local}`);
    }
    let imports = '';
    for (const target of evaluationEdges.get(source) || []) {
        let specifier = path.relative(path.dirname(source.fileName), target.fileName).replaceAll(path.sep, '/');
        if (!specifier.startsWith('.')) specifier = './' + specifier;
        imports += `import ${JSON.stringify(specifier)};` + newline; importLines++;
    }
    for (const group of groups.values()) {
        imports += group.namespace ? `import ${group.value ? '' : 'type '}* as ${group.namespace} from ${JSON.stringify(group.specifier)};` + newline : `import ${group.value ? '' : 'type '}{ ${group.names.join(', ')} } from ${JSON.stringify(group.specifier)};` + newline;
        importLines++;
    }
    let result = imports;
    function append(node, from, to, kind = ts.SyntaxKind[node.kind]) {
        const text = source.text.slice(from, to);
        const start = source.getLineAndCharacterOfPosition(node.getStart(source));
        const record = {file: relative, kind, start: from, end: to,
            line: start.line+1, names: node.name ? [node.name.text] : ts.isVariableStatement(node) ? node.declarationList.declarations.map(d => d.name.getText(source)) : [],
            bytes: Buffer.byteLength(text), lines: text.split('\n').length,
            sha256: crypto.createHash('sha256').update(text).digest('hex'), output_start: result.length};
        result += text; record.output_end = result.length; ledger.push(record);
        lines += record.lines; bytes += record.bytes;
    }
    function containsReached(node) {
        if (reached.has(node)) return true;
        return ts.isModuleDeclaration(node) && node.body && (ts.isModuleBlock(node.body)
            ? node.body.statements.some(containsReached) : containsReached(node.body));
    }
    function emit(node) {
        if (ts.isModuleDeclaration(node) && node.body) {
            const body = node.body;
            if (!ts.isModuleBlock(body)) {
                append(node, node.getFullStart(), body.getFullStart(), 'NamespaceHeader');
                emit(body); return;
            }
            append(node, node.getFullStart(), body.getStart(source)+1, 'NamespaceHeader');
            for (const member of body.statements) if (containsReached(member)) emit(member);
            const tail = body.statements.length ? body.statements[body.statements.length-1].end : body.getStart(source)+1;
            append(node, tail, node.end, 'NamespaceClosing');
        } else append(node, node.getFullStart(), node.end);
    }
    for (const node of nodes.sort((a,b) => a.pos-b.pos)) emit(node);
    fs.mkdirSync(path.dirname(destination), {recursive: true}); fs.writeFileSync(destination, result);
}
const summary = {typescript: ts.version, tree, entries, declarations: ledger.filter(r => !r.kind.startsWith("Namespace")).length,
    copied_spans: ledger.length, namespace_wrapper_spans: ledger.filter(r => r.kind.startsWith("Namespace")).length,
    value_declarations: ledger.filter(r => !['InterfaceDeclaration','TypeAliasDeclaration'].includes(r.kind) && !r.kind.startsWith('Namespace')).length,
    type_declarations: ledger.filter(r => ['InterfaceDeclaration','TypeAliasDeclaration'].includes(r.kind)).length,
    files: byFile.size, declaration_files: new Set(ledger.filter(r => !r.kind.startsWith('Namespace') && r.kind !== 'ExportDeclaration').map(r => r.file)).size, code_declarations: ledger.filter(r => !r.kind.startsWith('Namespace') && r.kind !== 'ExportDeclaration').length, export_facade_records: ledger.filter(r => r.kind === 'ExportDeclaration').length, evaluation_only_files: [...byFile.values()].filter(nodes => nodes.every(node => ts.isExportDeclaration(node))).length, preserves_evaluation: preserveEvaluation, declaration_lines: lines, declaration_bytes: bytes, rewritten_import_lines: importLines};
fs.writeFileSync(path.join(output, 'slice.json'), JSON.stringify({summary, declarations: ledger, why: explanations, evaluation: [...evaluationEdges].map(([source, targets]) => ({file: path.relative(tree, source.fileName), imports: targets.map(t => path.relative(tree, t.fileName))}))}, null, 2)+'\n');
if (evaluationRoot && preserveEvaluation) fs.writeFileSync(path.join(output, 'slice-entry.a'), ['import \"./src/compiler/_namespaces/ts.ts\";', ...entrySpecs.map(entry => `import ${JSON.stringify('./' + path.relative(tree, entry.file).replaceAll(path.sep, '/'))};`)].join('\n')+'\n');
for (const explanation of explanations) console.error(JSON.stringify(explanation, null, 2));
console.log(JSON.stringify(summary, null, 2));

if (applyAdaptations) require("./apply-adaptations.cjs").apply(output);
