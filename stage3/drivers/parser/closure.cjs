// File-level import/re-export closure, including type-only edges.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.PARSER_TYPESCRIPT);
const tree = path.resolve(process.argv[2]);
const directory = path.join(tree, 'src/compiler');
const roots = ts.sys.readDirectory(directory, ['.ts']).sort();
const options = { moduleResolution: ts.ModuleResolutionKind.Bundler, module: ts.ModuleKind.ESNext };
const edges = [];
for (const file of roots) {
    const source = ts.createSourceFile(file, fs.readFileSync(file, 'utf8'), ts.ScriptTarget.Latest, true);
    for (const node of source.statements) {
        if (!(ts.isImportDeclaration(node) || ts.isExportDeclaration(node)) || !node.moduleSpecifier) continue;
        const specifier = node.moduleSpecifier.text;
        const target = ts.resolveModuleName(specifier, file, options, ts.sys).resolvedModule?.resolvedFileName;
        edges.push({file: path.relative(tree, file), target: target && path.relative(tree, target), specifier, typeOnly: !!(node.isTypeOnly || node.importClause?.isTypeOnly)});
    }
}
function closure(start) {
    const seen = new Set([start]);
    const todo = [start];
    while (todo.length) {
        const file = todo.pop();
        for (const edge of edges) if (edge.file === file && edge.target?.startsWith('src/compiler/') && !seen.has(edge.target)) {
            seen.add(edge.target); todo.push(edge.target);
        }
    }
    return [...seen].sort();
}
const parser = closure('src/compiler/parser.ts');
const scanner = closure('src/compiler/scanner.ts');
fs.writeFileSync(process.argv[3], JSON.stringify({definition: 'transitive resolved import and re-export file graph, including type-only edges', parser, scanner, parserBeyondScanner: parser.filter(f => !scanner.includes(f)), edges}, null, 2) + '\n');
console.log(JSON.stringify({parser: parser.length, scanner: scanner.length, parserBeyondScanner: parser.filter(f => !scanner.includes(f))}));
