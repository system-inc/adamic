// Resolve source imports/re-exports and generate names from stock TypeScript 6.0.3.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.SCANNER_TYPESCRIPT);
const [tree, output] = process.argv.slice(2).map(p => path.resolve(p));
const files = new Set();
const edges = [];
const external = new Set();
function visit(file) {
    if (files.has(file)) return;
    files.add(file);
    const source = ts.createSourceFile(file, fs.readFileSync(file, 'utf8'), ts.ScriptTarget.Latest, true);
    for (const statement of source.statements) {
        if (!(ts.isImportDeclaration(statement) || ts.isExportDeclaration(statement)) || !statement.moduleSpecifier) continue;
        const specifier = statement.moduleSpecifier.text;
        const result = ts.resolveModuleName(specifier, file, {moduleResolution: ts.ModuleResolutionKind.Bundler}, ts.sys).resolvedModule;
        const position = source.getLineAndCharacterOfPosition(statement.getStart(source));
        const edge = {file: path.relative(tree, file), line: position.line + 1, column: position.character + 1,
            specifier, type_only: ts.isImportDeclaration(statement) ? !!statement.importClause?.isTypeOnly : !!statement.isTypeOnly,
            target: result ? path.relative(tree, result.resolvedFileName) : null};
        edges.push(edge);
        if (result && !result.isExternalLibraryImport) visit(path.resolve(result.resolvedFileName));
        else external.add(specifier);
    }
}
visit(path.join(tree, 'src/compiler/scanner.ts'));
fs.writeFileSync(path.join(output, 'closure.json'), JSON.stringify({files: [...files].map(p => path.relative(tree, p)).sort(),
    external: [...external].sort(), edges}, null, 2) + '\n');
const names = Array.from({length: ts.SyntaxKind.Count}, (_, kind) => ts.SyntaxKind[kind]);
fs.writeFileSync(path.join(output, 'token-names.a'), '// Stock TypeScript 6.0.3 SyntaxKind names, including reverse-map aliases.\nexport const tokenNames: readonly string[] = ' + JSON.stringify(names) + ';\n');
console.log(JSON.stringify({closure_files: files.size, external: [...external].sort(), token_names: names.length}));
