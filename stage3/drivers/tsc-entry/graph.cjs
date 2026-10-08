// Resolve the entry source closure with stock TypeScript 6.0.3.
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
    function edge(node, specifier, typeOnly, kind) {
        const result = ts.resolveModuleName(specifier, file, {moduleResolution: ts.ModuleResolutionKind.Bundler}, ts.sys).resolvedModule;
        const position = source.getLineAndCharacterOfPosition(node.getStart(source));
        edges.push({file: path.relative(tree, file), line: position.line + 1, column: position.character + 1,
            specifier, type_only: typeOnly, kind,
            target: result ? path.relative(tree, result.resolvedFileName) : null});
        if (result && !result.isExternalLibraryImport) visit(path.resolve(result.resolvedFileName));
        else external.add(specifier);
    }
    function walk(node) {
        if ((ts.isImportDeclaration(node) || ts.isExportDeclaration(node)) && node.moduleSpecifier) {
            edge(node, node.moduleSpecifier.text,
                ts.isImportDeclaration(node) ? !!node.importClause?.isTypeOnly : !!node.isTypeOnly, 'module');
        } else if (ts.isImportTypeNode(node) && ts.isLiteralTypeNode(node.argument) && ts.isStringLiteral(node.argument.literal)) {
            edge(node, node.argument.literal.text, true, 'import-type');
        } else if (ts.isCallExpression(node) && node.arguments.length && ts.isStringLiteral(node.arguments[0])
            && (node.expression.kind === ts.SyntaxKind.ImportKeyword || (ts.isIdentifier(node.expression) && node.expression.text === 'require'))) {
            edge(node, node.arguments[0].text, false, 'host-load');
        }
        node.forEachChild(walk);
    }
    walk(source);
}
visit(path.join(tree, 'src/tsc/tsc.ts'));
fs.writeFileSync(path.join(output, 'closure.json'), JSON.stringify({files: [...files].map(p => path.relative(tree, p)).sort(),
    external: [...external].sort(), edges}, null, 2) + '\n');
fs.writeFileSync(path.join(output, 'outside-compiler.json'), JSON.stringify([...files].map(p => path.relative(tree, p)).filter(p => !p.startsWith('src/compiler/')).sort(), null, 2) + '\n');
console.log(JSON.stringify({files: files.size, external: [...external].sort()}));
