'use strict';
// Read-only tooling: source expressions and origin evidence, never source edits.
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
if (ts.version !== '6.0.3') throw new Error('expected TypeScript 6.0.3');
const [tree, input, output] = process.argv.slice(2);
const rows = JSON.parse(fs.readFileSync(input, 'utf8'));
const roots = ts.sys.readDirectory(path.join(tree, 'src/compiler'), ['.ts']);
const program = ts.createProgram(roots, {
    strict: true, exactOptionalPropertyTypes: true, noUncheckedIndexedAccess: true,
    noImplicitReturns: true, noFallthroughCasesInSwitch: true, erasableSyntaxOnly: true,
    verbatimModuleSyntax: true, allowImportingTsExtensions: true, noEmit: true,
    module: ts.ModuleKind.ESNext, moduleDetection: ts.ModuleDetectionKind.Force,
    moduleResolution: ts.ModuleResolutionKind.Bundler, target: ts.ScriptTarget.ES2024,
    lib: ['lib.es2024.d.ts'], types: [],
});
const checker = program.getTypeChecker();
function at(node, position) {
    let result = node;
    ts.forEachChild(node, child => {
        if (child.getStart() <= position && position < child.end) result = at(child, position);
    });
    return result;
}
function loc(node) {
    const file = node.getSourceFile();
    const point = file.getLineAndCharacterOfPosition(node.getStart());
    return {file: path.relative(tree, file.fileName), line: point.line + 1,
        column: point.character + 1, text: node.getText().slice(0, 1200)};
}
for (const row of rows) {
    const file = program.getSourceFile(path.join(tree, row.file));
    if (!file) throw new Error('missing source ' + row.file);
    const position = file.getPositionOfLineAndCharacter(row.line - 1, row.column - 1);
    const node = at(file, position);
    row.token = node.getText();
    row.ancestors = [];
    for (let n = node; n && row.ancestors.length < 5; n = n.parent) row.ancestors.push({kind: ts.SyntaxKind[n.kind], ...loc(n)});
    let arg;
    for (let n = node; n; n = n.parent) {
        if ((ts.isCallExpression(n) || ts.isNewExpression(n)) && n.arguments) {
            arg = n.arguments.find(a => a.getStart() <= position && position < a.end);
            if (arg) break;
        }
    }
    if (arg) {
        row.argument = {kind: ts.SyntaxKind[arg.kind], ...loc(arg)};
        row.stock_argument_type = checker.typeToString(checker.getTypeAtLocation(arg));
        let symbol = ts.isIdentifier(arg) && checker.getSymbolAtLocation(arg);
        row.argument_declarations = symbol?.declarations?.map(loc) || [];
    }
    row.context = file.text.split(/\r?\n/).slice(Math.max(0, row.line - 4), row.line + 3).join('\n');
}
fs.writeFileSync(output, JSON.stringify(rows, null, 2) + '\n');
