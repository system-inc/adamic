'use strict';
const fs = require('node:fs'), path = require('node:path'), assert = require('node:assert/strict');
const ts = require(path.resolve(process.argv[2])), observations = [];
let candidates = 0, throws = 0;
for (const fixture of ['excess.js', 'excess.ts']) for (const noImplicitAny of [false, true]) {
    const name = '/95f/' + fixture, text = fs.readFileSync(path.join(__dirname, fixture + '.txt'), 'utf8');
    const options = {strict: true, noImplicitAny, allowJs: true, checkJs: true, noEmit: true, target: ts.ScriptTarget.ESNext, module: ts.ModuleKind.CommonJS, moduleResolution: ts.ModuleResolutionKind.Node10, ignoreDeprecations: '6.0', experimentalDecorators: true};
    const host = ts.createCompilerHost(options), getSourceFile = host.getSourceFile;
    host.getSourceFile = (file, ...args) => file === name ? ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true) : getSourceFile.call(host, file, ...args);
    const program = ts.createProgram([name], options, host), checker = program.getTypeChecker(), source = program.getSourceFile(name);
    const diagnostics = ts.getPreEmitDiagnostics(program);
    const row = {fixture, noImplicitAny, diagnosticOutput: ts.formatDiagnostics(diagnostics, {getCanonicalFileName: f => f, getCurrentDirectory: () => '/', getNewLine: () => '\n'}), diagnostics: diagnostics.map(d => ({code: d.code, file: d.file?.fileName, start: d.start, length: d.length, message: ts.flattenDiagnosticMessageText(d.messageText, '\n')})), queries: []};
    function visit(node) {
        if (ts.isTypeNode(node)) {
            const position = node.parent.typeArguments ? node.parent.typeArguments.indexOf(node) : -1;
            const candidate = fixture === 'excess.js' && position > 0 && node.parent.getText().startsWith('C<');
            const query = {text: node.getText(), parent: node.parent.getText(), position, candidate};
            try { const result = checker.getTypeArgumentConstraint(node); query.result = result === undefined ? 'undefined' : checker.typeToString(result); }
            catch (error) { query.error = error.name + ': ' + error.message; query.stack = error.stack; throws++; }
            if (candidate) { candidates++; if (process.argv.includes('--expect-fixed')) assert.equal(query.result, 'undefined', JSON.stringify(query)); }
            row.queries.push(query);
        }
        ts.forEachChild(node, visit);
        if (node.jsDoc) for (const doc of node.jsDoc) visit(doc);
    }
    // Keep the complete result, including stack and compiler diagnostics, even on a mutant.
    try { visit(source); } catch (error) { observations.push(row); console.log(JSON.stringify({observations, candidates, throws}, null, 2)); throw error; }
    observations.push(row);
}
console.log(JSON.stringify({observations, candidates, throws}, null, 2));
if (process.argv.includes('--expect-fixed')) { assert.equal(candidates, 8); assert.equal(throws, 0); }
