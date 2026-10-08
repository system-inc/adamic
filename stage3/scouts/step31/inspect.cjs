'use strict';
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const ts = require(process.env.STEP31_TYPESCRIPT);
const tree = path.resolve(process.argv[2]);
const entries = ['binder', 'checker', 'emitter'].map(name => path.join(tree, `src/compiler/${name}.ts`));
const program = ts.createProgram(entries, {target: ts.ScriptTarget.ESNext, module: ts.ModuleKind.ESNext,
    moduleResolution: ts.ModuleResolutionKind.Bundler, skipLibCheck: true, allowImportingTsExtensions: true});
const checker = program.getTypeChecker();
const result = {};
for (const file of entries) {
    const source = program.getSourceFile(file);
    const imports = [];
    const functions = [];
    for (const statement of source.statements) {
        if (!ts.isImportDeclaration(statement)) continue;
        const names = statement.importClause?.namedBindings;
        const bindings = [];
        if (names && ts.isNamedImports(names)) for (const element of names.elements) {
            const symbol = checker.getSymbolAtLocation(element.name);
            const actual = symbol && symbol.flags & ts.SymbolFlags.Alias ? checker.getAliasedSymbol(symbol) : symbol;
            bindings.push({name: element.name.text, typeOnly: !!(statement.importClause.isTypeOnly || element.isTypeOnly),
                declaringFiles: [...new Set((actual?.declarations || []).map(node => path.relative(tree, node.getSourceFile().fileName)))].sort()});
        }
        imports.push({namespace: names && ts.isNamespaceImport(names) ? names.name.text : null, module: statement.moduleSpecifier.text, line: source.getLineAndCharacterOfPosition(statement.getStart()).line + 1, bindings});
    }
    function visit(node) {
        if (ts.isFunctionDeclaration(node) && node.name) {
            const start = source.getLineAndCharacterOfPosition(node.getStart());
            const end = source.getLineAndCharacterOfPosition(node.end);
            functions.push({name: node.name.text, line: start.line + 1, endLine: end.line + 1});
        }
        ts.forEachChild(node, visit);
    }
    visit(source);
    result[path.relative(tree, file)] = {sha256: crypto.createHash('sha256').update(fs.readFileSync(file)).digest('hex'), imports, functions};
}
process.stdout.write(JSON.stringify(result, null, 2) + '\n');
