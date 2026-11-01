'use strict';
const fs = require('node:fs');
const path = require('node:path');
const ts = require(path.join(process.argv[2], 'built/local/typescript.js'));
const results = [];
for (const fixture of ['excess.ts', 'excess.js', 'index.ts', 'commonjs.js', 'consumer.ts', 'common-consumer.ts', 'consumer-present.ts']) {
    const file = '/95d/' + fixture;
    const text = fs.readFileSync(path.join(__dirname, fixture + '.txt'), 'utf8');
    for (const noImplicitAny of [false, true]) {
        const options = {strict: true, noImplicitAny, allowJs: true, checkJs: true, noEmit: true, target: ts.ScriptTarget.ESNext, experimentalDecorators: true, module: ts.ModuleKind.CommonJS, moduleResolution: ts.ModuleResolutionKind.Node10, ignoreDeprecations: "6.0"};
        const host = ts.createCompilerHost(options);
        const sources = new Map([[file, text]]);
        for (const dependency of ['exported.js', 'exported-present.js', 'commonjs.js', 'defs.ts']) {
            if (!sources.has('/95d/' + dependency)) sources.set('/95d/' + dependency, fs.readFileSync(path.join(__dirname, dependency + '.txt'), 'utf8'));
        }
        const exists = host.fileExists, read = host.readFile, original = host.getSourceFile;
        host.fileExists = name => sources.has(name) || exists.call(host, name);
        host.readFile = name => sources.has(name) ? sources.get(name) : read.call(host, name);
        host.directoryExists = name => name === '/95d' || ts.sys.directoryExists(name);
        host.getSourceFile = (name, ...args) => sources.has(name) ? ts.createSourceFile(name, sources.get(name), ts.ScriptTarget.Latest, true) : original.call(host, name, ...args);
        const program = ts.createProgram([file], options, host), checker = program.getTypeChecker(), source = program.getSourceFile(file);
        global.__95dTest = fixture + ':noImplicitAny=' + noImplicitAny;
        const row = {fixture, noImplicitAny, diagnosticOutput: ts.formatDiagnostics(ts.getPreEmitDiagnostics(program), {getCanonicalFileName: f => f, getCurrentDirectory: () => '/', getNewLine: () => '\n'}), diagnostics: ts.getPreEmitDiagnostics(program).map(d => d.code), queries: []};
        function visit(node) {
            if (ts.isTypeNode(node)) {
                try { const constraint = checker.getTypeArgumentConstraint(node); row.queries.push({text: node.getText(), constraint: constraint && checker.typeToString(constraint)}); }
                catch (error) { row.queries.push({text: node.getText(), error: error.name + ': ' + error.message, stack: error.stack}); }
            }
            if (ts.isPropertyAccessExpression(node)) {
                try { const symbol = checker.getSymbolAtLocation(node); const type = checker.getTypeAtLocation(node.expression); row.queries.push({access: node.getText(), symbol: symbol && symbol.name, type: checker.typeToString(type), typeSymbol: type.symbol && type.symbol.name, alias: type.aliasSymbol && type.aliasSymbol.name, members: type.members && [...type.members.keys()], indexes: checker.getIndexInfosOfType(type).map(i => ({declaration: !!i.declaration, key: checker.typeToString(i.keyType)}))}); }
                catch (error) { row.queries.push({access: node.getText(), error: error.name + ': ' + error.message, stack: error.stack}); }
            }
            ts.forEachChild(node, visit);
            if (node.jsDoc) for (const doc of node.jsDoc) visit(doc);
        }
        visit(source);
        results.push(row);
    }
}
console.log(JSON.stringify(results, null, 2));
