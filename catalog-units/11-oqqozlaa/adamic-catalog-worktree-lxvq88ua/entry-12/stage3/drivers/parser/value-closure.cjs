// Supplementary declaration-level value reachability. This does not replace
// the compiler loader's file graph or prove that unreachable code can be cut.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.PARSER_TYPESCRIPT);
const tree = path.resolve(process.argv[2]);
const directory = path.join(tree, 'src/compiler');
const roots = ts.sys.readDirectory(directory, ['.ts']).sort();
const program = ts.createProgram(roots, { moduleResolution: ts.ModuleResolutionKind.Bundler, module: ts.ModuleKind.ESNext, target: ts.ScriptTarget.Latest });
const checker = program.getTypeChecker();
function owner(node) {
    while (node.parent && !ts.isSourceFile(node.parent)) node = node.parent;
    return ts.isSourceFile(node) ? undefined : node;
}
function reachable(file, exported) {
    const source = program.getSourceFile(path.join(directory, file));
    const symbol = checker.getSymbolAtLocation(source);
    const root = checker.getExportsOfModule(symbol).find(s => s.name === exported);
    if (!root) throw new Error(`missing ${file}:${exported}`);
    const seen = new Set();
    const pending = [];
    function add(symbol) {
        if (!symbol) return;
        if (symbol.flags & ts.SymbolFlags.Alias) symbol = checker.getAliasedSymbol(symbol);
        for (const declaration of symbol.declarations || []) {
            if (!declaration.getSourceFile().fileName.startsWith(directory + path.sep)) continue;
            const top = owner(declaration);
            if (top && !seen.has(top)) { seen.add(top); pending.push(top); }
        }
    }
    add(root);
    while (pending.length) {
        const top = pending.pop();
        function visit(node) {
            if (ts.isTypeNode(node) || ts.isInterfaceDeclaration(node) || ts.isTypeAliasDeclaration(node) ||
                ts.isImportDeclaration(node) || ts.isExportDeclaration(node)) return;
            if (ts.isIdentifier(node)) add(checker.getSymbolAtLocation(node));
            ts.forEachChild(node, visit);
        }
        visit(top);
    }
    return {
        files: [...new Set([...seen].map(n => path.relative(tree, n.getSourceFile().fileName)))].sort(),
        declarations: [...seen].map(n => ({file: path.relative(tree, n.getSourceFile().fileName), start: n.getStart(), end: n.end, kind: ts.SyntaxKind[n.kind]})).sort((a,b) => a.file.localeCompare(b.file) || a.start - b.start),
    };
}
const parser = reachable('parser.ts', 'createSourceFile');
const scanner = reachable('scanner.ts', 'createScanner');
const result = { definition: 'runtime identifier references reachable from exported createSourceFile/createScanner; declarations grouped by top-level statement; type nodes excluded; namespaces retained whole', parser, scanner, parserBeyondScanner: parser.files.filter(f => !scanner.files.includes(f)) };
fs.writeFileSync(process.argv[3], JSON.stringify(result, null, 2) + '\n');
console.log(JSON.stringify({parser: parser.files.length, scanner: scanner.files.length, parserBeyondScanner: result.parserBeyondScanner}, null, 2));
