// Static evidence supplies the domain erased from the stock generic call.
const assert = require('node:assert/strict');
const path = require('node:path');
require('./probe.cjs');
const ts = require(process.argv[2]);
const root = process.argv[3];
const options = {
    strict: true, target: ts.ScriptTarget.ES2020, module: ts.ModuleKind.NodeNext,
    moduleResolution: ts.ModuleResolutionKind.NodeNext, types: [], skipLibCheck: true,
};
const host = ts.createCompilerHost(options);
if (process.argv[4] === '--mutant') {
    const read = host.readFile;
    host.readFile = file => {
        const text = read(file);
        if (file === path.join(root, 'src/compiler/checker.ts')) {
            const signature = 'function getTypeOfExpression(node: Expression)';
            assert.equal(text.split(signature).length, 2);
            return text.replace(signature, 'function getTypeOfExpression(node: Expression & { readonly flags: NodeFlags.Synthesized })');
        }
        return text;
    };
}
const program = ts.createProgram([path.join(root, 'src/compiler/checker.ts')], options, host);
const checker = program.getTypeChecker();
const calls = [];
for (const file of program.getSourceFiles()) {
    if (file.isDeclarationFile) continue;
    function walk(node) {
        if (ts.isCallExpression(node) && node.expression.getText(file) === 'setNodeFlags') {
            const holder = checker.getTypeAtLocation(node.arguments[0]);
            const flags = holder.getProperty('flags');
            const field = checker.getTypeOfSymbolAtLocation(flags, node.arguments[0]);
            const location = file.getLineAndCharacterOfPosition(node.getStart(file));
            calls.push({ file: path.relative(root, file.fileName), line: location.line + 1,
                holder: checker.typeToString(holder), flags: checker.typeToString(field),
                expression: node.getText(file) });
        }
        ts.forEachChild(node, walk);
    }
    walk(file);
}
assert.equal(calls.length, 1);
assert.equal(calls[0].flags, 'NodeFlags');
assert.equal(calls[0].holder, 'Expression');
console.log(JSON.stringify(calls, null, 2));
