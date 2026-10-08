// Independently compare the AST edge walk with TypeScript's program loader.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.SCANNER_TYPESCRIPT);
const [tree, graphFile] = process.argv.slice(2).map(p => path.resolve(p));
const program = ts.createProgram([path.join(tree, 'src/tsc/tsc.ts')], {
    noLib: true, types: [], moduleResolution: ts.ModuleResolutionKind.Bundler,
    module: ts.ModuleKind.ESNext, target: ts.ScriptTarget.ESNext,
});
const expected = program.getSourceFiles().map(file => path.relative(tree, file.fileName)).sort();
const actual = JSON.parse(fs.readFileSync(graphFile, 'utf8')).files;
if (JSON.stringify(actual) !== JSON.stringify(expected)) {
    console.error(JSON.stringify({expected, actual}));
    process.exit(1);
}
console.log(`graph matches stock TypeScript program: ${expected.length} files`);
